# Kiểm Soát Số Lô

Hệ thống quản lý số lô hàng theo chi nhánh. Chi nhánh dùng app để quét barcode SKU, nhập số lô, ngày sản xuất, HSD và số lượng.

## Tính năng

- Đăng nhập theo tài khoản chi nhánh
- Tìm kiếm SKU nhanh (hỗ trợ quét barcode)
- Quản lý số lô: thêm / cập nhật / xóa
- Theo dõi: ngày SX, HSD, số lượng mỗi lô
- Tổng hợp số lượng tự động theo SKU
- Admin: tạo SKU, quản lý tài khoản chi nhánh

## Tech stack

| Thành phần | Công nghệ |
|---|---|
| Backend | Go (Gin) |
| Database | SQLite (file local) |
| Frontend | React 18 + TypeScript + Vite |
| Auth | JWT (24h) |

## Cài đặt và chạy

### Yêu cầu

- [Go 1.21+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)

---

### Backend

```bash
cd backend

# Tải dependencies
go mod tidy

# Chạy server (port 8080)
go run main.go
```

Database SQLite sẽ tự động tạo tại `backend/data/lot_control.db`.

Tài khoản mặc định: **admin / admin123**

---

### Frontend

```bash
cd frontend

# Cài dependencies
npm install

# Chạy dev server (port 5173)
npm run dev
```

Mở trình duyệt: http://localhost:5173

---

### Build production

```bash
# Frontend
cd frontend && npm run build
# Output: frontend/dist/

# Backend
cd backend && go build -o lot-control.exe .
```

## Database Schema

```
users     – tài khoản đăng nhập (username, branch, role)
skus      – danh sách SKU (sku_code, name, unit)
lots      – số lô (lot_number, manufacture_date, expiry_date, qty, branch)
```

## API Endpoints

| Method | Endpoint | Mô tả |
|---|---|---|
| POST | `/api/auth/login` | Đăng nhập |
| GET | `/api/auth/me` | Thông tin user hiện tại |
| GET | `/api/skus` | Danh sách SKU (có search `?q=`) |
| POST | `/api/skus` | Tạo SKU mới |
| GET | `/api/skus/:id` | Chi tiết SKU + danh sách lô |
| PUT | `/api/skus/:id` | Cập nhật SKU |
| DELETE | `/api/skus/:id` | Xóa SKU |
| GET | `/api/lots?sku_id=X` | Danh sách lô của SKU |
| POST | `/api/lots` | Thêm / cập nhật lô (upsert) |
| PUT | `/api/lots/:id` | Cập nhật lô |
| DELETE | `/api/lots/:id` | Xóa lô |
| GET | `/api/users` | Danh sách user (admin) |
| POST | `/api/users` | Tạo user (admin) |
| DELETE | `/api/users/:id` | Xóa user (admin) |

## Ví dụ luồng sử dụng

```
SKU: 422493107 — Serum Vitamin C
├── Lô LA001  qty: 48  NSX: 2024-01-01  HSD: 2025-12-31
└── Lô LA002  qty: 52  NSX: 2024-02-01  HSD: 2026-01-31
                       ─────────────────────────
                       Tổng:            100
```

1. Nhân viên đăng nhập bằng tài khoản chi nhánh
2. Quét barcode SKU → app hiển thị trang chi tiết SKU
3. Bấm "Thêm / Cập nhật lô" → nhập số lô + qty
4. Tổng số lượng tự cộng từ tất cả lô

## Quản lý tài khoản chi nhánh

Admin đăng nhập tại `/users` để:
- Tạo tài khoản nhân viên (username, mật khẩu, chi nhánh)
- Xóa tài khoản

## Biến môi trường (tùy chọn)

```bash
PORT=8080          # Port backend (default: 8080)
JWT_SECRET=...     # JWT secret key (default: lot-control-secret-2024)
```
