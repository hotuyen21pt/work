# Thiết kế: Quét Barcode/QR cho mã SKU

- **Ngày:** 2026-06-24
- **Phạm vi:** Chỉ frontend (React). Không đổi backend, không đổi database.
- **Mục tiêu:** Cho phép nhân viên chi nhánh tìm SKU bằng cách quét barcode hoặc QR, thay vì gõ tay mã SKU.

## 1. Bối cảnh

Dashboard hiện chỉ có một ô tìm kiếm văn bản (`DashboardPage.tsx`), live-search theo `sku_code` hoặc tên sản phẩm với debounce 300ms qua `listSKUs(q)`. README đã liệt kê "Tìm kiếm SKU nhanh (hỗ trợ quét barcode)" như một tính năng nhưng chưa được triển khai.

## 2. Quyết định đã chốt

| Câu hỏi | Quyết định |
|---|---|
| Thiết bị quét | **Cả hai**: camera trình duyệt + máy quét cầm tay (USB/Bluetooth) |
| Nội dung mã | Giá trị quét **chính là `sku_code`** → khớp trực tiếp, không cần trường mới |
| Hành vi sau khi quét | **Đổ giá trị vào ô tìm kiếm** rồi hiển thị danh sách kết quả (không tự chuyển trang) |
| Thư viện đọc camera | **`html5-qrcode`** (đọc cả QR và barcode 1D, tự xử lý quyền + UI camera, chạy tốt trên iOS Safari & Android Chrome) |

## 3. Hai con đường quét

### 3.1. Máy quét cầm tay (USB/Bluetooth) — keyboard wedge
Loại này hoạt động như bàn phím: gõ chuỗi mã rồi gửi Enter. Ô tìm kiếm hiện tại đã live-search theo từng ký tự nên **đã tương thích sẵn**.

- Cải tiến: đảm bảo ô tìm kiếm được focus khi mở Dashboard (đã có `autoFocus`).
- Không cần code thêm cho luồng này ngoài việc giữ nguyên hành vi hiện tại.

### 3.2. Camera trình duyệt — phần chính cần làm
Thêm nút 📷 cạnh ô tìm kiếm trong `.toolbar`. Luồng:

1. Bấm nút 📷 → mở `BarcodeScannerModal`.
2. Modal xin quyền camera và bắt đầu stream, hiển thị khung quét.
3. Khi đọc được một mã hợp lệ → gọi `onScan(code)`.
4. Dashboard `setQuery(code)` và đóng modal → `useEffect` hiện có tự gọi lại `fetchSKUs` và hiển thị kết quả.

## 4. Thành phần code

### 4.1. `frontend/src/components/BarcodeScannerModal.tsx` (mới)
- **Props:** `{ onScan: (code: string) => void; onClose: () => void }`
- **Trách nhiệm:** bọc `html5-qrcode`; khởi động camera khi mount, dừng + dọn dẹp (`clear()`/`stop()`) khi unmount hoặc đóng.
- **Một purpose rõ ràng:** lấy một chuỗi mã từ camera rồi trả về qua `onScan`. Không biết gì về SKU/API.
- **Hành vi:**
  - Quét thành công đầu tiên → gọi `onScan(decodedText)` một lần rồi dừng camera (tránh bắn nhiều lần).
  - Có nút "Đóng" / bấm overlay để hủy → `onClose()`.
- **Cấu hình quét:** ưu tiên camera sau (`facingMode: "environment"`), bật cả định dạng QR và barcode 1D.

### 4.2. `frontend/src/pages/DashboardPage.tsx` (sửa nhẹ)
- Thêm state `showScanner: boolean`.
- Thêm state/khả năng phát hiện hỗ trợ camera để quyết định bật/tắt nút 📷.
- Thêm nút 📷 vào `.toolbar`, cạnh ô tìm kiếm.
- `onScan(code)` → `setQuery(code)` + `setShowScanner(false)`.
- Render `<BarcodeScannerModal>` khi `showScanner === true`.

### 4.3. CSS (`frontend/src/index.css`)
- Style cho nút quét 📷 (đồng bộ với `.btn`).
- Style cho khung modal camera, tái dùng `.modal-overlay` / `.modal` có sẵn.

### 4.4. `frontend/package.json`
- Thêm dependency `html5-qrcode`.

## 5. Xử lý lỗi

| Tình huống | Hành vi |
|---|---|
| Trình duyệt/thiết bị không có camera | Ẩn hoặc vô hiệu hóa nút 📷; gợi ý dùng máy quét cầm tay hoặc gõ tay |
| Người dùng từ chối quyền camera | Hiện thông báo rõ trong modal, có nút đóng |
| Lỗi khởi tạo camera khác | Hiện thông báo lỗi thân thiện, không làm crash trang |
| Quét ra mã không có SKU khớp | Không phải lỗi — ô tìm kiếm hiện trạng thái rỗng "Không tìm thấy SKU nào cho ..." (đã có sẵn) |

## 6. Lưu ý triển khai / vận hành

- Camera của trình duyệt yêu cầu **HTTPS** (hoặc `localhost`). Khi deploy thật cần phục vụ qua HTTPS, nếu không nút camera sẽ không hoạt động. → Ghi chú trong README.
- Máy quét cầm tay không bị ràng buộc HTTPS (chỉ là bàn phím).

## 7. Ngoài phạm vi (YAGNI)

- Không thêm trường `barcode` riêng vào bảng `skus` (vì mã quét = `sku_code`).
- Không tự động chuyển sang trang chi tiết SKU sau khi quét.
- Không tạo SKU mới trực tiếp từ màn hình quét.
- Không hỗ trợ quét hàng loạt nhiều mã liên tiếp.

## 8. Kiểm thử

- **BarcodeScannerModal:** kiểm tra logic gọi `onScan` đúng một lần khi có kết quả, và dọn dẹp camera khi đóng (mock `html5-qrcode`).
- **DashboardPage:** quét → `query` được set → danh sách lọc lại đúng (mock modal trả về một mã).
- **Thủ công:** thử trên điện thoại thật (Android Chrome + iOS Safari) với một QR/barcode chứa mã SKU thật; thử với máy quét cầm tay nếu có.
