package token

import (
	"errors"

	"github.com/golang-jwt/jwt"
)

type Tokenizer struct {
	sign string
}

type TokenClaims = jwt.MapClaims

func NewTokenizer(sign string) *Tokenizer {
	return &Tokenizer{sign: sign}
}

func (t *Tokenizer) GetToken(claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))

	return token.SignedString([]byte(t.sign))
}

func (t *Tokenizer) VerifyToken(tokenString string) (TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.sign), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
