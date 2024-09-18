package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func (repo *MyRepository) RegisterOrderSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool,
	ad_id,
	renter_id,
	total_price int,
	starts_at,
	ends_at time.Time,
) (err error) {
	type Product struct {
		AdID     int       `json:"ad_id"`
		OrderID  int       `json:"order_id"`
		Amount   int       `json:"amount"`
		StartsAt time.Time `json:"starts_at"`
		EndsAt   time.Time `json:"ends_at"`
	}
	products := []Product{}

	wallet_id, err := rep.Query(
		ctx,
		"SELECT id FROM Finance.wallets WHERE user_id = $1;",

		renter_id,
	)
	var wallet_id_int int
	for wallet_id.Next() {
		err := wallet_id.Scan(&wallet_id_int)
		if err != nil {
			log.Fatal(err)
		}
	}

	errorr(err)

	user_2, err := rep.Query(
		ctx,
		"SELECT owner_id FROM Ads.ads WHERE id = $1;",

		ad_id,
	)
	var user_2_int int
	for user_2.Next() {
		err := user_2.Scan(&user_2_int)
		if err != nil {
			log.Fatal(err)
		}
	}
	errorr(err)

	var transaction_reg_and_id_int int
	transaction_reg_and_id, err := rep.Query(
		ctx,
		"INSERT INTO Finance.transactions(wallet_id, amount, user_2, typee) VALUES($1, $2, $3, 3) RETURNING id;",

		wallet_id_int,
		total_price,
		user_2_int,
	)
	for transaction_reg_and_id.Next() {
		err := transaction_reg_and_id.Scan(&transaction_reg_and_id_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	errorr(err)

	request, err := rep.Query(
		ctx,
		`
			WITH i AS (
				INSERT INTO Orders.orders(ad_id, renter_id, total_price)
				VALUES ($1, $2, $3)
				RETURNING id, ad_id
			),
			b AS (
				INSERT INTO Orders.bookings (order_id, starts_at, ends_at, amount, transaction_id)
				SELECT i.id, $4, $5, $6, $7
				FROM i
				RETURNING order_id, starts_at, ends_at, amount, transaction_id
			)
			-- Здесь мы объединяем данные из CTE i (Orders.orders) и b (Orders.bookings)
			SELECT i.ad_id, b.order_id, b.amount , b.starts_at, b.ends_at
			FROM i, b;
		`,

		ad_id,
		renter_id,
		total_price,
		starts_at,
		ends_at,
		total_price,
		transaction_reg_and_id_int,
	)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.AdID,
			&p.OrderID,
			&p.Amount,
			&p.StartsAt,
			&p.EndsAt,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, Product{AdID: p.AdID, OrderID: p.OrderID, Amount: p.Amount, StartsAt: p.StartsAt, EndsAt: p.EndsAt})
	}

	type Response struct {
		Status  string  `json:"status"`
		Data    Product `json:"data,omitempty"`
		Message string  `json:"message"`
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Объявление не найдено",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return err
	}

	response := Response{
		Status:  "success",
		Data:    products[0],
		Message: "Объявление показано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) RebookBookingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool,
	id_old_book,
	id_users int,
	starts_at,
	ends_at time.Time,
	amount int,
) (err error) {
	type Product struct {
		Id_book    int
		Id_users_2 int
		Starts_at  time.Time
		Ends_at    time.Time
		Amount     int
	}
	products := []Product{}

	var order_id_int int
	order_id, err := rep.Query(
		ctx,
		"SELECT order_id FROM Orders.bookings WHERE id = $1;",

		id_old_book,
	)

	for order_id.Next() {
		err := order_id.Scan(&order_id_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	errorr(err)

	_, err = rep.Query(
		ctx,
		"UPDATE Orders.bookings SET type = 3 WHERE id = $1",

		id_old_book,
	)
	errorr(err)

	var wallet_id_int int
	wallet_id, err := rep.Query(
		ctx,
		"SELECT id FROM Finance.wallets WHERE user_id = $1;",

		id_users,
	)
	for wallet_id.Next() {
		err := wallet_id.Scan(&wallet_id_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	errorr(err)

	var user_2_int int
	user_2, err := rep.Query(
		ctx,
		"SELECT renter_id FROM Orders.orders WHERE id = $1;",

		order_id_int,
	)
	for user_2.Next() {
		err := user_2.Scan(&user_2_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	request, err := rep.Query(
		ctx,
		`
			WITH i AS (
				INSERT INTO Finance.transactions(wallet_id, amount, user_2, typee)
				VALUES ($1, $2, $3, 3)
				RETURNING id, user_2
			),
			j AS (
				INSERT INTO Orders.bookings(order_id, starts_at, ends_at, amount, transaction_id, type)
				SELECT $4, $5, $6, $7, i.id, 2
				FROM i
				RETURNING bookings.id, bookings.starts_at, bookings.ends_at, bookings.amount
			)
			SELECT j.id, i.user_2, j.starts_at, j.ends_at, j.amount
			FROM i, j;
		`,

		wallet_id_int,
		amount,
		user_2_int,

		order_id_int,
		starts_at,
		ends_at,
		amount,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id_book,
			&p.Id_users_2,
			&p.Starts_at,
			&p.Ends_at,
			&p.Amount,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, Product{Id_book: p.Id_book, Id_users_2: p.Id_users_2, Starts_at: p.Starts_at, Ends_at: p.Ends_at, Amount: p.Amount})
	}

	fmt.Println(
		wallet_id_int,
		amount,
		user_2_int,

		order_id_int,
		starts_at,
		ends_at,
		amount,
	)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
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
		Data:    products,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) GroupOrdersByRentedSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Renter_id   int
		Total_price int
		Created_at  time.Time
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, renter_id, total_price, created_at FROM Orders.orders WHERE status = 1;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Renter_id,
			&p.Total_price,
			&p.Created_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
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
		Data:    products,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}

func (repo *MyRepository) GroupOrdersByUnRentedSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool) (err error) {
	type Product struct {
		Id          int
		Ad_id       int
		Renter_id   int
		Total_price int
		Created_at  time.Time
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		"SELECT id, ad_id, renter_id, total_price, created_at FROM Orders.orders WHERE status = 2;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, renter_id, total_price, created_at")

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Id,
			&p.Ad_id,
			&p.Renter_id,
			&p.Total_price,
			&p.Created_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
	}

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if err != nil || len(products) == 0 {
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
		Data:    products,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return
}
