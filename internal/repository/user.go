package repository

import (
	"context"
	"fmt"
	"time"
)

type User struct {
	Id             int    `json:"id" db:"id"`
	Name           string `json:"name" db:"name"`
	Surname        string `json:"surname" db:"surname"`
	Patronymic     string `json:"patronymic" db:"patronymic"`
	Email          string `json:"email" db:"email"`
	PhoneNum       string `json:"phoneNum" db:"phoneNum"`
	HashedPassword string `json:"hashed_password" db:"hashed_password"`
}

func (r *Repository) Login(ctx context.Context, login, hashedPassword string) (u User, err error) {
	row := r.pool.QueryRow(ctx, `SELECT id, email FROM Users.users WHERE (email = $1 AND password_hash = $2) OR (phone_number = $1 AND password_hash = $2);`, login, hashedPassword)
	if err != nil {
		err = fmt.Errorf("failed to query data: %w", err)
		return
	}

	err = row.Scan(&u.Id, &u.Email)
	if err != nil {
		err = fmt.Errorf("failed to query data: %w", err)
		return
	}

	return
}

func (r *Repository) AddNewNaturUser(ctx context.Context, name, surname, patronymic, email, phoneNum, hashedPassword string) (err error) {
	_, err = r.pool.Exec(ctx, `WITH i AS (INSERT INTO Users.users (user_type, password_hash, email, phone_number, created_at, updated_at, avatar_path) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id) INSERT INTO Users.individual_user (user_id, name, surname, patronymic)  SELECT i.id, $8, $9, $10  FROM i;`,
		1,
		hashedPassword,
		email,
		phoneNum,
		time.Date(2017, 7, 12, 4, 56, 12, 0, time.FixedZone("null", 0)),
		time.Date(2017, 7, 12, 4, 56, 12, 0, time.FixedZone("null", 0)),
		"C:/",

		name,
		surname,
		patronymic)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}

func (r *Repository) AddNewLegalUser(ctx context.Context, ind_num_taxp, name_of_company, address_name, email, phoneNum, hashedPassword string) (err error) {
	_, err = r.pool.Exec(ctx, `WITH i AS (INSERT INTO Users.users (user_type, password_hash, email, phone_number, created_at, updated_at, avatar_path) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id) INSERT INTO Users.company_user (user_id, ind_num_taxp, name_of_company, address_name)  SELECT i.id, $8, $9, $10  FROM i;`,
		1,
		hashedPassword,
		email,
		phoneNum,
		time.Date(2017, 7, 12, 4, 56, 12, 0, time.FixedZone("null", 0)),
		time.Date(2017, 7, 12, 4, 56, 12, 0, time.FixedZone("null", 0)),
		"C:/",

		ind_num_taxp,
		name_of_company,
		address_name,
	)

	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	return
}
