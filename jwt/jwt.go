package jwt

import (
	"github.com/SKF/go-utility/jwk"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type Token jwt.Token

func (t Token) GetClaims() Claims {
	c, ok := t.Claims.(*Claims)
	if ok {
		return *c
	}
	return Claims{}
}

type UserClaims struct {
	UserID         string `json:"enlightUserId"`
	Email          string `json:"enlightEmail"`
	CompanyID      string `json:"enlightCompanyId"`
	EulaAgreedDate string `json:"enlightEulaAgreedDate"`
	ValidEula      string `json:"enlightValidEula"`
	Username       string `json:"enlightName"`
	UserStatus     string `json:"enlightStatus"`
	UserRoles      string `json:"enlightRoles"`
	UserAccess     string `json:"enlightAccess"`
}

type Claims struct {
	jwt.StandardClaims
	UserClaims
	Picture string `json:"picture"`
}

func (c Claims) Valid() (err error) {
	if err = c.StandardClaims.Valid(); err != nil {
		return
	}

	if c.Email == "" {
		return errors.New("Missing email in claims")
	}

	return
}

func Parse(jwtToken string) (_ Token, err error) {
	keySets, err := jwk.GetKeySets()
	if err != nil {
		return
	}

	token, err := jwt.ParseWithClaims(
		jwtToken,
		&Claims{},
		func(token *jwt.Token) (_ interface{}, err error) {
			keyID, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("Expecting JWT header to have string `kid`")
			}

			key, err := keySets.LookupKeyID(keyID)
			if err != nil {
				return
			}

			return key.GetPublicKey()
		},
	)

	if err != nil {
		return
	}

	if !token.Valid {
		err = errors.New("Token is not valid")
		return
	}

	if err = token.Claims.Valid(); err != nil {
		return
	}

	return Token(*token), nil
}
