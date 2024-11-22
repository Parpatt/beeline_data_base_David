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

func dateInterpreter(starts_at, ends_at time.Time) (int, float64) {
	//интервал
	duration := ends_at.Sub(starts_at)

	// Получаем количество дней (целое число)
	days := int(duration.Hours() / 24)

	// Получаем количество часов (оставшиеся часы после вычитания дней)
	hours := duration.Hours() - float64(days*24)

	return days, hours
}

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

func (repo *MyRepository) RegBookingHourlySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, ads_id int, starts_at, ends_at time.Time) (err error) {
	days, hours := dateInterpreter(starts_at, ends_at)

	type Product struct {
		Transaction_id       int
		Booking_id           int
		Updated_frozen_funds float64
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
			WITH ads AS (
				SELECT ad_id FROM orders.orders WHERE id = $1
			),
			owner AS (
				SELECT owner_id, hourly_rate, daily_rate FROM ads.ads WHERE id = (SELECT ad_id FROM ads)
			),
			wallet AS (
				SELECT id AS wallet_id FROM finance.wallets WHERE user_id = $2
			), 
			transact AS (
				INSERT INTO finance.transactions(wallet_id, amount, typee, user_2) 
				VALUES ((SELECT wallet_id FROM wallet),
					$3 * (SELECT daily_rate FROM owner) + 
					$4 * (SELECT hourly_rate FROM owner), 3, (SELECT owner_id FROM owner)) 
				RETURNING id, amount
			), 
			booking AS (
				INSERT INTO orders.bookings(order_id, starts_at, ends_at, amount, transaction_id) 
				VALUES ($5, $6, $7, (SELECT amount FROM transact), (SELECT id FROM transact)) 
				RETURNING id
			),
			wallet_update AS (
				UPDATE finance.wallets
				SET frozen_funds = frozen_funds + (SELECT amount FROM transact)
				WHERE user_id = (SELECT owner_id FROM owner)
				RETURNING frozen_funds
			)
			SELECT 
				(SELECT id FROM transact) AS transaction_id,
				(SELECT id FROM booking) AS booking_id,
				(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds;
		`,

		//order_id,
		user_id,
		days,
		hours,
		//order_id,
		starts_at,
		ends_at)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Transaction_id,
			&p.Booking_id,
			&p.Updated_frozen_funds)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) BiddingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, chat_id, global_rate int) (err error) {
	// Выполняем SQL-запрос и проверяем ошибки
	request, err := rep.Query(
		ctx,
		`
		INSERT INTO Finance.bidding (chat_id, global_rate, type)
		VALUES ($1, $2, 1)
		RETURNING id;
		`,
		chat_id,
		global_rate,
	)
	if err != nil {
		errorr(err)
		return err
	}
	defer request.Close() // Закрываем запрос в конце выполнения функции

	// Переменная для хранения возвращаемого id
	var bidding_id int
	if request.Next() {
		// Используем &bidding_id, так как Scan требует указатель
		err := request.Scan(&bidding_id)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	// Структура ответа
	type Response struct {
		Status  string `json:"status"`
		Data    int    `json:"data,omitempty"`
		Message string `json:"message"`
	}

	// Проверяем, был ли id успешным
	if err != nil || bidding_id == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	// Успешный ответ с данными
	response := Response{
		Status:  "success",
		Data:    bidding_id,
		Message: "Транзакция прошла успешно",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)
	return nil
}

func (repo *MyRepository) RegBookingWithBiddingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, order_id, bidding_id int, starts_at, ends_at time.Time) (err error) {
	type Product struct {
		Transaction_id       int
		Booking_id           int
		Updated_frozen_funds float64
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH ads AS (
			SELECT ad_id FROM orders.orders WHERE id = $1
		),
		owner AS (
			SELECT owner_id FROM ads.ads WHERE id = (SELECT ad_id FROM ads)
		),
		wallet AS (
			SELECT id AS wallet_id FROM finance.wallets WHERE user_id = $2
		),
		amount AS (
			SELECT global_rate FROM finance.bidding WHERE id = $3
		), 
		transact AS (
			INSERT INTO finance.transactions(wallet_id, amount, typee, user_2) 
			VALUES ((SELECT wallet_id FROM wallet), (SELECT global_rate FROM amount), 3, (SELECT owner_id FROM owner)) 
			RETURNING id, amount
		), 
		booking AS (
			INSERT INTO orders.bookings(order_id, starts_at, ends_at, amount, transaction_id) 
			VALUES ($4, $5, $6, (SELECT amount FROM transact), (SELECT id FROM transact)) 
			RETURNING id
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact)
			WHERE user_id = (SELECT owner_id FROM owner)
			RETURNING frozen_funds
		)
		SELECT 
			(SELECT id FROM transact) AS transaction_id,
			(SELECT id FROM booking) AS booking_id,
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds;
		`,

		order_id,
		user_id,
		bidding_id,
		order_id,
		starts_at,
		ends_at)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Transaction_id,
			&p.Booking_id,
			&p.Updated_frozen_funds)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) RebookBookingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, order_id int, starts_at, ends_at time.Time) (err error) {
	days, hours := dateInterpreter(starts_at, ends_at)

	type Product struct {
		Transaction_id       int
		Booking_id           int
		Updated_frozen_funds float64
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
			WITH ads AS (
				SELECT ad_id FROM orders.orders WHERE id = $1
			),
			owner AS (
				SELECT owner_id, hourly_rate, daily_rate FROM ads.ads WHERE id = (SELECT ad_id FROM ads)
			),
			wallet AS (
				SELECT id AS wallet_id FROM finance.wallets WHERE user_id = $2
			), 
			transact AS (
				INSERT INTO finance.transactions(wallet_id, amount, typee, user_2) 
				VALUES ((SELECT wallet_id FROM wallet),
					$3 * (SELECT daily_rate FROM owner) + 
					$4 * (SELECT hourly_rate FROM owner), 3, (SELECT owner_id FROM owner)) 
				RETURNING id, amount
			), 
			booking AS (
				INSERT INTO orders.bookings(order_id, starts_at, ends_at, amount, transaction_id) 
				VALUES ($5, $6, $7, (SELECT amount FROM transact), (SELECT id FROM transact)) 
				RETURNING id
			),
			wallet_update AS (
				UPDATE finance.wallets
				SET frozen_funds = frozen_funds + (SELECT amount FROM transact)
				WHERE user_id = (SELECT owner_id FROM owner)
				RETURNING frozen_funds
			)
			SELECT 
				(SELECT id FROM transact) AS transaction_id,
				(SELECT id FROM booking) AS booking_id,
				(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds;
		`,

		order_id,
		user_id,
		days,
		hours,
		order_id,
		starts_at,
		ends_at)
	errorr(err)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Transaction_id,
			&p.Booking_id,
			&p.Updated_frozen_funds)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) SucBookingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, order_id []int, id_user int) (err error) {
	type Product struct {
		Amount  float64
		Type    int
		User_id int
	}
	products := []Product{}

	for i := range len(order_id) {
		request, err := rep.Query(
			ctx,
			`
			WITH booking AS (
				SELECT amount, transaction_id FROM Orders.bookings WHERE id = 80 AND type != 3
			),
			del_booking AS (
				UPDATE Orders.bookings
				SET type = 3
				WHERE id = 80 AND type != 3
				RETURNING type
			),
			wallet AS (
				SELECT user_2 FROM Finance.transactions
				WHERE id = (SELECT transaction_id FROM booking)
			),
			calculation_money AS (
				UPDATE Finance.wallets
				SET total_balance = total_balance + (SELECT amount FROM booking),
				frozen_funds = frozen_funds + (SELECT amount FROM booking)
				WHERE user_id = (SELECT user_2 FROM wallet) AND user_id = 29
				RETURNING user_id
			)
			SELECT
				(SELECT amount FROM booking),
				(SELECT type FROM del_booking) AS deleted_type,
				(SELECT user_id FROM calculation_money);
		`,

			order_id[i],
			id_user)
		errorr(err)

		for request.Next() {
			p := Product{}
			err := request.Scan(
				&p.Amount,
				&p.Type,
				&p.User_id)
			if err != nil {
				fmt.Println(err)
				continue
			}

			products = append(products, Product{Amount: p.Amount, Type: p.Type, User_id: p.User_id})
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
	}
	return
}

func (repo *MyRepository) BookingListSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_users int) (err error) {
	type Product struct {
		Booking_id int
	}
	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH wallet AS (
			SELECT id AS wallet_id FROM finance.wallets WHERE user_id = $1
		),
		transact AS (
			SELECT id AS transact_id, wallet_id FROM finance.transactions
		),
		bookings AS (
			SELECT id AS booking_id, transaction_id FROM orders.bookings
		)
		SELECT bookings.booking_id
		FROM wallet
		JOIN transact ON wallet.wallet_id = transact.wallet_id
		JOIN bookings ON bookings.transaction_id = transact.transact_id;
		`,

		id_users,
	)

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Booking_id,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		products = append(products, Product{Booking_id: p.Booking_id})
	}

	type Response struct {
		Status  string
		Data    []Product
		Message string
	}

	if err != nil || len(products) == 0 {
		response := Response{
			Status:  "fatal",
			Message: "Операция не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Бронирования показаны",
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
