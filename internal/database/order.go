package database

import (
	"context"
	"encoding/json"
	"fmt"
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

func (repo *MyRepository) RegOrderHourlySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, ads_id int, starts_at, ends_at time.Time, positionX, positionY float64) (err error) {
	type Product struct {
		Order_id             *int `json:"Order_id"`
		Updated_frozen_funds *int `json:"Updated_frozen_funds"`
		Booking_id           *int `json:"Booking_id"`
		Transaction_id       *int `json:"Transaction_id"`
	}

	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH owner AS ( -- данные хозяина объявления
			SELECT owner_id, hourly_rate FROM ads.ads WHERE id = $1
		),
		wallet AS ( -- данные арендатора
			SELECT id AS wallet_id, total_balance FROM finance.wallets WHERE user_id = $2
		), 
		transact AS (
			INSERT INTO finance.transactions (wallet_id, amount, typee, user_2)
			SELECT (SELECT wallet_id FROM wallet),
				(SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner),
				3,
				(SELECT owner_id FROM owner)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id, amount
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact),
				total_balance = total_balance - (SELECT amount FROM transact)
			WHERE user_id = $2 AND frozen_funds + (SELECT amount FROM transact) <= total_balance
			RETURNING frozen_funds
		),
		orderr AS (
			INSERT INTO orders.orders(position)
			SELECT POINT($5, $6)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		),
		booking AS (
			INSERT INTO orders.bookings(ads_id, starts_at, ends_at, transaction_id, typee, order_id)
			SELECT $1, $3, $4,
				(SELECT id FROM transact), 1, (SELECT id FROM orderr)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		)
		SELECT
			(SELECT id FROM orderr) AS order_id,
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds,
			(SELECT id FROM booking) AS booking_id,
			(SELECT id FROM transact) AS transaction_id;
		`,

		ads_id,
		user_id,
		starts_at,
		ends_at,
		positionX,
		positionY,
	)
	errorr(err)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if request == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Order_id,
			&p.Updated_frozen_funds,
			&p.Booking_id,
			&p.Transaction_id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Order_id: p.Order_id, Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) RegOrderDailySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, ads_id int, starts_at, ends_at time.Time, positionX, positionY float64) (err error) {
	type Product struct {
		Order_id             *int `json:"Order_id"`
		Updated_frozen_funds *int `json:"Updated_frozen_funds"`
		Booking_id           *int `json:"Booking_id"`
		Transaction_id       *int `json:"Transaction_id"`
	}

	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH owner AS ( -- данные хозяина объявления
			SELECT owner_id, daily_rate FROM ads.ads WHERE id = $1
		),
		wallet AS ( -- данные арендатора
			SELECT id AS wallet_id, total_balance FROM finance.wallets WHERE user_id = $2
		), 
		transact AS (
			INSERT INTO finance.transactions (wallet_id, amount, typee, user_2)
			SELECT (SELECT wallet_id FROM wallet),
				(SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner),
				3,
				(SELECT owner_id FROM owner)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id, amount
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact),
				total_balance = total_balance - (SELECT amount FROM transact)
			WHERE user_id = $2 AND frozen_funds + (SELECT amount FROM transact) <= total_balance
			RETURNING frozen_funds
		),
		orderr AS (
			INSERT INTO orders.orders(position)
			SELECT POINT($5, $6)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		),
		booking AS (
			INSERT INTO orders.bookings(ads_id, starts_at, ends_at, transaction_id, typee, order_id)
			SELECT $1,
				$3,
				$4,
				(SELECT id FROM transact), 1, (SELECT id FROM orderr)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		)
		SELECT
			(SELECT id FROM orderr) AS order_id,
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds,
			(SELECT id FROM booking) AS booking_id,
			(SELECT id FROM transact) AS transaction_id;
		`,

		ads_id,
		user_id,
		starts_at,
		ends_at,
		positionX,
		positionY,
	)
	errorr(err)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if request == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Order_id,
			&p.Updated_frozen_funds,
			&p.Booking_id,
			&p.Transaction_id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Order_id: p.Order_id, Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) BiddingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, ads_id, global_rate, user_id int, start_at, end_at time.Time, positX, positY float64) (err error) {
	// Выполняем SQL-запрос и проверяем ошибки
	request, err := rep.Query(
		ctx,
		`
		WITH owner AS (
			SELECT owner_id FROM ads.ads WHERE id = $1
		)
		INSERT INTO Finance.bidding (ads_id, renter_id, global_rate, start_at, end_at, position)
		SELECT $1, $2, $3, $4, $5, POINT($6, $7)
		WHERE (SELECT owner_id FROM owner) != $2
		RETURNING id;
		`,
		ads_id,
		user_id,
		global_rate,
		start_at,
		end_at,
		positX,
		positY)
	errorr(err)

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

func (repo *MyRepository) RegOrderWithBiddingSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, bidding_id int) (err error) {
	type Product struct {
		Order_id             *int `json:"Order_id"`
		Updated_frozen_funds *int `json:"Updated_frozen_funds"`
		Booking_id           *int `json:"Booking_id"`
		Transaction_id       *int `json:"Transaction_id"`
	}

	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH bidding AS (
			SELECT bidding.global_rate, bidding.start_at, bidding.end_at, bidding.position, bidding.renter_id, ads.id AS ads_id
			FROM finance.bidding, ads.ads WHERE ads.owner_id = $1 AND bidding.id = $2 AND ads.id = bidding.ads_id
		),
		ads_price AS (
			SELECT hourly_rate FROM ads.ads WHERE id = (SELECT ads_id FROM bidding) AND owner_id = $1
		),
		wallet AS (
			SELECT id AS wallet_id, total_balance FROM finance.wallets WHERE user_id = (SELECT renter_id FROM bidding)
		),
		transact AS (
			INSERT INTO finance.transactions (wallet_id, amount, typee, user_2)
			SELECT (SELECT wallet_id FROM wallet),
				(SELECT global_rate FROM bidding),
				4,
				$1
			WHERE (SELECT global_rate FROM bidding) <= (SELECT total_balance FROM wallet)
			RETURNING id, amount
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact),
				total_balance = total_balance - (SELECT amount FROM transact)
			WHERE user_id = (SELECT renter_id FROM bidding)
			AND frozen_funds + (SELECT amount FROM transact) <= total_balance
			RETURNING frozen_funds
		),
		orderr AS (
			INSERT INTO orders.orders(position)
			SELECT (SELECT position FROM bidding)
			WHERE (SELECT global_rate FROM bidding) <= (SELECT total_balance FROM wallet)
			RETURNING id
		),
		booking AS (
			INSERT INTO orders.bookings(ads_id, starts_at, ends_at, transaction_id, typee, order_id)
			SELECT (SELECT ads_id FROM bidding),
				(SELECT start_at FROM bidding),
				(SELECT end_at FROM bidding),
				(SELECT id FROM transact), 1, (SELECT id FROM orderr)
			WHERE (SELECT global_rate FROM bidding) <= (SELECT total_balance FROM wallet)
			RETURNING id
		)
		SELECT
			(SELECT id FROM orderr) AS order_id,
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds,
			(SELECT id FROM booking) AS booking_id,
			(SELECT id FROM transact) AS transaction_id;
		`,

		user_id,
		bidding_id)
	errorr(err)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if request == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Order_id,
			&p.Updated_frozen_funds,
			&p.Booking_id,
			&p.Transaction_id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Order_id: p.Order_id, Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) RebookOrderHourlySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, order_id int, starts_at, ends_at time.Time) (err error) {
	type Product struct {
		Order_id             int  `json:"Order_id"`
		Updated_frozen_funds *int `json:"Updated_frozen_funds"`
		Booking_id           *int `json:"Booking_id"`
		Transaction_id       *int `json:"Transaction_id"`
	}

	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH ads AS (
			SELECT bookings.ads_id
			FROM orders.bookings, ads.ads, finance.transactions, finance.wallets, users.users
			WHERE bookings.order_id = $1 AND bookings.ads_id = ads.id AND
				bookings.transaction_id = transactions.id AND transactions.wallet_id = wallets.id AND
				wallets.user_id = users.id AND users.id = $2
			LIMIT 1
		),
		owner AS ( -- данные хозяина объявления
			SELECT owner_id, hourly_rate FROM ads.ads WHERE id = (SELECT ads_id FROM ads)
		),
		wallet AS ( -- данные арендатора
			SELECT id AS wallet_id, total_balance FROM finance.wallets WHERE user_id = $2
		), 
		transact AS (
			INSERT INTO finance.transactions (wallet_id, amount, typee, user_2)
			SELECT (SELECT wallet_id FROM wallet),
				(SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner),
				3,
				(SELECT owner_id FROM owner)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id, amount
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact),
				total_balance = total_balance - (SELECT amount FROM transact)
			WHERE user_id = $2 AND frozen_funds + (SELECT amount FROM transact) <= total_balance
			RETURNING frozen_funds
		),
		booking AS (
			INSERT INTO orders.bookings(ads_id, starts_at, ends_at, transaction_id, typee, order_id)
			SELECT (SELECT ads_id FROM ads),
				$3,
				$4,
				(SELECT id FROM transact), 2, $1
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT hourly_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		)
		SELECT
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds,
			(SELECT id FROM booking) AS booking_id,
			(SELECT id FROM transact) AS transaction_id;
		`,

		order_id,
		user_id,
		starts_at,
		ends_at,
	)
	errorr(err)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if request == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Updated_frozen_funds,
			&p.Booking_id,
			&p.Transaction_id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Order_id: order_id, Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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

