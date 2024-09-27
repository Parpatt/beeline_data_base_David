package database

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
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

type FileUploadRequest struct {
	Filename string `json:"filename"`
	Filetype string `json:"filetype"`
	Data     []byte `json:"data"`
}

func errorr(err error) {
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
}

func addImage(r *http.Request, rw http.ResponseWriter, Filename, Filetype, Data string) {
	// Декодируем Base64-строку в байты
	imgData, err := base64.StdEncoding.DecodeString(Data)
	errorr(err)

	// Создаём multipart/form-data запрос
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Создаем часть для передачи файла
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		panic("Не удалось создать часть для файла: " + err.Error())
	}

	// Записываем байты файла в запрос
	_, err = part.Write(imgData)
	if err != nil {
		panic("Не удалось записать файл: " + err.Error())
	}

	// Закрываем writer для завершения формирования запроса
	writer.Close()

	// Создаем HTTP POST запрос
	req, err := http.NewRequest("POST", "http://176.124.192.39:8080/upload", &body)
	errorr(err)

	// Устанавливаем заголовок Content-Type для multipart/form-data
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Отправляем запрос на сервер с использованием HTTP клиента
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic("Ошибка при отправке запроса: " + err.Error())
	}
	defer resp.Body.Close()

	// Читаем ответ от сервера
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		panic("Не удалось прочитать ответ от сервера: " + err.Error())
	}

	// Выводим ответ от сервера
	fmt.Println("Ответ от сервера:", string(respBody))
}

func (repo *MyRepository) AddNewLegalUserSQL(
	ctx context.Context,
	rep *pgxpool.Pool,
	rw http.ResponseWriter,
	r *http.Request,
	ind_num_taxp int,
	name_of_company,
	address_name,
	email,
	phoneNum,
	hashedPassword,
	Filename,
	Filetype,
	Data string) (err error) {
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
		addImage(r, rw, Filename, Filetype, Data)

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

func (repo *MyRepository) AddNewNaturUserSQL(
	ctx context.Context,
	rep *pgxpool.Pool,
	rw http.ResponseWriter,
	r *http.Request,
	name,
	surname,
	patronymic,
	email,
	phoneNum,
	hashedPassword,
	Filename,
	Filetype,
	Data string) (err error) {
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

	type Response struct {
		Status  string    `json:"status"`
		Data    NaturUser `json:"data,omitempty"`
		Message string    `json:"message"`
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
		addImage(r, rw, Filename, Filetype, Data)

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
			Status:  "fatal",
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

func (repo *MyRepository) EditingNaturUserDataSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, avatar_path string, name string, surname string, patronymic string) (err error) {
	type NaturUser struct {
		Id_1 int `json:"id_1" db:"id_1"`
		Id_2 int `json:"id_2" db:"id_2"`

		Avatar_path string `json:"avatar_path" db:"avatar_path"`
		Name        string `json:"name" db:"name"`
		Surname     string `json:"surname" db:"surname"`
		Patronymic  string `json:"patronymic" db:"patronymic"`
	}
	products := []NaturUser{}

	request, err := rep.Query(
		ctx,
		`WITH i AS (
    		SELECT id, avatar_path FROM users.users WHERE id = $1
		),
		j AS (
			SELECT name, surname, patronymic FROM users.individual_user WHERE user_id = $1
		)
		SELECT i.id, i.avatar_path, j.name, j.surname, j.patronymic
		FROM i, j;`, user_id)

	errorr(err)

	var id int

	for request.Next() {
		p := NaturUser{}
		err := request.Scan(
			&id,
			&p.Avatar_path,
			&p.Name,
			&p.Surname,
			&p.Patronymic,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	if avatar_path == "" {
		avatar_path = products[0].Avatar_path
	}
	if name == "" {
		name = products[0].Name
	}
	if surname == "" {
		surname = products[0].Surname
	}
	if patronymic == "" {
		patronymic = products[0].Patronymic
	}

	request, err = rep.Query(
		ctx,
		`
			WITH i AS (
				UPDATE users.users SET avatar_path = $1 WHERE id = $2 RETURNING id
			),
			j AS (
				UPDATE users.individual_user SET name = $3, surname = $4, patronymic = $5 WHERE user_id = $6 RETURNING user_id
			)
			SELECT i.id AS user_id, j.user_id AS individual_user_id
			FROM i, j;
		`,
		avatar_path,
		user_id,
		name,
		surname,
		patronymic,
		user_id)

	var user_idd int
	var individual_user_id int

	for request.Next() {
		err := request.Scan(
			&user_idd,
			&individual_user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	products[0].Id_1 = user_idd
	products[0].Id_2 = individual_user_id

	type Response struct {
		Status  string      `json:"status"`
		Data    []NaturUser `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	if err == nil || id != 0 || user_idd != 0 || individual_user_id != 0 {
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

func (repo *MyRepository) EditingLegalUserDataSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, avatar_path string, ind_num_taxp int, name_of_company string, address_name string) (err error) {
	type LegalUser struct {
		Id_1 int `json:"id_1" db:"id_1"`
		Id_2 int `json:"id_2" db:"id_2"`

		Avatar_path     string `json:"avatar_path" db:"avatar_path"`
		Ind_num_taxp    int    `json:"ind_num_taxp" db:"ind_num_taxp"`
		Name_of_company string `json:"name_of_company" db:"name_of_company"`
		Address_name    string `json:"address_name" db:"address_name"`
	}
	products := []LegalUser{}

	request, err := rep.Query(
		ctx,
		`WITH i AS (
    		SELECT id, avatar_path FROM users.users WHERE id = $1
		),
		j AS (
			SELECT ind_num_taxp, name_of_company, address_name FROM users.company_user WHERE user_id = $1
		)
		SELECT i.id, i.avatar_path, j.ind_num_taxp, j.name_of_company, j.address_name
		FROM i, j;`, user_id)

	errorr(err)

	var id int

	for request.Next() {
		p := LegalUser{}
		err := request.Scan(
			&id,
			&p.Avatar_path,
			&p.Ind_num_taxp,
			&p.Name_of_company,
			&p.Address_name,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, p)
	}

	if avatar_path == "" {
		avatar_path = products[0].Avatar_path
	}
	if ind_num_taxp == 0 {
		ind_num_taxp = products[0].Ind_num_taxp
	}
	if name_of_company == "" {
		name_of_company = products[0].Name_of_company
	}
	if address_name == "" {
		address_name = products[0].Address_name
	}

	request, err = rep.Query(
		ctx,
		`
			WITH i AS (
				UPDATE users.users SET avatar_path = $1 WHERE id = $2 RETURNING id
			),
			j AS (
				UPDATE users.company_user SET ind_num_taxp = $3, name_of_company = $4, address_name = $5 WHERE user_id = $6 RETURNING user_id
			)
			SELECT i.id AS user_id, j.user_id AS individual_user_id
			FROM i, j;
		`,
		avatar_path,
		user_id,
		ind_num_taxp,
		name_of_company,
		address_name,
		user_id)

	var user_idd int
	var individual_user_id int

	for request.Next() {
		err := request.Scan(
			&user_idd,
			&individual_user_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}

	products[0].Id_1 = user_idd
	products[0].Id_2 = individual_user_id

	type Response struct {
		Status  string      `json:"status"`
		Data    []LegalUser `json:"data,omitempty"`
		Message string      `json:"message"`
	}

	if err == nil || id != 0 || user_idd != 0 || individual_user_id != 0 {
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
