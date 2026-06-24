package usecase

import (
	"context"
	"time"

	"lot-control/internal/models"
	"lot-control/internal/modules/lot/dto"
	httperrors "lot-control/pkg/httperrors"
)

func (uc *lotUseCase) Upsert(ctx context.Context, params *dto.UpsertLotRequest) (*models.Lot, error) {
	exists, err := uc.lotRepo.SKUExists(params.SKUID)
	if err != nil {
		uc.logger.Errorf("Upsert lot SKUExists err: %v", err)
		return nil, err
	}
	if !exists {
		return nil, httperrors.NewNotFound("không tìm thấy SKU")
	}

	lot := &models.Lot{
		SKUID:           params.SKUID,
		LotNumber:       params.LotNumber,
		ManufactureDate: params.ManufactureDate,
		ExpiryDate:      params.ExpiryDate,
		Qty:             params.Qty,
		Branch:          params.Branch,
		CountedByID:     params.CountedByID,
		CountedAt:       time.Now(),
		Notes:           params.Notes,
	}

	if err := uc.lotRepo.Upsert(lot); err != nil {
		uc.logger.Errorf("Upsert lot err: %v", err)
		return nil, err
	}

	if err := uc.lotRepo.TouchSKU(params.SKUID); err != nil {
		uc.logger.Errorf("Upsert lot TouchSKU err: %v", err)
		return nil, err
	}

	result, err := uc.lotRepo.GetBySKUAndLotNumber(params.SKUID, params.LotNumber)
	if err != nil {
		uc.logger.Errorf("Upsert lot reload err: %v", err)
		return nil, err
	}
	return result, nil
}
