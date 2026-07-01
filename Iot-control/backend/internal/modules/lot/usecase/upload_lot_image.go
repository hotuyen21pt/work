package usecase

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"lot-control/internal/models"
	httperrors "lot-control/pkg/httperrors"
)

// maxImageSize giới hạn dung lượng mỗi ảnh: 10MB.
const maxImageSize = 10 << 20

// allowedImageTypes là các content-type ảnh được chấp nhận.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// UploadImages lưu ảnh lô kèm box (boxes) và cờ chỉnh tay (edited) đi song song
// theo thứ tự với files. boxes[i] là JSON danh sách box của ảnh i (nil nếu không
// có); edited[i]=true thì ghi thêm file nhãn dataset cho ảnh đó.
func (uc *lotUseCase) UploadImages(ctx context.Context, lotID int64, files []*multipart.FileHeader, counts []int, boxes [][]byte, edited []bool) ([]models.LotImage, error) {
	exists, err := uc.lotRepo.LotExists(lotID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, httperrors.NewNotFound("không tìm thấy lô")
	}
	if len(files) == 0 {
		return nil, httperrors.NewBadRequest("không có file nào được tải lên")
	}

	saved := make([]models.LotImage, 0, len(files))
	for i, fh := range files {
		// counts/boxes/edited đi song song với files theo thứ tự; thiếu thì mặc định.
		count := 0
		if i < len(counts) {
			count = counts[i]
		}
		var boxJSON []byte
		if i < len(boxes) {
			boxJSON = boxes[i]
		}
		isEdited := i < len(edited) && edited[i]
		img, err := uc.uploadOne(ctx, lotID, fh, count, boxJSON, isEdited)
		if err != nil {
			return nil, err
		}
		saved = append(saved, *img)
	}
	return saved, nil
}

func (uc *lotUseCase) uploadOne(ctx context.Context, lotID int64, fh *multipart.FileHeader, count int, boxJSON []byte, edited bool) (*models.LotImage, error) {
	if fh.Size > maxImageSize {
		return nil, httperrors.NewBadRequest(fmt.Sprintf("ảnh %q vượt quá 10MB", fh.Filename))
	}

	contentType := fh.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return nil, httperrors.NewBadRequest(fmt.Sprintf("định dạng %q không được hỗ trợ (chỉ JPEG/PNG/WebP/GIF)", contentType))
	}

	file, err := fh.Open()
	if err != nil {
		return nil, httperrors.NewBadRequest(fmt.Sprintf("không đọc được file %q", fh.Filename))
	}
	defer file.Close()

	// Ảnh upload lưu ở folder lots; tên đồng bộ <YYYYMMDD><stt> để khi đưa vào
	// dataset thì ảnh (bản sao) & nhãn dùng chung tên.
	day := datasetDay()
	seq, err := uc.lotRepo.NextDatasetSeq(day)
	if err != nil {
		return nil, httperrors.NewInternal("không tạo được số thứ tự ảnh")
	}
	ext := strings.ToLower(filepath.Ext(fh.Filename))
	objectKey := fmt.Sprintf("lots/%d/%s%s", lotID, datasetName(day, seq), ext)

	url, err := uc.storage.Upload(ctx, objectKey, file, fh.Size, contentType)
	if err != nil {
		uc.logger.Errorf("upload ảnh lô %d thất bại: %v", lotID, err)
		return nil, httperrors.NewInternal("không tải được ảnh lên kho lưu trữ")
	}

	img := &models.LotImage{
		LotID:     lotID,
		ObjectKey: objectKey,
		URL:       url,
		Count:     count,
	}
	// Chỉ lưu boxes khi có dữ liệu hợp lệ; tránh ghi chuỗi rỗng (JSON không hợp lệ).
	if len(boxJSON) > 0 {
		img.Boxes = boxJSON
	}
	if err := uc.lotRepo.CreateImage(img); err != nil {
		// Cố gắng dọn object đã upload để không để rác.
		if rmErr := uc.storage.Remove(ctx, objectKey); rmErr != nil {
			uc.logger.Warnf("không dọn được object %q sau lỗi DB: %v", objectKey, rmErr)
		}
		return nil, err
	}
	// Ảnh được chỉnh tay thì lưu vào dataset (ảnh + nhãn); lỗi chỉ cảnh báo.
	if edited && len(boxJSON) > 0 {
		if err := uc.saveDataset(ctx, objectKey, boxJSON); err != nil {
			uc.logger.Warnf("không lưu được dataset cho %q: %v", objectKey, err)
		}
	}
	return img, nil
}
