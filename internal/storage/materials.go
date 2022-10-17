package storage

import (
	"context"
	"database/sql"
	"errors"
	"gophkeeper/internal/domain"
)

type MaterialsStorage struct {
	db *sql.DB
}

func NewMaterialsStorage(db *sql.DB) *MaterialsStorage {
	return &MaterialsStorage{
		db: db,
	}
}

func (r *MaterialsStorage) GetAllTextData(ctx context.Context, userID int) ([]domain.TextData, error) {
	getTextDataStmt, err := r.db.PrepareContext(ctx, "SELECT id,data,metadata FROM text_data WHERE user_id=$1;")
	if err != nil {
		return nil, &StatementPSQLError{Err: err}
	}
	defer getTextDataStmt.Close()

	rows, err := getTextDataStmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, &ExecutionPSQLError{Err: err}
	}
	defer rows.Close()

	allTextData := make([]domain.TextData,0)
	for rows.Next() {
		var textData domain.TextData
		err = rows.Scan(&textData.ID, &textData.Text, &textData.Metadata)
		if err != nil {
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return nil, &NotFoundError{Err: domain.ErrDataNotFound}
			default:
				return nil, &ExecutionPSQLError{Err: err}
			}
		}

		allTextData = append(allTextData, textData)
	}

	err = rows.Err()
	if err != nil {
		return nil, &ExecutionPSQLError{Err: err}
	}

	return allTextData, nil
}

func (r *MaterialsStorage) UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error {
	updateTextDataStmt, err := r.db.PrepareContext(ctx, "UPDATE text_data SET data = $1, metadata = $2 WHERE user_id = $3 and id = $4;")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer updateTextDataStmt.Close()

	_, err = updateTextDataStmt.ExecContext(ctx, data.Text, data.Metadata, userID, data.ID)
	if err != nil {
		return &ExecutionPSQLError{Err: err}
	}
	return nil
}

func (r *MaterialsStorage) CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error {
	crUserStmt, err := r.db.PrepareContext(ctx, "INSERT INTO text_data (user_id, data, metadata) VALUES ($1, $2, $3);")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if _, err := crUserStmt.ExecContext(ctx, userID, data.Text, data.Metadata); err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *MaterialsStorage) Close() error {
	return r.db.Close()
}