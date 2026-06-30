# Khoanh vùng box, chỉnh sửa thủ công & thu thập dataset

Ngày: 2026-06-30

## Mục tiêu

Từ ảnh người dùng tải lên, hiển thị các vùng (box) mà model phát hiện, cho phép
người dùng vẽ thêm / xoá bớt box để chỉnh số đếm cho chính xác, và lưu lại ảnh
gốc kèm nhãn (label) theo định dạng YOLO để làm dữ liệu huấn luyện sau này.

Ngoài ra nâng ngưỡng tin cậy (confidence threshold) mặc định lên 0.5.

## Phạm vi

Bao gồm:
- Dịch vụ CV trả về toạ độ box (chuẩn hoá 0..1) bên cạnh số đếm.
- Backend chuyển tiếp toạ độ box và thêm endpoint lưu dataset.
- Frontend: modal xem lại + vẽ/xoá box, cộng số đếm, lưu dataset.

Cố tình KHÔNG làm (YAGNI):
- Không lưu ảnh đã-vẽ-box (annotated) lên server — chỉ lưu ảnh gốc + label.
- Không sửa box cho ảnh đã lưu trước đó.
- Không kéo-giãn (resize) box đã vẽ — chỉ vẽ mới / xoá.

## Kiến trúc & luồng dữ liệu

```
Upload ảnh
   │
   ▼
POST /lots/count-boxes ──► box-counter /count
   │                         (YOLO, conf=0.5)
   │  ◄── { total, per_image:[{ filename, count, width, height,
   │                            boxes:[{x1,y1,x2,y2,conf}] }] }   (toạ độ 0..1)
   ▼
BoxReviewModal (từng ảnh một, theo hàng đợi)
   │  vẽ thêm box (+1) / xoá box sai (-1) → final boxes
   ├──► POST /lots/dataset  (image gốc + labels YOLO)  ──► MinIO dataset/
   ▼
count cuối = số box còn lại → cộng vào ô Số lượng + lưu kèm ảnh (luồng cũ)
```

## Thành phần

### 1. Dịch vụ CV — `box-counter-service/app.py`

- `CONF_THRESHOLD` mặc định: `0.3` → `0.5`.
- Hàm đếm trả về cả danh sách box (không chỉ số đếm). Mỗi box gồm toạ độ
  chuẩn hoá theo ảnh đã sửa EXIF: `x1, y1, x2, y2 ∈ [0,1]` và `conf`.
  Áp dụng cùng bộ lọc hiện có (đúng lớp `box`, qua ngưỡng diện tích
  `MIN_BOX_AREA_PCT`). `count` = số box trả về.
- `/count` response mở rộng (tương thích ngược):
  ```json
  {
    "total": 12,
    "per_image": [
      { "filename": "a.jpg", "count": 12, "width": 4032, "height": 3024,
        "boxes": [ { "x1": 0.10, "y1": 0.20, "x2": 0.25, "y2": 0.40, "conf": 0.87 } ] }
    ]
  }
  ```
- Lý do dùng toạ độ chuẩn hoá: khớp ảnh hiển thị ở mọi độ phân giải, và đổi
  thẳng sang định dạng YOLO (`xc yc w h`) khi lưu dataset.

### 2. Backend Go

**a. `internal/modules/lot/usecase/count_boxes.go`**
- `BoxCountItem` thêm: `Width float64`, `Height float64`, `Boxes []Box`.
- Thêm struct `Box { X1, Y1, X2, Y2, Conf float64 }` với JSON tag khớp Python.
- Chỉ pass-through; không xử lý thêm.

**b. Endpoint mới `POST /lots/dataset`**
- Handler `dataset.go` + usecase `save_dataset.go`, theo pattern
  `upload_lot_image.go`.
- Nhận multipart: `image` (file gốc), `labels` (text YOLO, mỗi dòng
  `0 xc yc w h`).
- Validate: kích thước ≤ 10MB, content-type ảnh hợp lệ (dùng lại
  `maxImageSize` / `allowedImageTypes`).
