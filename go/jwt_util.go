package friezechat

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func GenerateToken(friezeAccessCode string, domainname string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"FriezeAccessCode": friezeAccessCode,
		"DomainName":       domainName,
		"nbf":              time.Now()
	})
	tokenString, err := token.SignedString([]byte("secret"))
	return tokenString, err
}

func VerifyToken(reqToken string) (map[string]interface{}, error) {
	token, err := jwt.Parse(reqToken, func(t *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return nil, errors.New("math: square root of negative number")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("math: square root of negative number")
	}

}
