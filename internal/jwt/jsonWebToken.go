package jwt

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

// MyCustomClaims - структура для хранения утверждений токена
type MyCustomClaims struct {
	Foo string `json:"foo"`
	jwt.StandardClaims
}

var mySigningKey = []byte("superSecretKey") //секретнйы ключь для подписи

func GenerateJWT(name string, user_id int) (string, error) { //функция, которая возвращает строковый формат JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{ //токен, который выдают пользователю
		"name": name,
		"id":   user_id,
		"exp":  time.Now().Add(time.Minute * 30).Unix(), //время жизни токена
	})

	tokenString, err := token.SignedString(mySigningKey) //генерация токена в строковом формате
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func IsAuthorized(rw http.ResponseWriter, tokenString string) (bool, int) { //функция, которая проверяет, корректный ли токен мы отправляем
	// Парсинг и декодирование токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil {
		log.Fatalf("Error parsing token: %v", err)
	}

	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				fmt.Println("Это не токен")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				fmt.Println("Токен истек")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				fmt.Println("Токен еще не действителен")
			} else {
				fmt.Println("Невалидный токен")
			}
			return false, 0
		} else {
			fmt.Println("Не удалось обработать токен")
			return false, 0
		}
	} else if token.Valid {
		fmt.Println("Токен валиден")
		var user_id int
		var flag bool
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Здесь можно получить полезную нагрузку
			fmt.Println("Claims:")
			for key, us_id := range claims {
				if key == "id" {
					user_id = int(us_id.(float64))
					flag = true
				}
			}
		} else {
			// user_id = 0
			// flag = false
			fmt.Println("Invalid token")
		}
		fmt.Println(flag, " ", user_id)
		return flag, user_id
	} else {
		fmt.Println("Недействительный токен")
		return false, 0
	}
}