func (repo *MyRepository) RebookOrderDailySQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, user_id, order_id int, starts_at, ends_at time.Time) (err error) {
	type Product struct {
		Order_id             int  `json:"Order_id"`
		Updated_frozen_funds *int `json:"Updated_frozen_funds"`
		Booking_id           *int `json:"Booking_id"`
		Transaction_id       *int `json:"Transaction_id"`
	}

	products := []Product{}

	request, err := rep.Query(
		ctx,
		`
		WITH ads AS (
			SELECT bookings.ads_id
			FROM orders.bookings, ads.ads, finance.transactions, finance.wallets, users.users
			WHERE bookings.order_id = $1 AND bookings.ads_id = ads.id AND
				bookings.transaction_id = transactions.id AND transactions.wallet_id = wallets.id AND
				wallets.user_id = users.id AND users.id = $2
			LIMIT 1
		),
		owner AS ( -- данные хозяина объявления
			SELECT owner_id, daily_rate FROM ads.ads WHERE id = (SELECT ads_id FROM ads)
		),
		wallet AS ( -- данные арендатора
			SELECT id AS wallet_id, total_balance FROM finance.wallets WHERE user_id = $2
		), 
		transact AS (
			INSERT INTO finance.transactions (wallet_id, amount, typee, user_2)
			SELECT (SELECT wallet_id FROM wallet),
				(SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner),
				3,
				(SELECT owner_id FROM owner)
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id, amount
		),
		wallet_update AS (
			UPDATE finance.wallets
			SET frozen_funds = frozen_funds + (SELECT amount FROM transact),
				total_balance = total_balance - (SELECT amount FROM transact)
			WHERE user_id = $2 AND frozen_funds + (SELECT amount FROM transact) <= total_balance
			RETURNING frozen_funds
		),
		booking AS (
			INSERT INTO orders.bookings(ads_id, starts_at, ends_at, transaction_id, typee, order_id)
			SELECT (SELECT ads_id FROM ads),
				$3,
				$4,
				(SELECT id FROM transact), 2, $1
			WHERE (SELECT $4::timestamp::date - $3::timestamp::date)
				* (SELECT daily_rate FROM owner) <= (SELECT total_balance FROM wallet)
			RETURNING id
		)
		SELECT
			(SELECT frozen_funds FROM wallet_update) AS updated_frozen_funds,
			(SELECT id FROM booking) AS booking_id,
			(SELECT id FROM transact) AS transaction_id;
		`,

		order_id,
		user_id,
		starts_at,
		ends_at,
	)
	errorr(err)

	type Response struct {
		Status  string    `json:"status"`
		Data    []Product `json:"data,omitempty"`
		Message string    `json:"message"`
	}

	if request == nil {
		response := Response{
			Status:  "fatal",
			Message: "Не прошла",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)
		return err
	}

	for request.Next() {
		p := Product{}
		err := request.Scan(
			&p.Updated_frozen_funds,
			&p.Booking_id,
			&p.Transaction_id)
		if err != nil {
			fmt.Println(err)
			continue
		}

		products = append(products, Product{Order_id: order_id, Transaction_id: p.Transaction_id, Booking_id: p.Booking_id, Updated_frozen_funds: p.Updated_frozen_funds})
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
