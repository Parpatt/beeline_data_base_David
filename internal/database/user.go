package database

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"myproject/internal/models"
	"net/http"
	"net/smtp"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Response struct {
	Status  string
	Data    LegalUser
	Message string
}

type LegalUser struct {
	Id            int       `json:"id" db:"id"`
	Password_hash string    `json:"password_hash" db:"password_hash"`
	Email         string    `json:"email" db:"email"`
	Phone_number  string    `json:"phone_number" db:"phone_number"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
	Updated_at    time.Time `json:"updated_at" db:"updated_at"`
	Avatar_path   string    `json:"avatar_path" db:"avatar_path"`
	User_type     int       `json:"user_type" db:"user_type"`
	User_role     int       `json:"user_role" db:"user_role"`

	Ind_num_taxp    int    `json:"ind_num_taxp" db:"ind_num_taxp"`
	Name_of_company string `json:"name_of_company" db:"name_of_company"`
	Address_name    string `json:"address_name" db:"address_name"`
}

func errorr(err error) {
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
}

func (repo *MyRepository) AddNewLegalUserSQL(ctx context.Context, rep *pgxpool.Pool, ind_num_taxp int, name_of_company, address_name, email, phoneNum, hashedPassword string, rw http.ResponseWriter) (err error) {
	result, errors := rep.Query(ctx, `
			WITH i AS (
				INSERT INTO Users.users (
					user_type, password_hash, email, phone_number, updated_at, avatar_path
				) 
				VALUES ($1, $2, $3, $4, $5, $6) 
				RETURNING id
			)
			INSERT INTO Users.company_user (
				user_id, ind_num_taxp, name_of_company, address_name
			)
			SELECT i.id, $7, $8, $9 FROM i
			RETURNING user_id;
		`,
		1,
		hashedPassword,
		email,
		phoneNum,
		time.Now(),
		"C:/",

		ind_num_taxp,
		name_of_company,
		address_name,
	)

	var user_id int
	for result.Next() {
		err := result.Scan(
			&user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	user := LegalUser{
		Id:            user_id,
		Password_hash: hashedPassword,
		Email:         email,
		Phone_number:  phoneNum,
		Updated_at:    time.Now(),
		Avatar_path:   "C:/",
		User_type:     1,
		User_role:     1,

		Ind_num_taxp:    ind_num_taxp,
		Name_of_company: name_of_company,
		Address_name:    address_name,
	}

	if user_id == 0 {
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(Response{
			Status: "fatal",
			Data: LegalUser{
				Id:            0,
				Password_hash: "",
				Email:         "",
				Phone_number:  "",
				Updated_at:    time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
				Avatar_path:   "",
				User_type:     0,
				User_role:     0,

				Ind_num_taxp:    0,
				Name_of_company: "",
				Address_name:    "",
			},
			Message: "Введите корректный & уникальный логин и пароль ",
		})

		return
	} else if errors != nil {
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(Response{
			Status: "fatal",
			Data: LegalUser{
				Id:            0,
				Password_hash: "",
				Email:         "",
				Phone_number:  "",
				Updated_at:    time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
				Avatar_path:   "",
				User_type:     0,
				User_role:     0,

				Ind_num_taxp:    0,
				Name_of_company: "",
				Address_name:    "",
			},
			Message: errors.Error(),
		})

		return
	} else {
		response := Response{
			Status:  "success",
			Data:    user,
			Message: "The user has been successfully registered",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}
}

type NaturUser struct {
	Id            int       `json:"id" db:"id"`
	Password_hash string    `json:"password_hash" db:"password_hash"`
	Email         string    `json:"email" db:"email"`
	Phone_number  string    `json:"phone_number" db:"phone_number"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
	Updated_at    time.Time `json:"updated_at" db:"updated_at"`
	Avatar_path   string    `json:"avatar_path" db:"avatar_path"`
	User_type     int       `json:"user_type" db:"user_type"`
	User_role     int       `json:"user_role" db:"user_role"`

	Name       string `json:"name" db:"name"`
	Surname    string `json:"surname" db:"surname"`
	Patronymic string `json:"patronymic" db:"patronymic"`
}

