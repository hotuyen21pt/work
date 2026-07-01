import { useEffect, useLayoutEffect, useRef, useState } from 'react'
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
const ZOOM_MIN = 1
const ZOOM_MAX = 6

type Mode = 'draw' | 'move' | 'pan'
type Corner = 'nw' | 'ne' | 'sw' | 'se'
type View = { zoom: number; x: number; y: number }

const CORNER_CURSOR: Record<Corner, string> = {
  nw: 'nwse-resize',
  se: 'nwse-resize',
  ne: 'nesw-resize',
  sw: 'nesw-resize',
}

export default function BoxReviewModal({ src, initialBoxes, index, total, onConfirm, onCancel }: Props) {
  const [boxes, setBoxes] = useState<DetBox[]>(() => initialBoxes.map(normalize))
  const [selected, setSelected] = useState<number | null>(null)
  const [draft, setDraft] = useState<DetBox | null>(null)
  // Chế độ thao tác: 'draw' = vẽ box mới; 'move' = chọn/di chuyển/đổi kích thước
  // box; 'pan' = kéo di chuyển ảnh (khi đã phóng to).
  const [mode, setMode] = useState<Mode>('draw')
  // Kích thước ảnh hiển thị cơ bản (px) khi zoom=1 — vừa khít khung.
  const [size, setSize] = useState({ w: 0, h: 0 })
  // Zoom + tịnh tiến (px) của lớp canvas chứa ảnh + svg.
  const [view, setView] = useState<View>({ zoom: 1, x: 0, y: 0 })

  const imgRef = useRef<HTMLImageElement>(null)
  const viewportRef = useRef<HTMLDivElement>(null)
  const sizeRef = useRef(size)
  sizeRef.current = size
  const drawingRef = useRef(false)
  const startRef = useRef({ x: 0, y: 0 })
  // Kéo di chuyển box: box gốc + vị trí con trỏ (chuẩn hoá) lúc bấm.
  const dragRef = useRef<{ index: number; startX: number; startY: number; orig: DetBox } | null>(null)
  // Kéo góc đổi kích thước: box gốc + góc đang kéo.
  const resizeRef = useRef<{ index: number; corner: Corner; orig: DetBox } | null>(null)
  // Kéo di chuyển ảnh (pan): vị trí con trỏ + view lúc bấm.
  const panRef = useRef<{ startX: number; startY: number; vx: number; vy: number } | null>(null)

  // Tính kích thước hiển thị: phóng ảnh vừa khít khung (92% rộng × 80% cao).
  const computeSize = () => {
    const el = imgRef.current
    if (!el || !el.naturalWidth || !el.naturalHeight) return
    const maxW = window.innerWidth * 0.92 - 64
    const maxH = window.innerHeight * 0.8
    const scale = Math.min(maxW / el.naturalWidth, maxH / el.naturalHeight)
    setSize({
      w: Math.max(1, Math.round(el.naturalWidth * scale)),
      h: Math.max(1, Math.round(el.naturalHeight * scale)),
    })
  }

  useLayoutEffect(() => {
    const onResize = () => computeSize()
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  // Kẹp tịnh tiến để ảnh luôn phủ kín khung (không để lộ khoảng trống).
  const clampView = (v: View, sz: { w: number; h: number }): View => {
    const minX = Math.min(0, sz.w - sz.w * v.zoom)
    const minY = Math.min(0, sz.h - sz.h * v.zoom)
    return {
      zoom: v.zoom,
      x: Math.min(0, Math.max(minX, v.x)),
      y: Math.min(0, Math.max(minY, v.y)),
    }
  }

  // Phóng to/thu nhỏ quanh vị trí con trỏ (clientX/Y), giữ điểm đó cố định.
  const applyZoom = (clientX: number, clientY: number, factor: number) => {
    const vp = viewportRef.current?.getBoundingClientRect()
    const sz = sizeRef.current
    if (!vp || !sz.w) return
    const cx = clientX - vp.left
    const cy = clientY - vp.top
    setView((v) => {
      const nz = Math.min(ZOOM_MAX, Math.max(ZOOM_MIN, v.zoom * factor))
      if (nz === v.zoom) return v
      const bx = (cx - v.x) / v.zoom
      const by = (cy - v.y) / v.zoom
      return clampView({ zoom: nz, x: cx - bx * nz, y: cy - by * nz }, sz)
    })
  }

  const zoomByButton = (factor: number) => {
    const vp = viewportRef.current?.getBoundingClientRect()
    if (!vp) return
    applyZoom(vp.left + vp.width / 2, vp.top + vp.height / 2, factor)
  }
  const resetView = () => setView({ zoom: 1, x: 0, y: 0 })

  // Lăn chuột để zoom — dùng listener native (non-passive) để chặn cuộn trang.
  useEffect(() => {
    const el = viewportRef.current
    if (!el) return
    const onWheel = (e: WheelEvent) => {
      e.preventDefault()
      applyZoom(e.clientX, e.clientY, e.deltaY < 0 ? 1.15 : 1 / 1.15)
    }
    el.addEventListener('wheel', onWheel, { passive: false })
    return () => el.removeEventListener('wheel', onWheel)
  }, [])

  // Lấy toạ độ chuẩn hoá 0..1 từ con trỏ — dùng rect thực của ảnh nên tự đúng
  // kể cả khi đã zoom/pan.
  const toNorm = (clientX: number, clientY: number) => {
    const el = imgRef.current
    if (!el) return { x: 0, y: 0 }
    const r = el.getBoundingClientRect()
    const x = (clientX - r.left) / r.width
    const y = (clientY - r.top) / r.height
    return { x: Math.min(1, Math.max(0, x)), y: Math.min(1, Math.max(0, y)) }
  }

  const onBgPointerDown = (e: React.PointerEvent) => {
    ;(e.target as Element).setPointerCapture(e.pointerId)
    if (mode === 'pan') {
      panRef.current = { startX: e.clientX, startY: e.clientY, vx: view.x, vy: view.y }
      return
    }
    setSelected(null)
    if (mode === 'draw') {
      const p = toNorm(e.clientX, e.clientY)
      startRef.current = p
      drawingRef.current = true
      setDraft({ x1: p.x, y1: p.y, x2: p.x, y2: p.y })
    }
  }

  const onBoxPointerDown = (i: number, e: React.PointerEvent) => {
    if (mode !== 'move') return
    e.stopPropagation()
    setSelected(i)
    ;(e.currentTarget as Element).setPointerCapture(e.pointerId)
    const p = toNorm(e.clientX, e.clientY)
    dragRef.current = { index: i, startX: p.x, startY: p.y, orig: boxes[i] }
  }

  const onResizePointerDown = (corner: Corner, e: React.PointerEvent) => {
    if (selected === null) return
    e.stopPropagation()
    ;(e.currentTarget as Element).setPointerCapture(e.pointerId)
    resizeRef.current = { index: selected, corner, orig: boxes[selected] }
  }

  const onPointerMove = (e: React.PointerEvent) => {
    const pan = panRef.current
    if (pan) {
      setView((v) => clampView({ zoom: v.zoom, x: pan.vx + (e.clientX - pan.startX), y: pan.vy + (e.clientY - pan.startY) }, sizeRef.current))
      return
    }
    const rz = resizeRef.current
    if (rz) {
      const p = toNorm(e.clientX, e.clientY)
      const o = rz.orig
      let nb: DetBox
      switch (rz.corner) {
        case 'nw': nb = { x1: p.x, y1: p.y, x2: o.x2, y2: o.y2, conf: o.conf }; break
        case 'ne': nb = { x1: o.x1, y1: p.y, x2: p.x, y2: o.y2, conf: o.conf }; break
        case 'sw': nb = { x1: p.x, y1: o.y1, x2: o.x2, y2: p.y, conf: o.conf }; break
        default: nb = { x1: o.x1, y1: o.y1, x2: p.x, y2: p.y, conf: o.conf }; break
      }
      const norm = normalize(nb)
      setBoxes((prev) => prev.map((b, idx) => (idx === rz.index ? norm : b)))
      return
    }
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
      const clamp = (val: number, max: number) => Math.min(max, Math.max(0, val))
      const nx1 = clamp(drag.orig.x1 + (p.x - drag.startX), 1 - w)
      const ny1 = clamp(drag.orig.y1 + (p.y - drag.startY), 1 - h)
      setBoxes((prev) =>
        prev.map((b, idx) => (idx === drag.index ? { x1: nx1, y1: ny1, x2: nx1 + w, y2: ny1 + h, conf: b.conf } : b)),
      )
    }
  }

  const onPointerUp = () => {
    panRef.current = null
    if (resizeRef.current) {
      resizeRef.current = null
      return
    }
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

  // Phím tắt: Delete/Backspace xoá box đang chọn; Esc bỏ chọn hoặc đóng.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Delete' || e.key === 'Backspace') {
        if (selected !== null) {
          e.preventDefault()
          removeSelected()
        }
      } else if (e.key === 'Escape') {
        if (selected !== null) setSelected(null)
        else onCancel()
      }
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [selected])

  const switchMode = (m: Mode) => {
    setMode(m)
    if (m !== 'move') setSelected(null)
  }

  const px = (b: DetBox) => ({
    x: b.x1 * size.w,
    y: b.y1 * size.h,
    w: (b.x2 - b.x1) * size.w,
    h: (b.y2 - b.y1) * size.h,
  })

  // Kích thước hiển thị theo màn hình (không đổi khi zoom) cho viền & nút góc.
  const s = (v: number) => v / view.zoom
  const cursor = mode === 'draw' ? 'crosshair' : mode === 'pan' ? (panRef.current ? 'grabbing' : 'grab') : 'default'

  const subtitle =
    mode === 'draw'
      ? 'Kéo để vẽ box mới (được phép bắt đầu bên trong box khác)'
      : mode === 'move'
        ? 'Bấm chọn box · kéo thân để di chuyển · kéo góc để đổi kích thước'
        : 'Kéo để di chuyển ảnh · lăn chuột hoặc nút +/− để phóng to'

  return (
    <div className="modal-overlay" onClick={onCancel}>
      <div className="modal" onClick={(e) => e.stopPropagation()} style={{ maxWidth: 'min(1100px, 95vw)', width: '95vw' }}>
        <button type="button" className="modal-close" onClick={onCancel} aria-label="Đóng">✕</button>
        <div className="modal-header modal-header-icon">
          <span className="modal-icon">🔲</span>
          <div>
            <h2>Kiểm tra & chỉnh box</h2>
            <p className="modal-subtitle">Ảnh {index + 1}/{total} · {subtitle}</p>
            <div style={{ display: 'flex', gap: '0.4rem', marginTop: '0.35rem', flexWrap: 'wrap', alignItems: 'center' }}>
              <button type="button" className={`btn btn-sm ${mode === 'draw' ? 'btn-primary' : 'btn-ghost'}`} onClick={() => switchMode('draw')}>
                ✏️ Vẽ box
              </button>
              <button type="button" className={`btn btn-sm ${mode === 'move' ? 'btn-primary' : 'btn-ghost'}`} onClick={() => switchMode('move')}>
                ✋ Chọn / di chuyển
              </button>
              <button type="button" className={`btn btn-sm ${mode === 'pan' ? 'btn-primary' : 'btn-ghost'}`} onClick={() => switchMode('pan')}>
                🖐 Kéo ảnh
              </button>
              <span style={{ width: 1, height: 20, background: 'var(--gray-300)', margin: '0 0.15rem' }} />
              <button type="button" className="btn btn-sm btn-ghost" onClick={() => zoomByButton(1 / 1.25)} aria-label="Thu nhỏ">➖</button>
              <span style={{ fontSize: '0.8rem', minWidth: 44, textAlign: 'center', color: 'var(--gray-500)' }}>
                {Math.round(view.zoom * 100)}%
              </span>
              <button type="button" className="btn btn-sm btn-ghost" onClick={() => zoomByButton(1.25)} aria-label="Phóng to">➕</button>
              <button type="button" className="btn btn-sm btn-ghost" onClick={resetView} aria-label="Đặt lại">⟲</button>
            </div>
          </div>
        </div>

        <div
          ref={viewportRef}
          style={{
            position: 'relative',
            width: size.w || 'auto',
            height: size.h || 'auto',
            margin: '0 auto',
            overflow: 'hidden',
            borderRadius: 8,
            lineHeight: 0,
            userSelect: 'none',
            touchAction: 'none',
          }}
        >
          <div
            style={{
              position: 'relative',
              width: size.w || 'auto',
              height: size.h || 'auto',
              transform: `translate(${view.x}px, ${view.y}px) scale(${view.zoom})`,
              transformOrigin: '0 0',
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
              style={{ position: 'absolute', top: 0, left: 0, cursor }}
              onPointerMove={onPointerMove}
              onPointerUp={onPointerUp}
            >
              {/* Lớp nền bắt thao tác vẽ box / kéo ảnh / bỏ chọn. */}
              <rect x={0} y={0} width={size.w} height={size.h} fill="transparent" onPointerDown={onBgPointerDown} />
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
                    strokeWidth={s(2)}
                    style={{ cursor: mode === 'move' ? 'move' : 'inherit', pointerEvents: mode === 'move' ? 'auto' : 'none' }}
                    onPointerDown={(e) => onBoxPointerDown(i, e)}
                  />
                )
              })}
              {/* Box đang vẽ. */}
              {draft && (() => {
                const r = px(normalize(draft))
                return (
                  <rect x={r.x} y={r.y} width={r.w} height={r.h} fill="rgba(16,185,129,0.15)" stroke="#10b981" strokeWidth={s(2)} strokeDasharray={`${s(4)} ${s(3)}`} pointerEvents="none" />
                )
              })()}
              {/* Nút 4 góc để đổi kích thước box đang chọn (chế độ di chuyển). */}
              {mode === 'move' && selected !== null && boxes[selected] && (() => {
                const r = px(boxes[selected])
                const hs = s(5)
                const corners: { c: Corner; x: number; y: number }[] = [
                  { c: 'nw', x: r.x, y: r.y },
                  { c: 'ne', x: r.x + r.w, y: r.y },
                  { c: 'sw', x: r.x, y: r.y + r.h },
                  { c: 'se', x: r.x + r.w, y: r.y + r.h },
                ]
                return corners.map((k) => (
                  <rect
                    key={k.c}
                    x={k.x - hs}
                    y={k.y - hs}
                    width={hs * 2}
                    height={hs * 2}
                    fill="#fff"
                    stroke="#dc2626"
                    strokeWidth={s(1.5)}
                    style={{ cursor: CORNER_CURSOR[k.c] }}
                    onPointerDown={(e) => onResizePointerDown(k.c, e)}
                  />
                ))
              })()}
              {/* Nút ✕ xoá box đang chọn — đặt lệch ra ngoài góc trên-phải để
                  không đè lên nút đổi kích thước. */}
              {mode === 'move' && selected !== null && boxes[selected] && (() => {
                const r = px(boxes[selected])
                return (
                  <g
                    transform={`translate(${r.x + r.w + s(13)}, ${r.y - s(13)})`}
                    style={{ cursor: 'pointer' }}
                    onPointerDown={(e) => {
                      e.stopPropagation()
                      removeSelected()
                    }}
                  >
                    <circle cx={0} cy={0} r={s(11)} fill="#dc2626" />
                    <text x={0} y={s(1)} textAnchor="middle" dominantBaseline="middle" fill="#fff" fontSize={s(14)}>✕</text>
                  </g>
                )
              })()}
            </svg>
          </div>
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
