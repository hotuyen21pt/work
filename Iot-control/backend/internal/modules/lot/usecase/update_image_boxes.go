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

	// Luôn cập nhật nhãn dataset theo box mới (mọi ảnh đều có nhãn).
	_ = edited
	if err := uc.writeDatasetLabel(ctx, img.ObjectKey, boxesJSON); err != nil {
		uc.logger.Warnf("không ghi được nhãn dataset cho %q: %v", img.ObjectKey, err)
	}
	return nil
}
