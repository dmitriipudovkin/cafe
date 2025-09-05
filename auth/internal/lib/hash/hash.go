package hash

import (
	"crypto/sha256"
)

type HasherInterface interface {
	Hash(password string) (string, error)
}

type Hasher struct {
	salt string
}

func NewHasher(salt string) *Hasher {
	return &Hasher{salt: salt}
}

func (h *Hasher) Hash(str string) (string, error) {
	hash := sha256.New()
	if _, err := hash.Write([]byte(str)); err != nil {
		return "", err
	}

	return string(hash.Sum([]byte(h.salt))), nil
}
