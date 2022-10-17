package service

import (
	"context"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/storage"
)

type MaterialsService struct {
	storage 		storage.Materials
}

func NewMaterialsService(s storage.Materials) *MaterialsService {
	return &MaterialsService{
		storage: s,
	}
}

func (s *MaterialsService)GetAllTextData(ctx context.Context, userID int) ([]domain.TextData, error) {
	return s.storage.GetAllTextData(ctx, userID)
}

func (s *MaterialsService)UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error {
	return s.storage.UpdateTextDataByID(ctx, userID, data)
}

func (s *MaterialsService)CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error {
	return s.storage.UpdateTextDataByID(ctx, userID, data)
}
