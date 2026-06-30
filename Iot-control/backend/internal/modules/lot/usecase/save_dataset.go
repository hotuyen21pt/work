package usecase

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	httperrors "lot-control/pkg/httperrors"
)

// SaveDataset lưu ảnh gốc kèm nhãn (định dạng YOLO) vào kho lưu trữ để dùng
// làm dữ liệu huấn luyện sau này. Ảnh và nhãn dùng chung <uuid> để ghép cặp:
//   dataset/images/<uuid>.<ext>
//   dataset/labels/<uuid>.txt
func (uc *lotUseCase) SaveDataset(ctx context.Context, image *multipart.FileHeader, labels string) error {
	if image == nil {
		return httperrors.NewBadRequest("không có ảnh để lưu dataset")
	}
	if image.Size > maxImageSize {
		return httperrors.NewBadRequest(fmt.Sprintf("ảnh %q vượt quá 10MB", image.Filename))
	}

	contentType := image.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		return httperrors.NewBadRequest(fmt.Sprintf("định dạng %q không được hỗ trợ (chỉ JPEG/PNG/WebP/GIF)", contentType))
	}

	file, err := image.Open()
	if err != nil {
		return httperrors.NewBadRequest(fmt.Sprintf("không đọc được file %q", image.Filename))
	}
	defer file.Close()

	id := uuid.NewString()
	ext := strings.ToLower(filepath.Ext(image.Filename))
	imageKey := fmt.Sprintf("dataset/images/%s%s", id, ext)
	labelKey := fmt.Sprintf("dataset/labels/%s.txt", id)

	if _, err := uc.storage.Upload(ctx, imageKey, file, image.Size, contentType); err != nil {
		uc.logger.Errorf("lưu ảnh dataset thất bại: %v", err)
		return httperrors.NewInternal("không lưu được ảnh dataset")
	}

	labelBytes := []byte(labels)
	if _, err := uc.storage.Upload(ctx, labelKey, bytes.NewReader(labelBytes), int64(len(labelBytes)), "text/plain"); err != nil {
		// Dọn ảnh đã lưu để không để cặp lệch (ảnh không có nhãn).
		if rmErr := uc.storage.Remove(ctx, imageKey); rmErr != nil {
			uc.logger.Warnf("không dọn được ảnh dataset %q sau lỗi lưu nhãn: %v", imageKey, rmErr)
		}
		uc.logger.Errorf("lưu nhãn dataset thất bại: %v", err)
		return httperrors.NewInternal("không lưu được nhãn dataset")
	}
	return nil
}
