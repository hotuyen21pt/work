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
}
