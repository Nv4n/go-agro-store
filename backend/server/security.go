package server

import "golang.org/x/crypto/bcrypt"

func hashPass(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), bcrypt.MinCost)
}

func comparePass(hash []byte, pass string) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(pass))
}
