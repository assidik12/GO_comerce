package hash

import "golang.org/x/crypto/bcrypt"

type CryptoHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(hashedPassword, password string) error
}

type cryptoHasher struct {
	cost int
}

func NewCryptoHasher(cost int) CryptoHasher {
	return &cryptoHasher{
		cost: cost,
	}
}

// ComparePassword implements [CryptoHasher].
func (c *cryptoHasher) ComparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// HashPassword implements [CryptoHasher].
func (c *cryptoHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), c.cost)
	return string(bytes), err
}
