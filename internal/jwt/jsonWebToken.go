package jwt

import (
	"fmt"
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

// Функция для проверки токена и извлечения полезной нагрузки
func IsAuthorized(tokenString string) (bool, int) {
	// Парсинг и верификация токена
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи токена
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return mySigningKey, nil
	})

	if err != nil {
		return false, 0 // Ошибка при парсинге или верификации токена
	}

	// Проверка валидности токена
	if !token.Valid {
		return false, 0
	}

	// Извлечение полезной нагрузки (claims) из токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Извлекаем user_id из claims
		if userIDFloat, ok := claims["id"].(float64); ok {
			// Приводим float64 к int
			return true, int(userIDFloat)
		}
		return false, 0
	}

	return false, 0
}
