import type { DetBox } from '../types'

// Chuẩn hoá box để x1<x2, y1<y2 (sau khi kéo có thể ngược chiều).
export const normalize = (b: DetBox): DetBox => ({
  x1: Math.min(b.x1, b.x2),
  y1: Math.min(b.y1, b.y2),
  x2: Math.max(b.x1, b.x2),
  y2: Math.max(b.y1, b.y2),
  conf: b.conf,
})

// Diện tích 1 box (toạ độ chuẩn hoá 0..1).
export const area = (b: DetBox): number => (b.x2 - b.x1) * (b.y2 - b.y1)

// Diện tích phần giao của 2 box (toạ độ chuẩn hoá 0..1).
export const interArea = (a: DetBox, b: DetBox): number => {
  const iw = Math.max(0, Math.min(a.x2, b.x2) - Math.max(a.x1, b.x1))
  const ih = Math.max(0, Math.min(a.y2, b.y2) - Math.max(a.y1, b.y1))
  return iw * ih
}

// IoU của 2 box (toạ độ chuẩn hoá 0..1).
export const iou = (a: DetBox, b: DetBox): number => {
  const inter = interArea(a, b)
  const uni = area(a) + area(b) - inter
  return uni > 0 ? inter / uni : 0
}

// Hai box coi là "chồng" (một trong hai thừa) khi diện tích giao ≥ hệ số này ×
// diện tích của box NHỎ hơn (tức chồng > 0.25 diện tích khung gần đó).
export const OVERLAP_RATIO = 0.25

// Dọn kết quả detect: (1) bỏ box có diện tích < nửa diện tích trung bình các box
// detect được, (2) khử box chồng nhau > OVERLAP_RATIO diện tích của khung gần đó
// (giữ box lớn hơn) — không còn cặp box nào chồng lên nhau quá ngưỡng.
export const cleanupDetections = (raw: DetBox[]): DetBox[] => {
  const normed = raw.map(normalize)
  if (normed.length === 0) return normed
  const avgArea = normed.reduce((s, b) => s + area(b), 0) / normed.length
  const minArea = avgArea / 2
  const kept: DetBox[] = []
  // Duyệt từ box lớn đến nhỏ: giữ box không quá nhỏ và không chồng > ngưỡng.
  for (const b of [...normed].sort((p, q) => area(q) - area(p))) {
    if (area(b) < minArea) continue
    const overlaps = kept.some(
      (k) => interArea(b, k) >= OVERLAP_RATIO * Math.min(area(b), area(k)),
    )
    if (!overlaps) kept.push(b)
  }
  return kept
}
