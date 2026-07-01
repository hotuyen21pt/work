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

// datasetLabelKey suy ra key file nhãn từ object_key của ảnh. Ảnh lưu thẳng ở
// folder dataset nên nhãn dùng CÙNG tên, chỉ đổi đuôi .txt:
//   dataset/<name>.<ext>  ->  dataset/<name>.txt
func datasetLabelKey(objectKey string) string {
	return strings.TrimSuffix(objectKey, path.Ext(objectKey)) + ".txt"
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

// writeDatasetLabel ghi file nhãn YOLO cho ảnh (ảnh đã nằm sẵn ở dataset/ nên
// chỉ cần ghi nhãn cùng tên). Nhãn rỗng vẫn ghi (ảnh không có box hợp lệ YOLO).
func (uc *lotUseCase) writeDatasetLabel(ctx context.Context, objectKey string, boxesJSON []byte) error {
	var boxes []boxAnno
	if len(boxesJSON) > 0 {
		if err := json.Unmarshal(boxesJSON, &boxes); err != nil {
			return fmt.Errorf("nhãn box không hợp lệ: %w", err)
		}
	}
	labelKey := datasetLabelKey(objectKey)
	content := []byte(boxesToYOLO(boxes))
	_, err := uc.storage.Upload(ctx, labelKey, bytes.NewReader(content), int64(len(content)), "text/plain")
	return err
}

// removeDatasetLabel xoá file nhãn dataset gắn với ảnh; lỗi chỉ cảnh báo.
func (uc *lotUseCase) removeDatasetLabel(ctx context.Context, objectKey string) {
	labelKey := datasetLabelKey(objectKey)
	if err := uc.storage.Remove(ctx, labelKey); err != nil {
		uc.logger.Warnf("không xóa được nhãn dataset %q: %v", labelKey, err)
	}
}
