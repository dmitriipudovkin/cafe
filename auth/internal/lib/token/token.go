package token

import (
	"errors"

	errfmt "auth/internal/lib/errors"

	"github.com/golang-jwt/jwt"
)

type Tokenizer struct {
	sign string
}

type TokenClaims = jwt.MapClaims

var (
	ErrInvalidToken = errors.New("invalid token")
)

func NewTokenizer(sign string) *Tokenizer {
	return &Tokenizer{sign: sign}
}

func (t *Tokenizer) GetToken(claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))

	return token.SignedString([]byte(t.sign))
}

func (t *Tokenizer) VerifyToken(tokenString string) (TokenClaims, error) {
	const op = "token.Tokenizer.VerifyToken"
	formatter := errfmt.NewOpErrorFormatter(op)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.sign), nil
	})

	if err != nil {
		return nil, formatter.Format(err)
	}

	if claims, ok := token.Claims.(TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, formatter.Format(ErrInvalidToken)
}
