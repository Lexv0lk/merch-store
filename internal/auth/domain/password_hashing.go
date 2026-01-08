package domain

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(password, hashedPassword string) (bool, error)
}
