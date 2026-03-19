package encrypt

import "golang.org/x/crypto/bcrypt"

// 密码加密（使用bcrypt,不可逆加密）
func BcryptEncrypt(password string) (string, error) {
	// 生成加密盐值，cos值越大越安全
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func BcryptVerify(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
