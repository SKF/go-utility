package model

type Config struct {
	Username     string `yaml:"Username"`
	RefreshToken string `yaml:"RefreshToken"`
	SSOURL       string `yaml:"SSOURL"`
}

type Tokens struct {
	IdentityToken string `json:"identityToken"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"RefreshToken"`
}
