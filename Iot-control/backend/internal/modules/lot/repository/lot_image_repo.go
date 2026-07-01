package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"lot-control/internal/models"
	httperrors "lot-control/pkg/httperrors"
)

func (r *lotRepository) LotExists(lotID int64) (bool, error) {
	var count int64
	if err := r.db.Model(&models.Lot{}).Where("id = ?", lotID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *lotRepository) CreateImage(img *models.LotImage) error {
	return r.db.Create(img).Error
}

func (r *lotRepository) ListImagesByLot(lotID int64) ([]models.LotImage, error) {
	var images []models.LotImage
	err := r.db.Where("lot_id = ?", lotID).Order("created_at ASC, id ASC").Find(&images).Error
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (r *lotRepository) GetImageByID(imageID int64) (*models.LotImage, error) {
	var img models.LotImage
	err := r.db.Where("id = ?", imageID).Limit(1).Find(&img).Error
	if err != nil {
		return nil, err
	}
	if img.ID == 0 {
		return nil, httperrors.NewNotFound("không tìm thấy ảnh")
	}
	return &img, nil
}

func (r *lotRepository) UpdateImageBoxes(imageID int64, boxes []byte, count int) error {
	// Rỗng thì lưu NULL (không lưu chuỗi rỗng — sẽ là JSON không hợp lệ khi đọc lại).
	var boxesVal interface{}
	if len(boxes) > 0 {
		boxesVal = boxes
	}
	res := r.db.Model(&models.LotImage{}).Where("id = ?", imageID).
		Updates(map[string]interface{}{"boxes": boxesVal, "count": count})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return httperrors.NewNotFound("không tìm thấy ảnh")
	}
	return nil
}

func (r *lotRepository) DeleteImage(imageID int64) error {
	res := r.db.Delete(&models.LotImage{}, imageID)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return httperrors.NewNotFound("không tìm thấy ảnh")
	}
	return nil
}

// NextDatasetSeq tăng bộ đếm của ngày (upsert) rồi đọc lại giá trị mới trong cùng
// một transaction để đảm bảo an toàn khi có nhiều ảnh upload đồng thời.
func (r *lotRepository) NextDatasetSeq(day string) (int, error) {
	var c models.DatasetCounter
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "day"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"seq": gorm.Expr("dataset_counters.seq + 1")}),
		}).Create(&models.DatasetCounter{Day: day, Seq: 1}).Error; err != nil {
			return err
		}
		return tx.Where("day = ?", day).Take(&c).Error
	})
	if err != nil {
		return 0, err
	}
	return c.Seq, nil
}
