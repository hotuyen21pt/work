package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"
)

// boxAnno là một bounding box chuẩn hoá 0..1 do frontend gửi lên (kèm conf tuỳ chọn).
type boxAnno struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

// datasetKeys suy ra key ảnh & nhãn trong folder dataset từ object_key của ảnh
// upload (nằm ở folder lots). Ảnh dataset là bản sao cùng tên, nhãn cùng tên .txt:
//   lots/<lotId>/<name>.<ext>  ->  dataset/<name>.<ext> , dataset/<name>.txt
func datasetKeys(objectKey string) (imageKey, labelKey string) {
	base := path.Base(objectKey)                     // <name>.<ext>
	stem := strings.TrimSuffix(base, path.Ext(base)) // <name>
	return "dataset/" + base, "dataset/" + stem + ".txt"
}

// datasetTZ là múi giờ Việt Nam để số thứ tự tính theo đúng ngày địa phương.
var datasetTZ = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Ho_Chi_Minh"); err == nil {
		return loc
	}
	return time.FixedZone("ICT", 7*3600)
}()

// datasetDay trả về ngày hiện tại dạng YYYYMMDD theo giờ Việt Nam.
func datasetDay() string {
	return time.Now().In(datasetTZ).Format("20060102")
}

// datasetName ghép ngày + số thứ tự 4 chữ số, vd 202607010001.
func datasetName(day string, seq int) string {
	return fmt.Sprintf("%s%04d", day, seq)
}

// countBoxesJSON đếm số box trong chuỗi JSON (0 nếu rỗng/không hợp lệ).
func countBoxesJSON(boxesJSON []byte) int {
	if len(boxesJSON) == 0 {
		return 0
	}
	var boxes []boxAnno
	if err := json.Unmarshal(boxesJSON, &boxes); err != nil {
		return 0
	}
	return len(boxes)
}

// boxesToYOLO chuyển danh sách box chuẩn hoá 0..1 sang nhãn YOLO (class 0):
// mỗi dòng "0 xc yc w h".
func boxesToYOLO(boxes []boxAnno) string {
	lines := make([]string, 0, len(boxes))
	for _, b := range boxes {
		xc := (b.X1 + b.X2) / 2
		yc := (b.Y1 + b.Y2) / 2
		w := b.X2 - b.X1
		h := b.Y2 - b.Y1
		if w < 0 {
			w = -w
		}
		if h < 0 {
			h = -h
		}
		lines = append(lines, fmt.Sprintf("0 %.6f %.6f %.6f %.6f", xc, yc, w, h))
	}
	return strings.Join(lines, "\n")
}

// saveDataset lưu một mẫu dữ liệu huấn luyện vào folder dataset: sao chép ảnh
// (từ folder lots) sang dataset/ và ghi file nhãn YOLO cùng tên.
func (uc *lotUseCase) saveDataset(ctx context.Context, objectKey string, boxesJSON []byte) error {
	var boxes []boxAnno
	if len(boxesJSON) > 0 {
		if err := json.Unmarshal(boxesJSON, &boxes); err != nil {
			return fmt.Errorf("nhãn box không hợp lệ: %w", err)
		}
	}
	imageKey, labelKey := datasetKeys(objectKey)
	// Sao chép ảnh sang dataset/ (server-side, không tải về).
	if err := uc.storage.Copy(ctx, objectKey, imageKey); err != nil {
		return err
	}
	content := []byte(boxesToYOLO(boxes))
	_, err := uc.storage.Upload(ctx, labelKey, bytes.NewReader(content), int64(len(content)), "text/plain")
	return err
}

// removeDataset xoá cặp ảnh + nhãn dataset gắn với ảnh lô; lỗi chỉ cảnh báo.
func (uc *lotUseCase) removeDataset(ctx context.Context, objectKey string) {
	imageKey, labelKey := datasetKeys(objectKey)
	if err := uc.storage.Remove(ctx, imageKey); err != nil {
		uc.logger.Warnf("không xóa được ảnh dataset %q: %v", imageKey, err)
	}
	if err := uc.storage.Remove(ctx, labelKey); err != nil {
		uc.logger.Warnf("không xóa được nhãn dataset %q: %v", labelKey, err)
	}
}