func (repo *MyRepository) AddNewNaturUserSQL(ctx context.Context, name, surname, patronymic, email, phoneNum, hashedPassword string, rep *pgxpool.Pool, rw http.ResponseWriter) (err error) {
	type Response struct {
		Status  string    `json:"status"`
		Data    NaturUser `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	result, errors := rep.Query(ctx, `
			WITH i AS (
			INSERT INTO Users.users (
				user_type, password_hash, email, phone_number, updated_at, avatar_path
			) 
			VALUES ($1, $2, $3, $4, $5, $6) 
			RETURNING id
			)
			INSERT INTO Users.individual_user (
				user_id, name, surname, patronymic
			)
			SELECT i.id, $7, $8, $9 FROM i
			RETURNING user_id;
		`,
		1,
		hashedPassword,
		email,
		phoneNum,
		time.Now(),
		"C:/",

		name,
		surname,
		patronymic)
	var user_id int
	for result.Next() {
		err := result.Scan(
			&user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	user := NaturUser{
		Id:            user_id,
		Password_hash: hashedPassword,
		Email:         email,
		Phone_number:  phoneNum,
		Updated_at:    time.Now(),
		Avatar_path:   "C:/",
		User_type:     1,
		User_role:     1,

		Name:       name,
		Surname:    surname,
		Patronymic: patronymic,
	}

	if user_id == 0 {
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(Response{
			Status:  "fatal",
			Message: "Введите корректный & уникальный логин и пароль ",
		})

		return
	} else if errors != nil {
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(Response{
			Status:  "fatal",
			Message: errors.Error(),
		})

		return
	} else {
		response := Response{
			Status:  "success",
			Data:    user,
			Message: "The user has been successfully registered",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}
}

func (repo *MyRepository) LoginSQL(ctx context.Context, login, hashedPassword string, rep *pgxpool.Pool, rw http.ResponseWriter) (u models.User, err error, id int) {
	row := rep.QueryRow(ctx,
		`SELECT id, email FROM Users.users WHERE (email = $1 AND password_hash = $2) OR (phone_number = $1 AND password_hash = $2);`,
		login, hashedPassword)
	err = row.Scan(&u.Id, &u.Email)

	errorr(err)

	type User struct {
		Id    int    `json:"id"`
		Login string `json:"login"`
	}

	type Response struct {
		Status  string `json:"status"`
		Data    User   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	user := User{
		Id:    u.Id,
		Login: login,
	}

	if err != nil {
		response := Response{
			Status: "fatal",
			Data: User{
				Id:    0,
				Login: "",
			},
			Message: err.Error(),
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return
	} else {
		response := Response{
			Status:  "success",
			Data:    user,
			Message: "You have successfully logged in",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return u, err, u.Id
	}
}

func (repo *MyRepository) DisputeChatPanelSQL(ctx context.Context, rep *pgxpool.Pool, rw http.ResponseWriter) (err error) {
	type Chat struct {
		ID        int
		User_1_id int
		User_2_id int
		Ad_id     int
	}
	products := []Chat{}

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)

		return
	}

	request, err := rep.Query(
		ctx,
		`SELECT id, user_1_id, user_2_id, ad_id FROM chat.chats 
			WHERE have_disput = true AND (mediator_id IS NULL) AND statee = false;`)

	errorr(err)

	for request.Next() {
		p := Chat{}
		err := request.Scan(
			&p.ID,
			&p.User_1_id,
			&p.User_2_id,
			&p.Ad_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	type Response struct {
		Status  string `json:"status"`
		Data    []Chat `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && request != nil && len(products) != 0 {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) RecoveryPassWithEmailSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, redisClient *redis.Client, user_id int, passwd string) (err error) {
	type Email struct {
		Email_name string
	}
	products := []Email{}

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)

		return
	}

	request, err := rep.Query(
		ctx,
		`SELECT email FROM users.users 
			WHERE id = $1;`,
		user_id)

	errorr(err)

	for request.Next() {
		p := Email{}
		err := request.Scan(
			&p.Email_name,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	type Response struct {
		Status  string  `json:"status"`
		Data    []Email `json:"data,omitempty"`
		Message string  `json:"message"`
	}

	if err == nil && request != nil && len(products) != 0 {
		// Настройки SMTP-сервера
		smtpHost := "smtp.mail.ru"
		smtpPort := "587"

		// Данные отправителя (ваша почта и пароль приложения)
		senderEmail := "parpatt_test@mail.ru"
		password := "X0h72ndPXchhjWZ4vbyT" // Пароль приложения

		// Получатель
		recipientEmail := products[0].Email_name

		// Сообщение
		subject := "Subject: Восстанови свой passwd !.\n"
		body := "Введи этот код.\n"
		codeNum := 777 // Здесь лучше использовать случайный код
		message := []byte(subject + "\n" + body + strconv.Itoa(codeNum))

		// Авторизация для отправки email
		auth := smtp.PlainAuth("", senderEmail, password, smtpHost)

		// Устанавливаем обычное нешифрованное соединение
		client, err := smtp.Dial(smtpHost + ":" + smtpPort)
		if err != nil {
			log.Fatal(err)
		}

		// Используем команду STARTTLS для начала TLS-сессии
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true, // Это нужно убрать в продакшене
			ServerName:         smtpHost,
		}

		if err = client.StartTLS(tlsConfig); err != nil {
			log.Fatal(err)
		}

		// Старт авторизации
		if err = client.Auth(auth); err != nil {
			log.Fatal(err)
		}

		// Установка адреса отправителя
		if err = client.Mail(senderEmail); err != nil {
			log.Fatal(err)
		}

		// Установка адреса получателя
		if err = client.Rcpt(recipientEmail); err != nil {
			log.Fatal(err)
		}

		// Отправка сообщения
		w, err := client.Data()
		if err != nil {
			log.Fatal(err)
		}

		_, err = w.Write(message)
		if err != nil {
			log.Fatal(err)
		}

		err = w.Close()
		if err != nil {
			log.Fatal(err)
		}

		// Завершение сеанса
		client.Quit()

		type Kesh struct {
			Passwd []byte
			Code   int
		}

		hash := sha256.Sum256([]byte(passwd))

		// Преобразуем структуру kesh в JSON
		keshData, err := json.Marshal(Kesh{Passwd: hash[:], Code: codeNum})
		if err != nil {
			log.Fatal("Ошибка при сериализации структуры kesh:", err)
			http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
			return err
		}

		// Сохраняем код подтверждения в Redis с TTL 10 минут
		err = redisClient.Set(ctx, "jwt", keshData, 10*time.Minute).Err()
		if err != nil {
			log.Fatal("Ошибка при сохранении кода в Redis:", err)
			http.Error(rw, "Ошибка сервера", http.StatusInternalServerError)
			return err
		}

		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) EnterCodeForRecoveryPassWithEmailSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, passwd string) (err error) {
	request, err := rep.Query(
		ctx,
		`UPDATE users.users SET password_hash = $1 WHERE id = $2 RETURNING id;`,
		passwd,
		user_id)

	errorr(err)

	var id int

	for request.Next() {
		err := request.Scan(
			&id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || id != 0 {
		response := Response{
			Status:  "success",
			Data:    id,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}
