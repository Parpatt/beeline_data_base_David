package jwt

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var mySigningKey = []byte("superSecretKey") //секретнйы ключь для подписи

func GenerateJWT(name string) (string, error) { //функция, которая возвращает строковый формат JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{ //токен, который выдают пользователю
		"name": name,
		"exp":  time.Now().Add(time.Minute * 30).Unix(), //время жизни токена
	})

	tokenString, err := token.SignedString(mySigningKey) //генерация токена в строковом формате
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func homePage(w http.ResponseWriter, r *http.Request) { //фукция, которая срабатыват, когда пользователь подтверждает, что он тот, кем является
	fmt.Fprintf(w, "Welcome to the Home Page!")
}

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler { //функция, которая проверяет, корректный ли токен мы отправляем
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}

			if token.Valid {
				endpoint(w, r)
			} else {
				fmt.Fprintf(w, "Invalid Authorization Token")
			}
		} else {
			fmt.Fprintf(w, "No Authorization Token provided")
		}
	})
}

func Autcation_and_autzation() {
	http.Handle("/autorization", IsAuthorized(homePage)) //вызываем, когда уже получили токен и хотим доказать свою личность

	http.HandleFunc("/autentification", func(w http.ResponseWriter, r *http.Request) {
		//запускаем в первый раз, когда хотим получить новый токен,
		//получаем ошибку, в случае провала, либо JWT, в текстовом типе

		validToken, err := GenerateJWT("имя") //получаем токен в строковом типе
		if err != nil {
			fmt.Fprintf(w, err.Error()) //выводим ошибку, при её наличии,
		}

		fmt.Fprintf(w, validToken) //либо, выводим токен
	})
}
