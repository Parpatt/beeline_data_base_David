package database

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v4/pgxpool"
)

func (repo *MyRepository) TransactionToAnotherSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, user_2 int, amount int) (err error) {
	your_wallet_id, err := rep.Query(ctx, "SELECT id FROM Finance.wallets WHERE user_id = $1;", user_id)
	errorr(err)

	var your_wallet_id_int int //дата создания репорта
	for your_wallet_id.Next() {
		err := your_wallet_id.Scan(&your_wallet_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	user_2_wallet_id, err := rep.Query(ctx, "SELECT id FROM Finance.wallets WHERE user_id = $1;", user_2)
	errorr(err)
	var user_2_wallet_id_int int //дата создания репорта
	for user_2_wallet_id.Next() {
		err := user_2_wallet_id.Scan(&user_2_wallet_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	transact_id, err := rep.Query(
		ctx,
		`
		WITH i AS (
			INSERT INTO Finance.transactions (wallet_id, amount, user_2, typee)
			VALUES ($1, $2, $3, $4)
			RETURNING id, amount
		)
		INSERT INTO Finance.transactions (wallet_id, amount, user_2, typee)
		SELECT $5, i.amount, $6, $7
		FROM i
		RETURNING id;
		`,

		your_wallet_id_int,
		amount,
		user_2,
		1,

		user_2_wallet_id_int,
		user_id,
		2,
	)

	var transact_id_int int
	for transact_id.Next() {
		err := transact_id.Scan(&transact_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || transact_id_int == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	response := Response{
		Status:  "success",
		Data:    transact_id_int,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) TransactionToSomethingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, amount int) (err error) {
	wallet_id, err := rep.Query(ctx, "SELECT id FROM Finance.wallets WHERE user_id = $1;", user_id)
	errorr(err)
	var wallet_id_int int
	for wallet_id.Next() {
		err := wallet_id.Scan(&wallet_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	transact_id, err := rep.Query(
		ctx,
		`INSERT INTO Finance.transactions (wallet_id, amount, typee) 
			VALUES ($1, $2, 3) 
			RETURNING id;`,

		wallet_id_int,
		amount,
	)
	errorr(err)

	var transact_id_int int //дата создания репорта
	for transact_id.Next() {
		err := transact_id.Scan(&transact_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || transact_id_int == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	response := Response{
		Status:  "success",
		Data:    transact_id_int,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return
}

func (repo *MyRepository) TransactionToReturnAmountSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id int, amount int) (err error) {
	wallet_id, err := rep.Query(ctx, "SELECT id FROM Finance.wallets WHERE user_id = $1;", user_id)
	errorr(err)
	var wallet_id_int int //дата создания репорта
	for wallet_id.Next() {
		err := wallet_id.Scan(&wallet_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	transact_id, err := rep.Query(
		ctx,
		`INSERT INTO Finance.transactions (wallet_id, amount, typee)
		 VALUES ($1, $2, 3)
		 RETURNING id;`,

		wallet_id_int,
		amount,
	)
	errorr(err)

	var transact_id_int int //дата создания репорта
	for transact_id.Next() {
		err := transact_id.Scan(&transact_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err != nil || transact_id_int == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}
	response := Response{
		Status:  "success",
		Data:    transact_id_int,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return
}
