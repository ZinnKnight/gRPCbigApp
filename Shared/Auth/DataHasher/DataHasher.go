package DataHasher

import "golang.org/x/crypto/bcrypt"

const Cost = bcrypt.DefaultCost

func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), Cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CompareHashAndPassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}
