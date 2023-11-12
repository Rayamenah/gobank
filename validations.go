package main

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func (a *Account) ValidatePassword(pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.EncryptedPassword), []byte(pw)) == nil
}

func ValidateEmail(email string) bool {
	//regex foremail validation
	emailRegex := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-z]{2,}$`
	//compileregex
	reg := regexp.MustCompile(emailRegex)
	//check if email matches regx
	return reg.MatchString(email)

}
