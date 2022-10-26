package v2

import (
	"database/sql"
	"encoding/json"
	"log"
	htt "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"gophkeeper/internal/app"
	"gophkeeper/internal/config"
	"gophkeeper/internal/delivery/http"
	"gophkeeper/internal/delivery/http/v2"
	"gophkeeper/internal/service"
	"gophkeeper/internal/storage"
	"gophkeeper/pkg/auth"
	"gophkeeper/pkg/hash"
)

type HandlersTestSuite struct {
	suite.Suite
	cfg          *config.Config
	db           *sql.DB
	storages     *storage.Storages
	hasher       *hash.SHA1Hasher
	tokenManager *auth.Manager
	handler      *v2.Handler
	ts           *httptest.Server
	refresh      string
	access       string
}

func (ht *HandlersTestSuite) SetupTest() {
	cfg := config.Config{}
	cfg.Addr = "http://127.0.0.1:8081"
	cfg.PasswordSalt = "secret"
	cfg.DatabaseDSN = "postgres://ivanmyagkov@localhost:5432/gophkeeper?sslmode=disable"

	db, err := app.NewInPSQL(cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	storages := storage.NewStorages(db)

	hasher := hash.NewSHA1Hasher(cfg.PasswordSalt)
	tokenManager, err := auth.NewManager(cfg.PasswordSalt)
	if err != nil {
		log.Fatal(err)
	}

	deps := service.Deps{
		Storages:        storages,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AccessTokenTTL:  10 * time.Minute,
		RefreshTokenTTL: 40 * 24 * time.Hour,
	}

	services := service.NewServices(deps)

	// HTTP server
	handlers := http.NewHandler(services, tokenManager)
	ht.cfg = &cfg
	ht.storages = deps.Storages
	ht.handler = v2.NewHandler(services, deps.TokenManager)
	ht.ts = httptest.NewServer(handlers.InitEcho())
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (ht *HandlersTestSuite) TestHandler_userSignUp() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "body is wrong",
			value: `{"login":111111, "password":"user"}`,
			want:  want{code: 400},
		},
		{
			name:  "success",
			value: `{"login":"user", "password":"user"}`,
			want:  want{code: 200},
		},
		{
			name:  "already exists",
			value: `{"login":"user", "password":"user"}`,
			want:  want{code: 409},
		},
	}
	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			resp, err := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-up")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_userSignIn() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "body is wrong",
			value: `{"login":111111, "password":"user"}`,
			want:  want{code: 400},
		},
		{
			name:  "user is not found",
			value: `{"login":"user123", "password":"user123"}`,
			want:  want{code: 401},
		},
		{
			name:  "success",
			value: `{"login":"user", "password":"user"}`,
			want:  want{code: 200},
		},
	}
	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			resp, err := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_userRefresh() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "token is wrong",
			value: "d2dd615e638b268d2acb4a3ad7f48c15ce370cf0bbeedf4fbf62f00670013",
			want:  want{code: 500},
		},
		{
			name:  "success",
			value: "",
			want:  want{code: 200},
		},
	}
	for _, tt := range tests {
		ht.T().Run(tt.name, func(t *testing.T) {
			client := resty.New()
			res := struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{
				AccessToken:  "",
				RefreshToken: "",
			}
			if tt.value == "" {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(`{"login":"user", "password":"user"}`).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			} else {
				res.RefreshToken = tt.value
			}
			resp, err := client.R().SetCookie(&htt.Cookie{Name: "RefreshToken", Value: res.RefreshToken}).Post(ht.ts.URL + "/api/user/auth/refresh")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}
