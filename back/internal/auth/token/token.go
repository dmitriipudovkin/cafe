package token

import "github.com/golang-jwt/jwt"

type Tokenizer struct {
	sign string
}

func NewTokenizer(sign string) *Tokenizer {
	return &Tokenizer{sign: sign}
}

func (t *Tokenizer) GetToken(claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))

	return token.SignedString([]byte(t.sign))
}
