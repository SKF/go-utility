package jwt

import (
	"errors"
	"fmt"

	"github.com/SKF/go-utility/v2/jwk"

	"github.com/golang-jwt/jwt/v5"
)

const (
	TokenUseAccess = "access"
	TokenUseID     = "id"
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

func (c Claims) Validate() error {
	switch c.TokenUse {
	case TokenUseAccess:
		if c.Username == "" {
			return fmt.Errorf("missing username in claims")
		}
	case TokenUseID:
		if c.EnlightUserID == "" {
			return fmt.Errorf("missing enlight user ID in claims")
		}
	default:
		return fmt.Errorf("wrong type of token: %s, should be %s or %s", c.TokenUse, TokenUseAccess, TokenUseID)
	}

	return nil
}

func keyFunc(token *jwt.Token) (any, error) {
	keySets, err := jwk.GetKeySets()
	if err != nil {
		return Token{}, fmt.Errorf("failed to get key sets: %w", err)
	}

	keyID, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("expecting JWT header to have string `kid`")
	}

	key, err := keySets.LookupKeyID(keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup key id: %w", err)
	}

	return key.GetPublicKey()
}

func Parse(jwtToken string) (Token, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &Claims{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
			return Token{}, ErrNotValidNow{underlyingErr: err}
		}

		return Token{}, fmt.Errorf("parse with claims failed: %w", err)
	}

	if !token.Valid {
		return Token{}, fmt.Errorf("token is not valid")
	}

	return Token(*token), nil
}
