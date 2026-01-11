package domain

import "github.com/alexedwards/argon2id"

type ArgonPasswordHasher struct {
}

func NewArgonPasswordHasher() *ArgonPasswordHasher {
	return &ArgonPasswordHasher{}
}

func (ph *ArgonPasswordHasher) HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func (ph *ArgonPasswordHasher) VerifyPassword(password, hashedPassword string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hashedPassword)
}
