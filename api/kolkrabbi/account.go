package kolkrabbi

import (
	"github.com/davidji99/simpleresty"
	"time"
)

type AccountInfo struct {
	ID        *string            `json:"id"`
	Heroku    *AccountInfoHeroku `json:"heroku"`
	Github    *AccountInfoGithub `json:"github"`
	CreatedAt *time.Time         `json:"created_at"`
	UpdatedAt *time.Time         `json:"updated_at"`
}

type AccountInfoGithub struct {
	UserID *int    `json:"user_id"`
	Token  *string `json:"token"`
}

type AccountInfoHeroku struct {
	UserID *string `json:"user_id"`
}

func (k *Kolkrabbi) GetAccountInfo() (*AccountInfo, *simpleresty.Response, error) {
	var result AccountInfo
	urlStr := k.http.RequestURL("/account/github/token")

	// Execute the request
	response, getErr := k.http.Get(urlStr, &result, nil)

	return &result, response, getErr
}
