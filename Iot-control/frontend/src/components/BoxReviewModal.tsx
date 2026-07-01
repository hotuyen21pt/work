import { useLayoutEffect, useRef, useState } from 'react'
import type { DetBox } from '../types'

interface Props {
  src: string // URL ảnh để hiển thị (object URL của file mới hoặc URL ảnh server)
  initialBoxes: DetBox[]
  index: number // vị trí ảnh trong hàng đợi (0-based)
  total: number // tổng số ảnh đang review
  onConfirm: (boxes: DetBox[], count: number) => void
  onCancel: () => void
}

// Chuẩn hoá box để x1<x2, y1<y2 (sau khi kéo có thể ngược chiều).
const normalize = (b: DetBox): DetBox => ({
  x1: Math.min(b.x1, b.x2),
  y1: Math.min(b.y1, b.y2),
  x2: Math.max(b.x1, b.x2),
  y2: Math.max(b.y1, b.y2),
  conf: b.conf,
})

// Box vẽ nhỏ hơn ngưỡng này (theo cạnh chuẩn hoá) bị bỏ qua — chống chạm nhầm.
const MIN_SIDE = 0.01

export default function BoxReviewModal({ src, initialBoxes, index, total, onConfirm, onCancel }: Props) {
  const [boxes, setBoxes] = useState<DetBox[]>(() => initialBoxes.map(normalize))
  const [selected, setSelected] = useState<number | null>(null)
  const [draft, setDraft] = useState<DetBox | null>(null)
  // Chế độ thao tác: 'draw' = kéo để vẽ box mới (bắt đầu ở bất kỳ đâu, kể cả bên
  // trong box khác); 'move' = bấm/kéo box để chọn & di chuyển.
  const [mode, setMode] = useState<'draw' | 'move'>('draw')
  // Kích thước ảnh hiển thị thực (px) để quy đổi toạ độ chuẩn hoá ↔ pixel.
  const [size, setSize] = useState({ w: 0, h: 0 })

  const imgRef = useRef<HTMLImageElement>(null)
  const drawingRef = useRef(false)
  const startRef = useRef({ x: 0, y: 0 })
  // Trạng thái kéo di chuyển một box đã có: vị trí con trỏ lúc bấm + box gốc.
  const dragRef = useRef<{ index: number; startX: number; startY: number; orig: DetBox } | null>(null)

  // Tính kích thước hiển thị: phóng ảnh vừa khít khung (92% rộng × 80% cao
  // màn hình), giữ đúng tỉ lệ. Cho phép phóng TO ảnh nhỏ để dễ vẽ box.
  const computeSize = () => {
    const el = imgRef.current
    if (!el || !el.naturalWidth || !el.naturalHeight) return
    const maxW = window.innerWidth * 0.92 - 64 // trừ padding modal
    const maxH = window.innerHeight * 0.8
    const scale = Math.min(maxW / el.naturalWidth, maxH / el.naturalHeight)
    setSize({
      w: Math.max(1, Math.round(el.naturalWidth * scale)),
      h: Math.max(1, Math.round(el.naturalHeight * scale)),
    })
  }

  // Tính lại khi đổi kích thước cửa sổ / xoay màn hình.
  useLayoutEffect(() => {
    const onResize = () => computeSize()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  // Lấy toạ độ chuẩn hoá 0..1 từ vị trí con trỏ trên ảnh.
  const toNorm = (clientX: number, clientY: number) => {
    const el = imgRef.current
    if (!el) return { x: 0, y: 0 }
    const r = el.getBoundingClientRect()
    const x = (clientX - r.left) / r.width
    const y = (clientY - r.top) / r.height
    return { x: Math.min(1, Math.max(0, x)), y: Math.min(1, Math.max(0, y)) }
  }

  const onPointerDownBg = (e: React.PointerEvent) => {
    // Bắt đầu vẽ box mới trên vùng trống.
    setSelected(null)
    ;(e.target as Element).setPointerCapture(e.pointerId)
    const p = toNorm(e.clientX, e.clientY)
    startRef.current = p
    drawingRef.current = true
    setDraft({ x1: p.x, y1: p.y, x2: p.x, y2: p.y })
  }

  // Bắt đầu kéo di chuyển một box đã có (bấm vào thân box).
  const onBoxPointerDown = (i: number, e: React.PointerEvent) => {
    e.stopPropagation()
    setSelected(i)
    ;(e.currentTarget as Element).setPointerCapture(e.pointerId)
    const p = toNorm(e.clientX, e.clientY)
    dragRef.current = { index: i, startX: p.x, startY: p.y, orig: boxes[i] }
  }

  const onPointerMove = (e: React.PointerEvent) => {
    if (drawingRef.current) {
      const p = toNorm(e.clientX, e.clientY)
      setDraft({ x1: startRef.current.x, y1: startRef.current.y, x2: p.x, y2: p.y })
      return
    }
    const drag = dragRef.current
    if (drag) {
      const p = toNorm(e.clientX, e.clientY)
      const w = drag.orig.x2 - drag.orig.x1
      const h = drag.orig.y2 - drag.orig.y1
      // Tịnh tiến theo delta con trỏ, kẹp trong ảnh và giữ nguyên kích thước box.
      const clamp = (v: number, max: number) => Math.min(max, Math.max(0, v))
      const nx1 = clamp(drag.orig.x1 + (p.x - drag.startX), 1 - w)
      const ny1 = clamp(drag.orig.y1 + (p.y - drag.startY), 1 - h)
      setBoxes((prev) =>
        prev.map((b, idx) =>
          idx === drag.index ? { x1: nx1, y1: ny1, x2: nx1 + w, y2: ny1 + h, conf: b.conf } : b,
        ),
      )
    }
  }

  const onPointerUp = () => {
    if (drawingRef.current) {
      drawingRef.current = false
      if (draft) {
        const b = normalize(draft)
        if (b.x2 - b.x1 >= MIN_SIDE && b.y2 - b.y1 >= MIN_SIDE) {
          setBoxes((prev) => [...prev, b])
        }
      }
      setDraft(null)
      return
    }
    dragRef.current = null
  }

  const removeSelected = () => {
    if (selected === null) return
    setBoxes((prev) => prev.filter((_, i) => i !== selected))
    setSelected(null)
  }

  const px = (b: DetBox) => ({
    x: b.x1 * size.w,
    y: b.y1 * size.h,
    w: (b.x2 - b.x1) * size.w,
    h: (b.y2 - b.y1) * size.h,
  })

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 'min(1100px, 95vw)', width: '95vw' }}>
        <button type="button" className="modal-close" onClick={onCancel} aria-label="Đóng">✕</button>
        <div className="modal-header modal-header-icon">
          <span className="modal-icon">🔲</span>
          <div>
            <h2>Kiểm tra & chỉnh box</h2>
            <p className="modal-subtitle">
              Ảnh {index + 1}/{total} ·{' '}
              {mode === 'draw' ? 'Kéo để vẽ box mới (được phép bắt đầu bên trong box khác)' : 'Bấm box để chọn · kéo để di chuyển · xoá bằng nút ✕'}
            </p>
            <div style={{ display: 'flex', gap: '0.4rem', marginTop: '0.35rem' }}>
              <button
                type="button"
                className={`btn btn-sm ${mode === 'draw' ? 'btn-primary' : 'btn-ghost'}`}
                onClick={() => {
                  setMode('draw')
                  setSelected(null)
                }}
              >
                ✏️ Vẽ box
              </button>
              <button
                type="button"
                className={`btn btn-sm ${mode === 'move' ? 'btn-primary' : 'btn-ghost'}`}
                onClick={() => setMode('move')}
              >
                ✋ Chọn / di chuyển
              </button>
            </div>
          </div>
        </div>

        <div
          style={{
            position: 'relative',
            display: 'block',
            width: size.w || 'auto',
            height: size.h || 'auto',
            margin: '0 auto',
            lineHeight: 0,
            userSelect: 'none',
          }}
        >
          <img
            ref={imgRef}
            src={src}
            alt="Ảnh kiểm đếm"
            onLoad={computeSize}
            width={size.w || undefined}
            height={size.h || undefined}
            style={{ display: 'block', borderRadius: 8 }}
            draggable={false}
          />
          <svg
            width={size.w}
            height={size.h}
            style={{ position: 'absolute', top: 0, left: 0, touchAction: 'none', cursor: 'crosshair' }}
            onPointerMove={onPointerMove}
            onPointerUp={onPointerUp}
          >
            {/* Lớp nền bắt thao tác vẽ box mới. */}
            <rect
              x={0}
              y={0}
              width={size.w}
              height={size.h}
              fill="transparent"
              onPointerDown={onPointerDownBg}
            />
            {boxes.map((b, i) => {
              const r = px(b)
              const isSel = i === selected
              return (
                <rect
                  key={i}
                  x={r.x}
                  y={r.y}
                  width={r.w}
                  height={r.h}
                  fill={isSel ? 'rgba(220,38,38,0.18)' : 'rgba(37,99,235,0.12)'}
                  stroke={isSel ? '#dc2626' : '#2563eb'}
                  strokeWidth={2}
                  // Ở chế độ vẽ: box không bắt sự kiện để con trỏ "xuyên qua" xuống
                  // lớp nền, cho phép bắt đầu vẽ box mới ngay bên trong box khác.
                  style={{ cursor: 'move', pointerEvents: mode === 'move' ? 'auto' : 'none' }}
                  onPointerDown={(e) => onBoxPointerDown(i, e)}
                />
              )
            })}
            {/* Box đang vẽ. */}
            {draft && (() => {
              const r = px(normalize(draft))
              return (
                <rect
                  x={r.x}
                  y={r.y}
                  width={r.w}
                  height={r.h}
                  fill="rgba(16,185,129,0.15)"
                  stroke="#10b981"
                  strokeWidth={2}
                  strokeDasharray="4 3"
                  pointerEvents="none"
                />
              )
            })()}
            {/* Nút xoá ở góc box đang chọn. */}
            {selected !== null && boxes[selected] && (() => {
              const r = px(boxes[selected])
              return (
                <g
                  transform={`translate(${r.x + r.w}, ${r.y})`}
                  style={{ cursor: 'pointer' }}
                  onPointerDown={(e) => {
                    e.stopPropagation()
                    removeSelected()
                  }}
                >
                  <circle cx={0} cy={0} r={11} fill="#dc2626" />
                  <text x={0} y={1} textAnchor="middle" dominantBaseline="middle" fill="#fff" fontSize={14}>✕</text>
                </g>
              )
            })()}
          </svg>
        </div>

        <div
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginTop: '0.75rem',
            flexWrap: 'wrap',
            gap: '0.5rem',
          }}
        >
          <strong style={{ fontSize: '1rem' }}>
            Số box: <span style={{ color: 'var(--primary, #2563eb)' }}>{boxes.length}</span>
          </strong>
          {selected !== null && (
            <button type="button" className="btn btn-ghost btn-sm" onClick={removeSelected}>
              🗑️ Xoá box đã chọn
            </button>
          )}
        </div>

        <div className="form-actions">
          <button type="button" className="btn btn-ghost" onClick={onCancel}>
            Bỏ qua ảnh này
          </button>
          <button type="button" className="btn btn-primary" onClick={() => onConfirm(boxes, boxes.length)}>
            {index + 1 < total ? 'Xác nhận · Ảnh tiếp →' : 'Xác nhận'}
          </button>
        </div>
      </div>
    </div>
  )
}
