package password

import "golang.org/x/crypto/bcrypt"

// Hash mengenkripsi password menggunakan bcrypt.
func Hash(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare mengembalikan true jika plain cocok dengan hash.
func Compare(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
