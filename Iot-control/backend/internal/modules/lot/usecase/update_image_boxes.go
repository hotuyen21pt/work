package usecase

import (
	"context"

	httperrors "lot-control/pkg/httperrors"
)

// UpdateImageBoxes cập nhật box của một ảnh lô (mở lại chỉnh tay). boxesJSON là
// JSON danh sách box chuẩn hoá 0..1. edited=true thì ghi/cập nhật nhãn dataset,
// ngược lại xoá nhãn (ảnh quay về auto-detect, không nằm trong dataset).
func (uc *lotUseCase) UpdateImageBoxes(ctx context.Context, lotID, imageID int64, boxesJSON []byte, edited bool) error {
	img, err := uc.lotRepo.GetImageByID(imageID)
	if err != nil {
		return err
	}
	// Đảm bảo ảnh thuộc đúng lô trên URL (tránh sửa nhầm qua lô khác).
	if img.LotID != lotID {
		return httperrors.NewNotFound("không tìm thấy ảnh trong lô này")
	}

	// Cập nhật cả boxes lẫn count (=số box) để số lượng luôn khớp sau khi chỉnh tay.
	if err := uc.lotRepo.UpdateImageBoxes(imageID, boxesJSON, countBoxesJSON(boxesJSON)); err != nil {
		return err
	}

	// Dataset (ảnh + nhãn) chỉ lưu cho ảnh chỉnh tay; còn lại xoá nếu trước đó có.
	if edited && len(boxesJSON) > 0 {
		if err := uc.saveDataset(ctx, img.ObjectKey, boxesJSON); err != nil {
			uc.logger.Warnf("không lưu được dataset cho %q: %v", img.ObjectKey, err)
		}
	} else {
		uc.removeDataset(ctx, img.ObjectKey)
	}
	return nil
}
