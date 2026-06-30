export interface User {
  id: number
  username: string
  full_name: string
  branch: string
  role: string
  created_at: string
}

export interface SKU {
  id: number
  sku_code: string
  name: string
  unit: string
  total_qty: number
  lot_count: number
  created_at: string
  updated_at: string
  lots?: Lot[]
}

export interface Lot {
  id: number
  sku_id: number
  lot_number: string
  manufacture_date: string
  expiry_date: string
  qty: number
  branch: string
  counted_by: number | null
  counted_by_name: string
  counted_at: string
  notes: string
  images?: LotImage[]
}

export interface LotImage {
  id: number
  lot_id: number
  url: string
  count: number
  created_at: string
}

// DetBox là toạ độ một bounding box, chuẩn hoá 0..1 theo ảnh.
export interface DetBox {
  x1: number
  y1: number
  x2: number
  y2: number
  conf?: number
}

export interface BoxCountItem {
  filename: string
  count: number
  width?: number
  height?: number
  boxes?: DetBox[]
  error?: string
}

export interface BoxCountResult {
  total: number
  per_image: BoxCountItem[]
}
