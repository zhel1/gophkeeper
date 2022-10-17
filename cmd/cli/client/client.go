package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gophkeeper/internal/domain"
	"io"
	"net/http"
	"time"
)

const (
	SignUpEndpoint = "/api/user/auth/sign-up"
	SignInEndpoint = "/api/user/auth/sign-in"
	RefreshEndpoint = "/api/user/auth/refresh"

	TextDataEndpoint = "/api/materials/text"
)

type GKClient struct {
	addr string
	refreshPeriod time.Duration
	client *http.Client

	tokens domain.Tokens
}

func NewGKClient(addr string) *GKClient {
	return &GKClient{
		addr: addr,
		client: &http.Client{},
		refreshPeriod: 5 * time.Second,
	}
}
//**********************************************************************************************************************
// Auth
//**********************************************************************************************************************
type AuthInput struct {
	Login        string		`json:"login" binding:"required,max=64"`
	Password     string		`json:"password" binding:"required,min=8,max=64"`
}

func (c *GKClient) UserSignUp(ctx context.Context, input AuthInput) (domain.Tokens, error) {
	var t domain.Tokens
	authJson, err := json.Marshal(input)
	if err != nil {
		return t, err
	}
	authJson=authJson
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr + SignUpEndpoint, bytes.NewBufferString(string(authJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil  {
		return t, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return t, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		if err = json.Unmarshal(body, &t); err != nil {
			return t, nil
		}
		c.tokens = t
		return t, nil
	case http.StatusBadRequest:
		return t, domain.ErrUserBadPassword
	case http.StatusConflict:
		return t, domain.ErrUserAlreadyExists
	case http.StatusInternalServerError:
		return t, domain.ErrInternalServerError
	default:
		return t, errors.New(response.Status)
	}
}

func (c *GKClient) UserSignIn(ctx context.Context, input AuthInput) (domain.Tokens, error) {
	var t domain.Tokens
	authJson, err := json.Marshal(input)
	if err != nil {
		return t, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr + SignInEndpoint, bytes.NewBufferString(string(authJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil  {
		return t, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return t, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		if err = json.Unmarshal(body, &t); err != nil {
			return t, err
		}
		c.tokens = t
		return t, nil
	case http.StatusBadRequest:
		return t, domain.ErrUserBadPassword
	case http.StatusUnauthorized:
		return t, domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return t, domain.ErrInternalServerError
	default:
		return t, errors.New(response.Status)
	}
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func (c *GKClient) UserRefresh(ctx context.Context, refreshToken string) (domain.Tokens, error) {
	var t domain.Tokens
	refreshJson, err := json.Marshal(refreshInput{refreshToken})
	if err != nil {
		return t, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr + RefreshEndpoint, bytes.NewBufferString(string(refreshJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil  {
		return t, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return t, err
	}

	switch response.StatusCode {
	case http.StatusOK:
		if err = json.Unmarshal(body, &t); err != nil {
			return t, err
		}
		c.tokens = t
		return t, nil
	case http.StatusBadRequest:
		return t, errors.New(string(http.StatusBadRequest))
	case http.StatusUnauthorized:
		return t, domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return t, domain.ErrInternalServerError
	default:
		return t, errors.New(response.Status)
	}
}

func (c *GKClient) KeepTokensFresh(ctx context.Context) <-chan error {
	errc := make(chan error)
	go func(){
		defer close(errc)
		ticker := time.NewTicker(c.refreshPeriod)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tokens, err := c.UserRefresh(ctx, c.tokens.RefreshToken)
				if err != nil {
					errc <- err
					return
				}
				c.tokens = tokens
			}
		}
	}()
	return errc
}

//**********************************************************************************************************************
// TextData
//**********************************************************************************************************************
func (c *GKClient) GetAllTextData(ctx context.Context) ([]domain.TextData, error) {
	endpoint := c.addr + TextDataEndpoint
	request, err := http.NewRequestWithContext(ctx, c.addr + http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		var result []domain.TextData
		if err = json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
		return result, nil
	case http.StatusUnauthorized:
		return nil, domain.ErrUserNotFound
	case http.StatusNoContent:
		return nil, domain.ErrDataNotFound
	case http.StatusInternalServerError:
		return nil, domain.ErrInternalServerError
	default:
		return nil, errors.New(response.Status)
	}
}

//
//func (c *GKClient) UpdateTextDataByID(ctx context.Context, userID int, data domain.TextData) error {
//
//}
//
//func (c *GKClient) CreateNewTextData(ctx context.Context, userID int, data domain.TextData) error {
//
//}