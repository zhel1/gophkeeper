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

//**********************************************************************************************************************
// Text
//**********************************************************************************************************************
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

	allTextData := make([]domain.TextData, 0)
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

	res, err := updateTextDataStmt.ExecContext(ctx, data.Text, data.Metadata, userID, data.ID)
	if err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return &NotFoundError{Err: domain.ErrDataNotFound}
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

//**********************************************************************************************************************
// Credit card
//**********************************************************************************************************************
func (r *MaterialsStorage) GetAllCardData(ctx context.Context, userID int) ([]domain.CardData, error) {
	getTextDataStmt, err := r.db.PrepareContext(ctx, "SELECT id,card_number,exp_date,cvv,name,surname,metadata FROM card_data WHERE user_id=$1;")
	if err != nil {
		return nil, &StatementPSQLError{Err: err}
	}
	defer getTextDataStmt.Close()

	rows, err := getTextDataStmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, &ExecutionPSQLError{Err: err}
	}
	defer rows.Close()

	allTextData := make([]domain.CardData, 0)
	for rows.Next() {
		var textData domain.CardData
		err = rows.Scan(&textData.ID, &textData.CardNumber, &textData.ExpDate, &textData.CVV, &textData.Name, &textData.Surname, &textData.Metadata)
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

func (r *MaterialsStorage) UpdateCardDataByID(ctx context.Context, userID int, data domain.CardData) error {
	updateTextDataStmt, err := r.db.PrepareContext(ctx, "UPDATE card_data SET card_number = $1, exp_date = $2, cvv = $3, name = $4, surname = $5, metadata = $6 WHERE user_id = $7 and id = $8;")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer updateTextDataStmt.Close()

	res, err := updateTextDataStmt.ExecContext(ctx, data.CardNumber, data.ExpDate, data.CVV, data.Name, data.Surname, data.Metadata, userID, data.ID)
	if err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return &NotFoundError{Err: domain.ErrDataNotFound}
	}

	return nil
}

func (r *MaterialsStorage) CreateNewCardData(ctx context.Context, userID int, data domain.CardData) error {
	crUserStmt, err := r.db.PrepareContext(ctx, "INSERT INTO card_data (user_id, card_number, exp_date, cvv, name, surname, metadata) VALUES ($1, $2, $3, $4, $5, $6, $7);")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if _, err := crUserStmt.ExecContext(ctx, userID, data.CardNumber, data.ExpDate, data.CVV, data.Name, data.Surname, data.Metadata); err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *MaterialsStorage) GetAllCredData(ctx context.Context, userID int) ([]domain.CredData, error) {

	getTextDataStmt, err := r.db.PrepareContext(ctx, "SELECT id,login, password, metadata FROM auth_data WHERE user_id=$1;")
	if err != nil {
		return nil, &StatementPSQLError{Err: err}
	}
	defer getTextDataStmt.Close()

	rows, err := getTextDataStmt.QueryContext(ctx, userID)
	if err != nil {
		return nil, &ExecutionPSQLError{Err: err}
	}
	defer rows.Close()

	allTextData := make([]domain.CredData, 0)
	for rows.Next() {
		var textData domain.CredData
		err = rows.Scan(&textData.ID, &textData.Login, &textData.Password, &textData.Metadata)
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

func (r *MaterialsStorage) UpdateCredDataByID(ctx context.Context, userID int, data domain.CredData) error {
	updateTextDataStmt, err := r.db.PrepareContext(ctx, "UPDATE auth_data SET login = $1, password = $2, metadata = $3 WHERE user_id = $4 and id = $5;")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer updateTextDataStmt.Close()

	res, err := updateTextDataStmt.ExecContext(ctx, data.Login, data.Password, data.Metadata, userID, data.ID)
	if err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return &NotFoundError{Err: domain.ErrDataNotFound}
	}

	return nil
}

func (r *MaterialsStorage) CreateNewCredData(ctx context.Context, userID int, data domain.CredData) error {
	crUserStmt, err := r.db.PrepareContext(ctx, "INSERT INTO auth_data (user_id, login, password, metadata) VALUES ($1, $2, $3, $4);")
	if err != nil {
		return &StatementPSQLError{Err: err}
	}
	defer crUserStmt.Close()

	if _, err := crUserStmt.ExecContext(ctx, userID, data.Login, data.Password, data.Metadata); err != nil {
		return &ExecutionPSQLError{Err: err}
	}

	return nil
}

func (r *MaterialsStorage) Close() error {
	return r.db.Close()
}
