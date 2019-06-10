package friezechat

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func GenerateTokenWithUserID(friezeAccessCode string, domainName string, userId string, fullname string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"FriezeAccessCode": friezeAccessCode,
		"DomainName":       domainName,
		"nbf":              time.Now(),
		"UserId":           userId,
		"Fullname":         fullname,
	})
	tokenString, err := token.SignedString([]byte("secret"))
	return tokenString, err
}
func GenerateToken(friezeAccessCode string, domainName string, fullname string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"FriezeAccessCode": friezeAccessCode,
		"DomainName":       domainName,
		"nbf":              time.Now(),
		"Fullname":         fullname,
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

/* 	func main() {
		t,_:=GenerateToken("076ae83c-6acf-4e60-98dd-804e642eda06", "localhost:5001" ,"@customer1:private")
		fmt.Println(t)
	}
} */
