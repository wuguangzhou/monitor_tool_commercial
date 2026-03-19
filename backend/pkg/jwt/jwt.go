package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// 自定义JWT密钥
var jwtSecret = []byte("monitor_tool_commercial_jwt_secret_2026")

// Claims JWT声明（存储用户核心信息，避免存储敏感数据）
type Claims struct {
	UserId int64
	Phone  string
	jwt.RegisteredClaims
}

// 生成JWT token（有效期7天）
func GenerateToken(userId int64, phone string) (string, error) {
	// 设置token有效期
	expirationTime := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		UserId: userId,
		Phone:  phone,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),     //生效时间
			Issuer:    "zhouyou",                          //签发者
		},
	}

	//生成token(使用HS256算法)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//签名并返回token字符串
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	//检验token并返回声明
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
