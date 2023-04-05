package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"

	"github.com/SKF/go-utility/v2/jwk"
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
	Username      string   `json:"username"`
	TokenUse      string   `json:"token_use"`
	CognitoGroups []string `json:"cognito:groups"`
}

type EnlightClaims struct {
	EnlightUserID    string `json:"enlightUserId"`
	EnlightCompanyID string `json:"enlightCompanyId"`
	EnlightAccess    string `json:"enlightAccess"`
	EnlightRoles     string `json:"enlightRoles"`
	EnlightEmail     string `json:"enlightEmail"`
}

type Claims struct {
	jwt.RegisteredClaims
	CognitoClaims
	EnlightClaims
}

const (
	TokenUseAccess = "access"
	TokenUseID     = "id"
)

func (c Claims) Valid() (err error) {
	if err = c.RegisteredClaims.Valid(); err != nil {
		return
	}

	switch c.TokenUse {
	case TokenUseAccess:
		if c.Username == "" {
			return errors.New("missing username in claims")
		}
	case TokenUseID:
		if c.EnlightUserID == "" {
			return errors.New("missing enlight user ID in claims")
		}
	default:
		return errors.Errorf("wrong type of token: %s, should be %s or %s", c.TokenUse, TokenUseAccess, TokenUseID)
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
		validationError := &jwt.ValidationError{}
		if errors.As(err, &validationError) {
			if validationError.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				err = errNotValidNowType{underLyingErr: err}

				return
			}
		}

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
