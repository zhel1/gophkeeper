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
	SignUpEndpoint  = "/api/user/auth/sign-up"
	SignInEndpoint  = "/api/user/auth/sign-in"
	RefreshEndpoint = "/api/user/auth/refresh"

	TextDataEndpoint = "/api/materials/text"
	CardDataEndpoint = "/api/materials/card"
	CredDataEndpoint = "/api/materials/cred"
)

type GKClient struct {
	addr          string
	refreshPeriod time.Duration
	client        *http.Client

	tokens domain.Tokens
}

func NewGKClient(addr string) *GKClient {
	return &GKClient{
		addr:          addr,
		client:        &http.Client{},
		refreshPeriod: 25 * time.Second,
	}
}

//**********************************************************************************************************************
// Auth
//**********************************************************************************************************************
type AuthInput struct {
	Login    string `json:"login" binding:"required,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

func (c *GKClient) UserSignUp(ctx context.Context, input AuthInput) (domain.Tokens, error) {
	var t domain.Tokens
	authJson, err := json.Marshal(input)
	if err != nil {
		return t, err
	}
	authJson = authJson
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+SignUpEndpoint, bytes.NewBufferString(string(authJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil {
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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+SignInEndpoint, bytes.NewBufferString(string(authJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil {
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
		return t, errors.New("bad request")
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

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+RefreshEndpoint, bytes.NewBufferString(string(refreshJson)))
	if err != nil {
		return t, err
	}

	response, err := c.client.Do(request)
	if err != nil {
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
		return t, errors.New("bad request")
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
	go func() {
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

func (c *GKClient) SetAccessToken(token string) {
	c.tokens.AccessToken = token
}

func (c *GKClient) SetRefreshToken(token string) {
	c.tokens.RefreshToken = token
}

//**********************************************************************************************************************
// Text
//**********************************************************************************************************************
func (c *GKClient) GetAllTextData(ctx context.Context) ([]domain.TextData, error) {
	endpoint := c.addr + TextDataEndpoint
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

type textDataInput struct {
	Text     string `json:"text"`
	Metadata string `json:"metadata"`
}

func (c *GKClient) CreateNewTextData(ctx context.Context, data domain.TextData) error {
	textDataJson, err := json.Marshal(textDataInput{Text: data.Text, Metadata: data.Text})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, c.addr+TextDataEndpoint, bytes.NewBufferString(string(textDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}

func (c *GKClient) UpdateTextData(ctx context.Context, data domain.TextData) error {
	textDataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+TextDataEndpoint, bytes.NewBufferString(string(textDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}

//**********************************************************************************************************************
// Credit card
//**********************************************************************************************************************
func (c *GKClient) GetAllCardData(ctx context.Context) ([]domain.CardData, error) {
	endpoint := c.addr + CardDataEndpoint
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

		var result []domain.CardData
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

type newCardDataInput struct {
	CardNumber string    `json:"card_number"`
	ExpDate    time.Time `json:"exp_data"`
	CVV        string    `json:"cvv"`
	Name       string    `json:"name"`
	Surname    string    `json:"surname"`
	Metadata   string    `json:"metadata"`
}

func (c *GKClient) CreateNewCardData(ctx context.Context, data domain.CardData) error {
	cardDataJson, err := json.Marshal(newCardDataInput{
		CardNumber: data.CardNumber,
		ExpDate:    data.ExpDate,
		CVV:        data.CVV,
		Name:       data.Name,
		Surname:    data.Surname,
		Metadata:   data.Metadata,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, c.addr+CardDataEndpoint, bytes.NewBufferString(string(cardDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}

func (c *GKClient) UpdateCardData(ctx context.Context, data domain.CardData) error {
	cardDataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+CardDataEndpoint, bytes.NewBufferString(string(cardDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}

//**********************************************************************************************************************
// Creds
//**********************************************************************************************************************
func (c *GKClient) GetAllCredsData(ctx context.Context) ([]domain.CredData, error) {
	endpoint := c.addr + CredDataEndpoint
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
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

		var result []domain.CredData
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

type newCredDataInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Metadata string `json:"metadata"`
}

func (c *GKClient) CreateNewCredData(ctx context.Context, data domain.CredData) error {
	credDataJson, err := json.Marshal(newCredDataInput{
		Login:    data.Login,
		Password: data.Password,
		Metadata: data.Metadata,
	})
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPut, c.addr+CredDataEndpoint, bytes.NewBufferString(string(credDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}

func (c *GKClient) UpdateCredData(ctx context.Context, data domain.CredData) error {
	credDataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.addr+CredDataEndpoint, bytes.NewBufferString(string(credDataJson)))
	if err != nil {
		return err
	}
	request.Header.Add("Authorization", c.tokens.AccessToken)

	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return errors.New("bad request")
	case http.StatusUnauthorized:
		return domain.ErrUserNotFound
	case http.StatusInternalServerError:
		return domain.ErrInternalServerError
	default:
		return errors.New(response.Status)
	}
}
