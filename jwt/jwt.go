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

type CognitoClaims struct {
	Username string `json:"username"`
	TokenUse string `json:"token_use"`
}

type EnlightClaims struct {
	EnlightUserID         string `json:"enlightUserId"`
	EnlightEmail          string `json:"enlightEmail"`
	EnlightCompanyID      string `json:"enlightCompanyId"`
	EnlightEulaAgreedDate string `json:"enlightEulaAgreedDate"`
	EnlightValidEula      string `json:"enlightValidEula"`
	EnlightName           string `json:"enlightName"`
	EnlightStatus         string `json:"enlightStatus"`
	EnlightRoles          string `json:"enlightRoles"`
	EnlightAccess         string `json:"enlightAccess"`
}

type Claims struct {
	jwt.StandardClaims
	CognitoClaims
	EnlightClaims
}

func (c Claims) Valid() (err error) {
	if err = c.StandardClaims.Valid(); err != nil {
		return
	}

	if c.Username == "" {
		return errors.New("missing username in claims")
	}

	const accessToken = "access"
	if c.TokenUse != accessToken {
		return errors.Errorf("wrong type of token: %s, should be: %s", c.TokenUse, accessToken)
	}

	return
}

func Parse(jwtToken string) (_ Token, err error) {
	keySets, err := jwk.GetKeySets()
	if err != nil {
		err = errors.Wrap(err, "failed to get key sets")
		return
	}

	token, err := jwt.ParseWithClaims(
		jwtToken,
		&Claims{},
		func(token *jwt.Token) (_ interface{}, err error) {
			keyID, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("expecting JWT header to have string `kid`")
			}

			key, err := keySets.LookupKeyID(keyID)
			if err != nil {
				err = errors.Wrap(err, "failed to look up key id")
				return
			}

			return key.GetPublicKey()
		},
	)

	if err != nil {
		err = errors.Wrap(err, "parse with claims failed")
		return
	}

	if !token.Valid {
		err = errors.New("token is not valid")
		return
	}

	if err = token.Claims.Valid(); err != nil {
		err = errors.Wrap(err, "failed to validate claims")
		return
	}

	return Token(*token), nil
}
