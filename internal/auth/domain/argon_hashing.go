package domain

import "github.com/alexedwards/argon2id"

var optimizedParams = &argon2id.Params{
	Memory:      19 * 1024, // 19 MB
	Iterations:  2,
	Parallelism: 1,
	SaltLength:  16,
	KeyLength:   32,
}

type ArgonPasswordHasher struct {
	params *argon2id.Params
}

func NewArgonPasswordHasher() *ArgonPasswordHasher {
	return &ArgonPasswordHasher{
		params: optimizedParams,
	}
}

func (ph *ArgonPasswordHasher) HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, ph.params)
}

func (ph *ArgonPasswordHasher) VerifyPassword(password, hashedPassword string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hashedPassword)
}
