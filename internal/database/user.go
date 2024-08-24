package database

import (
	"context"
	"encoding/json"
	"fmt"
	"myproject/internal/models"
	"net/http"
	"time"

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
			Status: "fatal",
			Data: NaturUser{
				Id:            0,
				Password_hash: "",
				Email:         "",
				Phone_number:  "",
				Updated_at:    time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
				Avatar_path:   "",
				User_type:     0,
				User_role:     0,

				Name:       "",
				Surname:    "",
				Patronymic: "",
			},
			Message: "Введите корректный & уникальный логин и пароль ",
		})

		return
	} else if errors != nil {
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(Response{
			Status: "fatal",
			Data: NaturUser{
				Id:            0,
				Password_hash: "",
				Email:         "",
				Phone_number:  "",
				Updated_at:    time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC),
				Avatar_path:   "",
				User_type:     0,
				User_role:     0,

				Name:       "",
				Surname:    "",
				Patronymic: "",
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

func (repo *MyRepository) LoginSQL(ctx context.Context, login, hashedPassword string, rep *pgxpool.Pool, rw http.ResponseWriter) (u models.User, err error, id int) {
	row := rep.QueryRow(ctx, `SELECT id, email FROM Users.users WHERE (email = $1 AND password_hash = $2) OR (phone_number = $1 AND password_hash = $2);`, login, hashedPassword)
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
