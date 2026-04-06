package utils

import "golang.org/x/crypto/bcrypt"

func EncryptGenerate(s string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(s), bcrypt.DefaultCost)
	return string(bytes)
}

func EncryptCheck(s, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(s))
	return err == nil
}
