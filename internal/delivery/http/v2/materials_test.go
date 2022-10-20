package v2

import (
	"encoding/json"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (ht *HandlersTestSuite) TestHandler_CreateNewTextData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		login string
		value string
		want  want
	}{
		{
			name:  "body is wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"text":12345, "metadata":"rrfrfrrf"}`,
			want:  want{code: 400},
		},
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"text":"12345", "metadata":"rrfrfrrf"}`,
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetBody(tt.value).SetHeader("Content-Type", "application/json; charset=utf8").Post(ht.ts.URL + "/api/materials/text")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_getAllTextData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "wrong user",
			value: `{"login":"test", "password":"test"}`,
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
			res := struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{
				AccessToken:  "",
				RefreshToken: "",
			}
			if tt.want.code == 200 {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			} else {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			}

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).Get(ht.ts.URL + "/api/materials/text")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_UpdateTextDataByID() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		login string
		want  want
	}{
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":1,"text":"user1", "metadata":"user1"}`,
			want:  want{code: 200},
		},
		{
			name:  "wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":123,"text":user1, "metadata":"user1"}`,
			want:  want{code: 400},
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Put(ht.ts.URL + "/api/materials/text")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_CreateNewCredData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		login string
		value string
		want  want
	}{
		{
			name:  "body is wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"login":12345, "password":"12344","metadata":"rrfrfrrf"}`,
			want:  want{code: 400},
		},
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"login":"12345", "password":"12344", "metadata":"rrfrfrrf"}`,
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetBody(tt.value).SetHeader("Content-Type", "application/json; charset=utf8").Post(ht.ts.URL + "/api/materials/cred")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_getAllCredData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "wrong user",
			value: `{"login":"test", "password":"test"}`,
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
			res := struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{
				AccessToken:  "",
				RefreshToken: "",
			}
			if tt.want.code == 200 {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			} else {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			}

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).Get(ht.ts.URL + "/api/materials/cred")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_UpdateCredDataByID() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		login string
		want  want
	}{
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":1,"login":"user1", "password":"213343",  "metadata":"user1"}`,
			want:  want{code: 200},
		},
		{
			name:  "wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":123,"login":user1, "password":"213343","metadata":"user1"}`,
			want:  want{code: 400},
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Put(ht.ts.URL + "/api/materials/cred")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_CreateNewCardData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		login string
		value string
		want  want
	}{
		{
			name:  "body is wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"card_number":12312312342,"exp_date": "2002-01-01 00:00:00 +0000 UTC","cvc":"234","name":"rgrgr","surname":"rgrgrg","metadata":"user"}`,
			want:  want{code: 400},
		},
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"card_number":"12312312342","exp_date": "2002-01-01 00:00:00 +0000 UTC","cvc":"234","name":"rgrgr","surname":"rgrgrg","metadata":"user"}`,
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetBody(tt.value).SetHeader("Content-Type", "application/json; charset=utf8").Post(ht.ts.URL + "/api/materials/card")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_getAllCardData() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		want  want
	}{
		{
			name:  "wrong user",
			value: `{"login":"test", "password":"test"}`,
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
			res := struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			}{
				AccessToken:  "",
				RefreshToken: "",
			}
			if tt.want.code == 200 {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			} else {
				respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Post(ht.ts.URL + "/api/user/auth/sign-in")
				json.Unmarshal(respSing.Body(), &res)
			}

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).Get(ht.ts.URL + "/api/materials/card")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}

func (ht *HandlersTestSuite) TestHandler_UpdateCardDataByID() {
	type want struct {
		code int
	}
	tests := []struct {
		name  string
		value string
		login string
		want  want
	}{
		{
			name:  "success",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":1,"card_number":"12312312342","month":"04","year":"2012","cvc":"234","name":"rgrgr","surname":"rgrgrg","metadata":"user"}`,
			want:  want{code: 200},
		},
		{
			name:  "wrong",
			login: `{"login":"user", "password":"user"}`,
			value: `{"id":1,"card_number":12312312342,"month":"04","year":"2012","cvc":"234","name":"rgrgr","surname":"rgrgrg","metadata":"user"}`,
			want:  want{code: 400},
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

			respSing, _ := client.R().SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.login).Post(ht.ts.URL + "/api/user/auth/sign-in")
			json.Unmarshal(respSing.Body(), &res)

			resp, err := client.R().SetHeader("Authorization", res.AccessToken).SetHeader("Content-Type", "application/json; charset=utf8").SetBody(tt.value).Put(ht.ts.URL + "/api/materials/card")
			require.NoError(t, err)
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}
