package model

type Config struct {
	Username     string
	RefreshToken string
	SSOURL       string
}

type Tokens struct {
	IdentityToken string `json:"identityToken"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"RefreshToken"`
}