- Lưu vào MinIO (bucket hiện có):
  - ảnh: `dataset/images/<uuid>.<ext>`
  - nhãn: `dataset/labels/<uuid>.txt`  (cùng `<uuid>` để ghép cặp)
- Trả `{ "ok": true }`. Nếu lưu nhãn lỗi thì dọn ảnh đã lưu (tránh rác).
- Đăng ký route trong `internal/modules/lot/routes.go` (cùng nhóm auth như
  `count-boxes`).

### 3. Frontend

**a. `types/index.ts`**
- `BoxCountItem` thêm `width?, height?, boxes?: DetBox[]`.
- `DetBox = { x1, y1, x2, y2, conf }` (chuẩn hoá 0..1).

**b. `api/client.ts`**
- `saveDataset(file: File, boxes: {x1,y1,x2,y2}[])`: dựng text YOLO
  (`0 xc yc w h`, xc=(x1+x2)/2, yc=(y1+y2)/2, w=x2-x1, h=y2-y1; mỗi dòng một
  box), gắn `image` + `labels` vào FormData, `POST /lots/dataset`.

**c. Component mới `components/BoxReviewModal.tsx`**
- Props: `file: File`, `initialBoxes: DetBox[]`, `index`, `total` (vị trí trong
  hàng đợi để hiển thị "Ảnh 1/3"), `onConfirm(boxes, count)`, `onCancel()`.
- Hiển thị ảnh gốc (object URL) trong khung; lớp **SVG phủ** lên trên vẽ các
  box. Box lưu ở dạng chuẩn hoá; quy đổi sang pixel theo kích thước ảnh hiển
  thị thực (đo bằng ref + `onLoad`/ResizeObserver).
- Tương tác:
  - **Vẽ thêm**: mousedown/touchstart kéo tạo box mới; thả chuột chốt box.
    Bỏ qua box quá nhỏ (chống chạm nhầm).
  - **Xoá**: bấm chọn 1 box → hiện nút ✕ → xoá. (Không resize.)
  - Hỗ trợ cả chuột và cảm ứng (dùng pointer events) cho điện thoại.
- Hiển thị số đếm trực tiếp = số box hiện có.
- Nút **Xác nhận** → gọi `onConfirm(boxes, boxes.length)`.

**d. `components/LotModal.tsx`**
- `handleUpload`: sau khi gọi `countBoxes`, KHÔNG cộng tổng ngay. Thay vào đó
  mở hàng đợi review: lần lượt mở `BoxReviewModal` cho từng ảnh với
  `initialBoxes` từ `per_image[i].boxes`.
- Mỗi lần Xác nhận ảnh i:
  - lưu `count_i = boxes.length`,
  - gọi `saveDataset(file_i, boxes_i)` (lỗi dataset chỉ cảnh báo, không chặn).
- Sau ảnh cuối: cộng `Σ count_i` vào ô Số lượng và đưa ảnh vào `pending`
  (lô mới) hoặc `uploadLotImages` (lô đang sửa) — đúng luồng hiện tại, chỉ
  thay nguồn `count` bằng số đã chỉnh tay.
- State mới: hàng đợi review `{ file, boxes }[]` + chỉ số ảnh đang review.
- Cho phép **Huỷ** một ảnh trong hàng đợi (bỏ qua ảnh đó, không tính, không
  lưu dataset).

## Xử lý lỗi

- `/count` lỗi: giữ hành vi hiện tại (hiện `imgError`, count = 0).
- `saveDataset` lỗi: chỉ cảnh báo nhẹ, không chặn việc cộng số lượng/upload ảnh.
- Box vẽ quá nhỏ: bỏ qua, không thêm.

## Kiểm thử

- Python: test hàm detect trả về boxes chuẩn hoá đúng (toạ độ trong [0,1],
  count khớp số box). Test `/count` response có trường `boxes`.
- Go: test endpoint `/lots/dataset` lưu đúng 2 object (image + labels) và dọn
  rác khi lỗi.
- Frontend: kiểm thử thủ công luồng upload → review → vẽ/xoá → xác nhận → số
  lượng cập nhật đúng; kiểm tra dataset xuất hiện trong MinIO.
```
