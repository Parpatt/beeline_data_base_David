package repository

import (
	"context"
	"fmt"

	"log"
	"net/http"
	"time"
)

func (r *Repository) RegisterRentedSQL(ctx context.Context, rw http.ResponseWriter,
	ad_id,
	renter_id,
	total_price int,
	starts_at,
	ends_at time.Time,
) (err error) {
	type product struct {
		ad_id       int
		renter_id   int
		total_price int

		starts_at time.Time
		ends_at   time.Time
	}
	products := []product{}

	wallet_id, err := r.pool.Query(
		ctx,
		"SELECT id FROM Finance.wallets WHERE user_id = $1;",

		renter_id,
	)
	var wallet_id_int int
	for wallet_id.Next() {
		err := wallet_id.Scan(&wallet_id_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	user_2, err := r.pool.Query(
		ctx,
		"SELECT owner_id FROM Ads.ads WHERE id = $1;",

		ad_id,
	)
	var user_2_int int
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

	var transaction_reg_and_id_int int
	transaction_reg_and_id, err := r.pool.Query(
		ctx,
		"WITH i AS (INSERT INTO Finance.transactions(wallet_id, amount, user_2, typee) VALUES($1, $2, $3, $4) RETURNING id) SELECT i.id FROM i",

		wallet_id_int,
		total_price,
		user_2_int,
		3,
	)
	for transaction_reg_and_id.Next() {
		err := transaction_reg_and_id.Scan(&transaction_reg_and_id_int) // Сканируем значение в переменную id
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	request, err := r.pool.Query(
		ctx,
		"WITH i AS (INSERT INTO Orders.orders(ad_id, renter_id, total_price) VALUES ($1, $2, $3) RETURNING id) INSERT INTO Orders.bookings (order_id, starts_at, ends_at, amount, transaction_id) SELECT i.id, $4, $5, $6, $7 FROM i;",

		ad_id,
		renter_id,
		total_price,
		starts_at,
		ends_at,
		total_price,
		transaction_reg_and_id_int,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.ad_id,
			&p.renter_id,
			&p.total_price,
			&p.starts_at,
			&p.ends_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) RebookRentedGET(ctx context.Context, rw http.ResponseWriter,
	id_old_book,
	id_users int,
	starts_at,
	ends_at time.Time,
	amount int,
) (err error) {
	type product struct {
		id_old_book int
		id_users    int
		starts_at   time.Time
		ends_at     time.Time
		amount      int
	}
	products := []product{}

	var order_id_int int
	order_id, err := r.pool.Query(
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
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
	fmt.Fprint(rw, "order_id: ")
	fmt.Fprintln(rw, order_id_int)

	_, err = r.pool.Query(
		ctx,
		"UPDATE Orders.bookings SET type = 3 WHERE id = $1",

		id_old_book,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	var wallet_id_int int
	wallet_id, err := r.pool.Query(
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
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}
	fmt.Fprint(rw, "wallet_id: ")
	fmt.Fprintln(rw, wallet_id_int)

	var user_2_int int
	user_2, err := r.pool.Query(
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
	fmt.Fprint(rw, "user_2: ")
	fmt.Fprintln(rw, user_2_int)

	request, err := r.pool.Query(
		ctx,
		"WITH i AS(INSERT INTO Finance.transactions(wallet_id, amount, user_2, typee) VALUES($1, $2, $3, 3) RETURNING id) INSERT INTO Orders.bookings(order_id, starts_at, ends_at, amount, transaction_id, type) SELECT $4, $5, $6, $7, i.id, 2 FROM i;",

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
		p := product{}
		err := request.Scan(
			&p.id_old_book,
			&p.id_users,
			&p.starts_at,
			&p.ends_at,
			&p.amount,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupOrdersByRentedSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		renter_id   int
		total_price int
		created_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, renter_id, total_price, created_at FROM Orders.orders WHERE status = 1;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, renter_id, total_price, created_at")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.renter_id,
			&p.total_price,
			&p.created_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}

func (r *Repository) GroupOrdersByUnRentedSQL(ctx context.Context, rw http.ResponseWriter) (err error) {
	type product struct {
		id          int
		ad_id       int
		renter_id   int
		total_price int
		created_at  time.Time
	}
	products := []product{}

	request, err := r.pool.Query(
		ctx,
		"SELECT id, ad_id, renter_id, total_price, created_at FROM Orders.orders WHERE status = 2;",
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	fmt.Fprintln(rw, "id, ad_id, renter_id, total_price, created_at")

	for request.Next() {
		p := product{}
		err := request.Scan(
			&p.id,
			&p.ad_id,
			&p.renter_id,
			&p.total_price,
			&p.created_at,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, p)
		fmt.Fprintln(rw, p)
	}

	return
}
