import { useEffect, useRef, useState } from 'react'
import { Html5Qrcode, Html5QrcodeScannerState } from 'html5-qrcode'

interface Props {
  onScan: (code: string) => void
  onClose: () => void
}

const READER_ID = 'barcode-reader'

/**
 * Modal quét barcode/QR bằng camera trình duyệt.
 * Chỉ có một nhiệm vụ: lấy một chuỗi mã từ camera và trả về qua onScan.
 * Không biết gì về SKU hay API.
 */
export default function BarcodeScannerModal({ onScan, onClose }: Props) {
  const scannerRef = useRef<Html5Qrcode | null>(null)
  const handledRef = useRef(false)
  // Giữ callback mới nhất để effect khởi tạo camera chỉ chạy một lần.
  const onScanRef = useRef(onScan)
  onScanRef.current = onScan

  const [error, setError] = useState('')

  useEffect(() => {
    const scanner = new Html5Qrcode(READER_ID)
    scannerRef.current = scanner

    const finish = (code: string) => {
      if (handledRef.current) return
      handledRef.current = true
      const value = code.trim()
      scanner
        .stop()
        .catch(() => {})
        .finally(() => onScanRef.current(value))
    }

    scanner
      .start(
        { facingMode: 'environment' },
        { fps: 10, qrbox: { width: 250, height: 250 } },
        (decodedText) => finish(decodedText),
        () => {
          /* Bỏ qua lỗi giải mã từng khung hình (frame không có mã) */
        }
      )
      .catch((err: { name?: string }) => {
        const name = err?.name || ''
        if (name === 'NotAllowedError') {
          setError('Bạn đã từ chối quyền camera. Vui lòng cấp quyền hoặc dùng máy quét cầm tay / gõ tay.')
        } else if (name === 'NotFoundError') {
          setError('Không tìm thấy camera trên thiết bị này.')
        } else {
          setError('Không thể khởi động camera. Vui lòng thử lại hoặc dùng máy quét cầm tay / gõ tay.')
        }
      })

    return () => {
      const s = scannerRef.current
      if (!s) return
      try {
        if (s.getState() === Html5QrcodeScannerState.SCANNING) {
          s.stop()
            .then(() => s.clear())
            .catch(() => {})
        } else {
          s.clear()
        }
      } catch {
        /* đã dừng / chưa khởi tạo */
      }
    }
  }, [])

  return (
    <div className="modal-overlay scanner-overlay" onClick={onClose}>
      <div className="modal scanner-modal" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Quét mã SKU</h2>
          <p className="modal-subtitle">Đưa barcode hoặc QR của sản phẩm vào khung hình</p>
        </div>

        {error ? (
          <div className="alert alert-error">{error}</div>
        ) : (
          <div id={READER_ID} className="scanner-viewport" />
        )}

        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={onClose}>
            Đóng
          </button>
        </div>
      </div>
    </div>
  )
}
