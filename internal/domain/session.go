package domain

import "time"

type Session struct {
	RefreshToken string
	ExpiresAt time.Time
}

type Tokens struct {
	AccessToken  string		`json:"access_token"`
	RefreshToken  string	`json:"refresh_token"`
}