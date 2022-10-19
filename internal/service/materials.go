package service

import (
	"context"
	"gophkeeper/internal/domain"
	"gophkeeper/internal/storage"
)

type MaterialsService struct {
	storage storage.Materials
}

func NewMaterialsService(s storage.Materials) *MaterialsService {
	return &MaterialsService{
		storage: s,
	}
}

func (s *MaterialsService) GetAllTextData(ctx context.Context, userID int) ([]domain.TextData, error) {
	return s.storage.GetAllTextData(ctx, userID)
}

func (s *MaterialsService) UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error {
	return s.storage.UpdateTextDataByID(ctx, userID, data)
}

func (s *MaterialsService) CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error {
	return s.storage.CreateNewTextData(ctx, userID, data)
}

//**********************************************************************************************************************
func (s *MaterialsService) GetAllCardData(ctx context.Context, userID int) ([]domain.CardData, error) {
	return s.storage.GetAllCardData(ctx, userID)
}

func (s *MaterialsService) UpdateCardDataByID(ctx context.Context, userID int, data domain.CardData) error {
	return s.storage.UpdateCardDataByID(ctx, userID, data)
}

func (s *MaterialsService) CreateNewCardData(ctx context.Context, userID int, data domain.CardData) error {
	return s.storage.CreateNewCardData(ctx, userID, data)
}

//**********************************************************************************************************************
func (s *MaterialsService) GetAllCredData(ctx context.Context, userID int) ([]domain.CredData, error) {
	return s.storage.GetAllCredData(ctx, userID)
}

func (s *MaterialsService) UpdateCredDataByID(ctx context.Context, userID int, data domain.CredData) error {
	return s.storage.UpdateCredDataByID(ctx, userID, data)
}

func (s *MaterialsService) CreateNewCredData(ctx context.Context, userID int, data domain.CredData) error {
	return s.storage.CreateNewCredData(ctx, userID, data)
}
