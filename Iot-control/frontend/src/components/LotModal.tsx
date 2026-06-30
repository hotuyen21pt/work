import { useEffect, useRef, useState } from 'react'
import {
  upsertLot,
  updateLot,
  listLotImages,
  uploadLotImages,
  deleteLotImage,
  countBoxes,
  saveDataset,
} from '../api/client'
import type { Lot, LotImage, DetBox } from '../types'
import BoxReviewModal from './BoxReviewModal'

interface Props {
  skuId: number
  skuCode: string
  skuName: string
  lot?: Lot
  userBranch: string
  onSave: () => void
  onClose: () => void
}

export default function LotModal({ skuId, skuCode, skuName, lot, userBranch, onSave, onClose }: Props) {
  const [form, setForm] = useState({
    lot_number: lot?.lot_number ?? '',
    manufacture_date: lot?.manufacture_date ?? '',
    expiry_date: lot?.expiry_date ?? '',
    qty: lot?.qty ?? 0,
    branch: lot?.branch ?? userBranch,
    notes: lot?.notes ?? '',
  })
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const isEdit = !!lot?.id

  // Ảnh đã lưu trên server (chế độ sửa).
  const [images, setImages] = useState<LotImage[]>([])
  // Ảnh chọn tạm khi tạo lô mới (chưa có lot.id) — sẽ upload sau khi tạo.
  // count: số box đếm được trên ảnh, để khi bỏ ảnh thì trừ lại khỏi Số lượng.
  const [pending, setPending] = useState<{ file: File; url: string; count: number }[]>([])
  const [uploading, setUploading] = useState(false)
  const [imgError, setImgError] = useState('')
  // Trạng thái đếm box tự động.
  const [counting, setCounting] = useState(false)
  const [countMsg, setCountMsg] = useState('')
  // Hàng đợi xem lại & chỉnh box cho từng ảnh vừa upload.
  const [reviewQueue, setReviewQueue] = useState<{ file: File; boxes: DetBox[] }[]>([])
  const [reviewIndex, setReviewIndex] = useState(0)
  // Kết quả đã xác nhận (ảnh + số box cuối cùng) tích luỹ qua hàng đợi.
  const [reviewResults, setReviewResults] = useState<{ file: File; count: number }[]>([])
  const fileInputRef = useRef<HTMLInputElement>(null)
  const cameraInputRef = useRef<HTMLInputElement>(null)

  // Giải phóng các object URL còn lại khi đóng modal.
  const pendingRef = useRef(pending)
  pendingRef.current = pending
  useEffect(() => () => pendingRef.current.forEach((p) => URL.revokeObjectURL(p.url)), [])

  useEffect(() => {
    if (!isEdit || !lot?.id) return
    listLotImages(lot.id)
      .then(setImages)
      .catch(() => setImgError('Không tải được danh sách ảnh'))
  }, [isEdit, lot?.id])

  const resetInputs = () => {
    if (fileInputRef.current) fileInputRef.current.value = ''
    if (cameraInputRef.current) cameraInputRef.current.value = ''
  }

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files ?? [])
    if (files.length === 0) {
      resetInputs()
      return
    }
    setImgError('')
    resetInputs()

    // Đếm/phát hiện box, rồi mở hàng đợi xem lại để người dùng chỉnh tay.
    setCounting(true)
    setCountMsg('Đang phát hiện box…')
    let res
    try {
      res = await countBoxes(files)
    } catch (err: any) {
      setCountMsg('')
      setImgError(err?.response?.data?.error || 'Đếm box thất bại')
      setCounting(false)
      return
    }
    setCounting(false)
    setCountMsg(`Đã phát hiện ${res.total} box — kiểm tra lại từng ảnh`)
    setReviewResults([])
    setReviewIndex(0)
    setReviewQueue(files.map((f, i) => ({ file: f, boxes: res.per_image?.[i]?.boxes ?? [] })))
  }

  // Sau khi review hết hàng đợi: cộng tổng vào Số lượng và đưa ảnh vào luồng
  // pending (lô mới) hoặc upload ngay (lô đang sửa), với số box đã chỉnh tay.
  const finalizeReview = async (results: { file: File; count: number }[]) => {
    if (results.length === 0) {
      setCountMsg('')
      return
    }
    const totalCount = results.reduce((s, r) => s + r.count, 0)
    setForm((prev) => ({ ...prev, qty: Number(prev.qty || 0) + totalCount }))
    setCountMsg(`Đã thêm: +${totalCount} box`)

    const files = results.map((r) => r.file)
    const counts = results.map((r) => r.count)
    if (!isEdit || !lot?.id) {
      setPending((prev) => [
        ...prev,
        ...files.map((f, i) => ({ file: f, url: URL.createObjectURL(f), count: counts[i] })),
      ])
    } else {
      setUploading(true)
      try {
        const created = await uploadLotImages(lot.id, files, counts)
        setImages((prev) => [...prev, ...created])
      } catch (err: any) {
        setImgError(err?.response?.data?.error || 'Tải ảnh thất bại')
      } finally {
        setUploading(false)
      }
    }
  }

  // Chuyển sang ảnh tiếp theo trong hàng đợi, hoặc kết thúc.
  const advanceReview = (results: { file: File; count: number }[]) => {
    const next = reviewIndex + 1
    if (next < reviewQueue.length) {
      setReviewResults(results)
      setReviewIndex(next)
    } else {
      void finalizeReview(results)
      setReviewQueue([])
      setReviewIndex(0)
      setReviewResults([])
    }
  }

  const handleReviewConfirm = (boxes: DetBox[], count: number) => {
    const item = reviewQueue[reviewIndex]
    if (item) {
      // Lưu dataset (ảnh gốc + nhãn) — lỗi chỉ cảnh báo, không chặn luồng.
      saveDataset(item.file, boxes).catch(() => setImgError('Lưu dataset thất bại (đã bỏ qua)'))
      advanceReview([...reviewResults, { file: item.file, count }])
    } else {
      advanceReview(reviewResults)
    }
  }

  const handleReviewCancel = () => {
    // Bỏ qua ảnh hiện tại: không tính số lượng, không lưu dataset.
    advanceReview(reviewResults)
  }

  const handleRemovePending = (index: number) => {
    const removed = pending[index]
    if (!removed) return
    URL.revokeObjectURL(removed.url)
    setPending((prev) => prev.filter((_, i) => i !== index))
    // Bỏ ảnh tạm thì trừ lại số box của ảnh đó khỏi Số lượng.
    setForm((f) => ({ ...f, qty: Math.max(0, Number(f.qty || 0) - (removed.count || 0)) }))
  }

  const handleDeleteImage = async (imageId: number) => {
    if (!lot?.id) return
    setImgError('')
    const target = images.find((img) => img.id === imageId)
    try {
      await deleteLotImage(lot.id, imageId)
      setImages((prev) => prev.filter((img) => img.id !== imageId))
      // Xóa ảnh thì trừ lại số box của ảnh đó khỏi Số lượng.
      if (target) {
        setForm((f) => ({ ...f, qty: Math.max(0, Number(f.qty || 0) - (target.count || 0)) }))
      }
    } catch (err: any) {
      setImgError(err?.response?.data?.error || 'Xóa ảnh thất bại')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      if (isEdit) {
        await updateLot(lot.id, {
          manufacture_date: form.manufacture_date,
          expiry_date: form.expiry_date,
          qty: Number(form.qty),
          notes: form.notes,
        })
      } else {
        const newLot = await upsertLot({
          sku_id: skuId,
          lot_number: form.lot_number,
          manufacture_date: form.manufacture_date,
          expiry_date: form.expiry_date,
          qty: Number(form.qty),
          branch: form.branch,
          notes: form.notes,
        })
        // Upload các ảnh đã chọn tạm sau khi lô được tạo.
        if (pending.length > 0 && newLot?.id) {
          try {
            await uploadLotImages(newLot.id, pending.map((p) => p.file), pending.map((p) => p.count))
          } catch {
            setImgError('Lô đã được lưu nhưng tải ảnh thất bại')
          }
          pending.forEach((p) => URL.revokeObjectURL(p.url))
          setPending([])
        }
      }
      onSave()
    } catch (err: any) {
      setError(err?.response?.data?.error || 'Có lỗi xảy ra, vui lòng thử lại')
    } finally {
      setLoading(false)
    }
  }

  const reviewing = reviewQueue.length > 0 && reviewIndex < reviewQueue.length

  return (
    <>
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()}>
        <button type="button" className="modal-close" onClick={onClose} aria-label="Đóng">✕</button>
        <div className="modal-header modal-header-icon">
          <span className="modal-icon">{isEdit ? '✏️' : '🏷️'}</span>
          <div>
            <h2>{isEdit ? 'Cập nhật lô' : 'Thêm / Cập nhật lô'}</h2>
            <p className="modal-subtitle">
              {skuCode} · {skuName}
            </p>
          </div>
        </div>

        {error && <div className="alert alert-error">{error}</div>}

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Số lô *</label>
            <input
              value={form.lot_number}
              onChange={(e) => setForm({ ...form, lot_number: e.target.value.toUpperCase() })}
              placeholder="LA001"
              required
              disabled={isEdit}
              autoFocus={!isEdit}
            />
            {!isEdit && (
              <small style={{ color: 'var(--gray-400)', fontSize: '0.75rem' }}>
                Nếu lô đã tồn tại, số lượng sẽ được cập nhật
              </small>
            )}
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Ngày sản xuất (NSX)</label>
              <input
                type="date"
                value={form.manufacture_date}
                onChange={(e) => setForm({ ...form, manufacture_date: e.target.value })}
              />
            </div>
            <div className="form-group">
              <label>Hạn sử dụng (HSD)</label>
              <input
                type="date"
                value={form.expiry_date}
                onChange={(e) => setForm({ ...form, expiry_date: e.target.value })}
              />
            </div>
          </div>

          <div className="form-group">
            <label>Ảnh bằng chứng</label>
            {imgError && <div className="alert alert-error">{imgError}</div>}

            <div
              style={{
                display: 'flex',
                flexWrap: 'wrap',
                gap: '0.5rem',
                marginBottom: '0.5rem',
              }}
            >
              {images.map((img) => (
                <div key={img.id} style={{ position: 'relative' }}>
                  <a href={img.url} target="_blank" rel="noreferrer">
                    <img
                      src={img.url}
                      alt="Ảnh lô"
                      style={{
                        width: 120,
                        height: 120,
                        objectFit: 'cover',
                        borderRadius: 8,
                        border: '1px solid var(--gray-200, #ddd)',
                      }}
                    />
                  </a>
                  <button
                    type="button"
                    onClick={() => handleDeleteImage(img.id)}
                    aria-label="Xóa ảnh"
                    style={{
                      position: 'absolute',
                      top: -6,
                      right: -6,
                      width: 20,
                      height: 20,
                      borderRadius: '50%',
                      border: 'none',
                      background: 'rgba(220,38,38,0.95)',
                      color: '#fff',
                      cursor: 'pointer',
                      fontSize: 12,
                      lineHeight: '20px',
                      padding: 0,
                    }}
                  >
                    ✕
                  </button>
                </div>
              ))}
              {pending.map((p, index) => (
                <div key={p.url} style={{ position: 'relative' }}>
                  <img
                    src={p.url}
                    alt="Ảnh chờ tải lên"
                    style={{
                      width: 120,
                      height: 120,
                      objectFit: 'cover',
                      borderRadius: 8,
                      border: '1px dashed var(--primary, #2563eb)',
                    }}
                  />
                  <button
                    type="button"
                    onClick={() => handleRemovePending(index)}
                    aria-label="Bỏ ảnh"
                    style={{
                      position: 'absolute',
                      top: -6,
                      right: -6,
                      width: 20,
                      height: 20,
                      borderRadius: '50%',
                      border: 'none',
                      background: 'rgba(220,38,38,0.95)',
                      color: '#fff',
                      cursor: 'pointer',
                      fontSize: 12,
                      lineHeight: '20px',
                      padding: 0,
                    }}
                  >
                    ✕
                  </button>
                </div>
              ))}
              {images.length === 0 && pending.length === 0 && (
                <small style={{ color: 'var(--gray-400)', fontSize: '0.75rem' }}>
                  Chưa có ảnh nào
                </small>
              )}
            </div>

            <div style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap' }}>
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => cameraInputRef.current?.click()}
                disabled={uploading || counting}
              >
                📷 Chụp hình
              </button>
              <button
                type="button"
                className="btn btn-ghost btn-sm"
                onClick={() => fileInputRef.current?.click()}
                disabled={uploading || counting}
              >
                🖼️ Tải ảnh
              </button>
            </div>

            <input
              ref={cameraInputRef}
              type="file"
              accept="image/*"
              capture="environment"
              onChange={handleUpload}
              disabled={uploading}
              style={{ display: 'none' }}
            />
            <input
              ref={fileInputRef}
              type="file"
              accept="image/jpeg,image/png,image/webp,image/gif"
              multiple
              onChange={handleUpload}
              disabled={uploading}
              style={{ display: 'none' }}
            />
            {uploading && (
              <small style={{ color: 'var(--gray-400)', fontSize: '0.75rem' }}>
                Đang tải ảnh lên...
              </small>
            )}
            {countMsg && (
              <small
                style={{
                  display: 'block',
                  marginTop: '0.25rem',
                  color: counting ? 'var(--gray-400)' : 'var(--primary, #2563eb)',
                  fontSize: '0.75rem',
                }}
              >
                {countMsg}
              </small>
            )}
          </div>

          <div className="form-row">
            <div className="form-group">
              <label>Số lượng *</label>
              <input
                type="number"
                inputMode="numeric"
                min="0"
                value={form.qty}
                onChange={(e) => setForm({ ...form, qty: Number(e.target.value) })}
                required
                autoFocus={isEdit}
              />
            </div>
            <div className="form-group">
              <label>Chi nhánh</label>
              <input
                value={form.branch}
                onChange={(e) => setForm({ ...form, branch: e.target.value })}
                placeholder="HCM01"
              />
            </div>
          </div>

          <div className="form-group">
            <label>Ghi chú</label>
            <textarea
              value={form.notes}
              onChange={(e) => setForm({ ...form, notes: e.target.value })}
              rows={2}
              placeholder="Ghi chú thêm về lô hàng..."
            />
          </div>

          <div className="form-actions">
            <button type="button" className="btn btn-ghost" onClick={onClose} disabled={loading}>
              Hủy
            </button>
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? 'Đang lưu...' : isEdit ? 'Cập nhật' : 'Lưu lô'}
            </button>
          </div>
        </form>
      </div>
    </div>

    {reviewing && (
      <BoxReviewModal
        key={reviewIndex}
        file={reviewQueue[reviewIndex].file}
        initialBoxes={reviewQueue[reviewIndex].boxes}
        index={reviewIndex}
        total={reviewQueue.length}
        onConfirm={handleReviewConfirm}
        onCancel={handleReviewCancel}
      />
    )}
    </>
  )
}
