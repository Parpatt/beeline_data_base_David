package database

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

func (repo *MyRepository) ChatButtonInAdsSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_user int, id_ads int) (err error) {
	request, err := rep.Query(
		ctx,
		"SELECT chat.reg_or_open_chat($1, $2);",

		id_user,
		id_ads,
	)
	errorr(err)

	var chat_id int
	for request.Next() {
		err := request.Scan(
			&chat_id,
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

	if err == nil && request != nil && chat_id != 0 {
		response := Response{
			Status:  "success",
			Data:    chat_id,
			Message: "Показано",
		}

		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Не показано, или не зарегистрировано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return err
}

func (repo *MyRepository) SigChatSQL(ctx context.Context, rw http.ResponseWriter, id_user int, id_ads int, rep *pgxpool.Pool) (err error) {
	request, err := rep.Query(
		ctx,
		"SELECT ads.owner_id FROM ads.ads WHERE id = $1;",

		id_ads,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	var id_buddy int
	for request.Next() {
		err := request.Scan(
			&id_buddy,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	request, err = rep.Query(
		ctx,
		"SELECT Chat.add_chat($1, $2, $3);",

		id_user,
		id_buddy,
		id_ads,
	)
	if err != nil {
		err = fmt.Errorf("failed to exec data: %w", err)
		return
	}

	var chat_id int
	for request.Next() {
		err := request.Scan(
			&chat_id,
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

	if err == nil && request != nil && chat_id != 0 {
		response := Response{
			Status:  "success",
			Data:    chat_id,
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

func (repo *MyRepository) OpenChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, user_id int) (err error) {
	type Product_user struct {
		User_id   int
		Name      string
		Text      string
		Media     []string
		Date      time.Time
		Media_pwd []string
	}
	Products_user_mass := []Product_user{}

	// request_1, err := rep.Query( //это запрос на вывод наших сообщений
	// 	ctx,
	// 	`
	// 	WITH i AS (
	// 		SELECT id, text, sent_at, sender_id
	// 		FROM chat.messages
	// 		WHERE chat_id = $1 AND sender_id != $2
	// 	),
	// 	j AS (
	// 		SELECT message_id, path_to_file
	// 		FROM chat.attachments
	// 		WHERE message_id IN (SELECT id FROM i)
	// 	),
	// 	company_user AS (
	// 		SELECT user_id, name_of_company
	// 		FROM users.company_user
	// 		WHERE user_id = (SELECT sender_id FROM i LIMIT 1)
	// 	),
	// 	individual_user AS (
	// 		SELECT user_id, name
	// 		FROM users.individual_user
	// 		WHERE user_id = (SELECT sender_id FROM i LIMIT 1)
	// 	)
	// 	SELECT i.sender_id AS user_id,
	// 		COALESCE(individual_user.name::TEXT, company_user.name_of_company::TEXT) AS name,
	// 		i.text, i.sent_at, j.path_to_file
	// 	FROM i
	// 	LEFT JOIN j ON j.message_id = i.id
	// 	LEFT JOIN company_user ON company_user.user_id = i.sender_id
	// 	LEFT JOIN individual_user ON individual_user.user_id = i.sender_id;
	// 	`,

	// 	id_chat,
	// 	user_id,
	// )
	// errorr(err)

	var User_iddd int
	var Namee string
	var Text string
	var Date time.Time
	var Media_pwd []string

	// for request_1.Next() {
	// 	err := request_1.Scan(
	// 		&User_iddd,
	// 		&Namee,
	// 		&Text,
	// 		&Date,
	// 		&Media_pwd,
	// 	)
	// 	if err != nil {
	// 		fmt.Println(err)

	// 		continue
	// 	}

	// 	Products_user_mass = append(Products_user_mass, Product_user{User_id: User_iddd, Name: Namee, Text: Text, Date: Date, Media_pwd: Media_pwd})
	// }

	request_2, err := rep.Query( //это запрос на вывод сообщений нашего кента
		ctx,
		`
		WITH i AS (
			SELECT id, text, sent_at, sender_id
			FROM chat.messages 
			WHERE chat_id = $1
		),
		j AS (
			SELECT message_id, path_to_file 
			FROM chat.attachments 
			WHERE message_id IN (SELECT id FROM i)
		),
		company_user AS (
			SELECT user_id, name_of_company 
			FROM users.company_user 
			WHERE user_id = (SELECT sender_id FROM i LIMIT 1)
		),
		individual_user AS (
			SELECT user_id, name 
			FROM users.individual_user 
			WHERE user_id = (SELECT sender_id FROM i LIMIT 1)
		)
		SELECT i.sender_id AS user_id, 
			COALESCE(individual_user.name, 'company_user.name_of_company') AS name,
			i.text, i.sent_at, j.path_to_file
		FROM i
		LEFT JOIN j ON j.message_id = i.id
		LEFT JOIN company_user ON company_user.user_id = i.sender_id
		LEFT JOIN individual_user ON individual_user.user_id = i.sender_id;
		`,

		id_chat,
	)
	errorr(err)

	for request_2.Next() {
		err := request_2.Scan(
			&User_iddd,
			&Namee,
			&Text,
			&Date,
			&Media_pwd,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}

		Products_user_mass = append(Products_user_mass, Product_user{User_id: User_iddd, Name: Namee, Text: Text, Date: Date, Media_pwd: Media_pwd})
	}

	type Response struct {
		Status  string         `json:"status"`
		Data    []Product_user `json:"data,omitempty"`
		Message string         `json:"message"`
	}

	if err == nil && (Products_user_mass != nil) {
		for i := 0; i < len(Products_user_mass); i++ {
			for j := 0; j < len(Products_user_mass[i].Media_pwd); j++ {
				media, err := DownloadFile(Products_user_mass[i].Media_pwd[j])
				Products_user_mass[i].Media_pwd[j] = media

				errorr(err)
			}
		}

		response := Response{
			Status:  "success",
			Data:    Products_user_mass,
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

func Notofication(rep *pgxpool.Pool, ctx context.Context, rw http.ResponseWriter, id_chat, id_user int, text string, date time.Time) {
	request, err := rep.Query(
		ctx,
		`
			WITH i AS (
				SELECT user_1_id FROM Chat.chats WHERE id = $1 AND user_1_id != $2 AND user_2_id = $2
			),
			j AS (
				SELECT user_2_id FROM Chat.chats WHERE id = $1 AND user_1_id = $2 AND user_2_id != $2
			)
			SELECT i.user_1_id FROM i
			UNION ALL
			SELECT j.user_2_id FROM j
			WHERE j.user_2_id IS NOT NULL
			LIMIT 1;  -- Ограничиваем результат до одного значения
		`,

		id_chat,
		id_user)
	errorr(err)

	var recipient int

	for request.Next() {
		err := request.Scan(
			&recipient,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	type A struct {
		Recipient int       `json:"Recipient"`
		Text      string    `json:"Text"`
		Date      time.Time `json:"Date"`
	}

	type Response struct {
		Status  string `json:"status"`
		Data    A      `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && recipient != 0 {
		response := Response{
			Status:  "success",
			Data:    A{Recipient: recipient, Text: text, Date: date},
			Message: "Уведомление показано",
		}

		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "fatal",
		Message: "Такого сообщения не найденно ",
	}

	if err != nil {
		response.Message = err.Error()
	}

	json.NewEncoder(rw).Encode(response)
}

func (repo *MyRepository) SendMessageAndMediaSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, id_user int, text string, file_paths []string) (err error) {
	request, err := rep.Query(
		ctx,
		`
		INSERT INTO Chat.messages(chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id;
	`,

		id_chat,
		id_user,
		text,
	)
	errorr(err)

	var mess_id int

	for request.Next() {
		err := request.Scan(
			&mess_id,
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

	if err == nil && mess_id != 0 {
		request, err := rep.Query(
			ctx,
			`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

			mess_id)
		errorr(err)

		var sent_at time.Time

		for request.Next() {
			err := request.Scan(&sent_at)

			if err != nil {
				fmt.Println(err)

				continue
			}
		}
	} else {

		response := Response{
			Status:  "fatal",
			Message: "Не отправленно",
		}

		json.NewEncoder(rw).Encode(response)
	}

	for i := range file_paths {
		request, err := rep.Query(
			ctx,
			`
				INSERT INTO Chat.Attachments (message_id, path_to_file)
				SELECT $1, $2
				RETURNING id;
			`,

			mess_id,
			file_paths[i],
		)
		errorr(err)

		var attachments_id int

		for request.Next() {
			err := request.Scan(
				&attachments_id,
			)
			if err != nil {
				fmt.Println(err)

				continue
			}
		}

		type A struct {
			Mess_id  int `json:"Mess_id"`
			Photo_id int `json:"Photo_id"`
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		image, err := DownloadFile(file_paths[i])

		if err == nil && mess_id != 0 {
			response := Response{
				Status:  "success",
				Data:    image,
				Message: fmt.Sprintf("Фото № %d доставленно. ", attachments_id, "Текст № %d доставленн.", mess_id),
			}

			request, err := rep.Query(
				ctx,
				`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

				mess_id)
			errorr(err)

			var sent_at time.Time

			for request.Next() {
				err := request.Scan(&sent_at)

				if err != nil {
					fmt.Println(err)

					continue
				}
			}

			json.NewEncoder(rw).Encode(response)

			// Notofication(rep, ctx, rw, id_chat, id_user, text, sent_at) //уведомление пользователя

		} else {

			response := Response{
				Status:  "fatal",
				Message: "Не отправленно",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)

		}
	}
	return err
}

func (repo *MyRepository) SendImageSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, id_user int, file_path []string) (err error) {
	request, err := rep.Query(
		ctx,
		`
		INSERT INTO Chat.messages(chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id;
	`,

		id_chat,
		id_user,
		"$IMAGE$",
	)
	errorr(err)

	var mess_id int

	for request.Next() {
		err := request.Scan(
			&mess_id,
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

	if err == nil && mess_id != 0 {
		response := Response{
			Status:  "success",
			Data:    mess_id,
			Message: "Текст сообщения принято",
		}

		request, err := rep.Query(
			ctx,
			`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

			mess_id)
		errorr(err)

		var sent_at time.Time

		for request.Next() {
			err := request.Scan(&sent_at)

			if err != nil {
				fmt.Println(err)

				continue
			}
		}

		json.NewEncoder(rw).Encode(response)

		// Notofication(rep, ctx, rw, id_chat, id_user, text, sent_at) //уведомление пользователя
	} else {

		response := Response{
			Status:  "fatal",
			Message: "Не отправленно",
		}

		json.NewEncoder(rw).Encode(response)
	}

	for i := range file_path {

		request, err := rep.Query(
			ctx,
			`
				INSERT INTO Chat.Attachments (message_id, path_to_file)
				SELECT $1, $2
				RETURNING id;
			`,

			mess_id,
			file_path[i],
		)
		errorr(err)

		var attachments_id int

		for request.Next() {
			err := request.Scan(
				&attachments_id,
			)
			if err != nil {
				fmt.Println(err)

				continue
			}
		}

		type A struct {
			Mess_id  int `json:"Mess_id"`
			Photo_id int `json:"Photo_id"`
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		image, err := DownloadFile(file_path[i])

		if err == nil && mess_id != 0 {
			response := Response{
				Status:  "success",
				Data:    image,
				Message: fmt.Sprintf("Фото № %d доставленно", attachments_id),
			}

			request, err := rep.Query(
				ctx,
				`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

				mess_id)
			errorr(err)

			var sent_at time.Time

			for request.Next() {
				err := request.Scan(&sent_at)

				if err != nil {
					fmt.Println(err)

					continue
				}
			}

			json.NewEncoder(rw).Encode(response)

			// Notofication(rep, ctx, rw, id_chat, id_user, text, sent_at) //уведомление пользователя

		} else {

			response := Response{
				Status:  "fatal",
				Message: "Не отправленно",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)

		}
	}
	return err
}

func (repo *MyRepository) SendMessageSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, id_user int, text string) (err error) {
	//текст есть, но изображений нет
	request, err := rep.Query(
		ctx,
		`
			INSERT INTO Chat.messages(chat_id, sender_id, text)
			VALUES ($1, $2, $3)
			RETURNING id;
		`,

		id_chat,
		id_user,
		text,
	)
	errorr(err)

	var mess_id int

	for request.Next() {
		err := request.Scan(
			&mess_id,
		)
		if err != nil {
			fmt.Println(err)

			continue
		}
	}

	type Ints struct {
		Mess_id int
		Text    string
		Sent_at time.Time
	}

	type Response struct {
		Status  string `json:"status"`
		Data    Ints   `json:"data,omitempty"`
		Message string `json:"message"`
	}

	if err == nil && true {
		request, err := rep.Query(
			ctx,
			`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

			mess_id)
		errorr(err)

		var sent_at time.Time

		for request.Next() {
			err := request.Scan(&sent_at)

			if err != nil {
				fmt.Println(err)

				continue
			}
		}
		response := Response{
			Status:  "success",
			Data:    Ints{Mess_id: mess_id, Text: text, Sent_at: sent_at},
			Message: "Показано",
		}

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

func (repo *MyRepository) SendVideoSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_chat, id_user int, file_path []string) (err error) {
	request, err := rep.Query(
		ctx,
		`
		INSERT INTO Chat.messages(chat_id, sender_id, text)
		VALUES ($1, $2, $3)
		RETURNING id;
	`,

		id_chat,
		id_user,
		"$VIDEO$",
	)
	errorr(err)

	var mess_id int

	for request.Next() {
		err := request.Scan(
			&mess_id,
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

	if err == nil && mess_id != 0 {
		request, err := rep.Query(
			ctx,
			`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

			mess_id)
		errorr(err)

		var sent_at time.Time

		for request.Next() {
			err := request.Scan(&sent_at)

			if err != nil {
				fmt.Println(err)

				continue
			}
		}
	} else {

		response := Response{
			Status:  "fatal",
			Message: "Не отправленно",
		}

		json.NewEncoder(rw).Encode(response)
	}

	for i := range file_path {

		request, err := rep.Query(
			ctx,
			`
				INSERT INTO Chat.Attachments (message_id, path_to_file)
				SELECT $1, $2
				RETURNING id;
			`,

			mess_id,
			file_path[i],
		)
		errorr(err)

		var attachments_id int

		for request.Next() {
			err := request.Scan(
				&attachments_id,
			)
			if err != nil {
				fmt.Println(err)

				continue
			}
		}

		type A struct {
			Mess_id  int `json:"Mess_id"`
			Photo_id int `json:"Photo_id"`
		}

		type Response struct {
			Status  string `json:"status"`
			Data    string `json:"data,omitempty"`
			Message string `json:"message"`
		}

		image, err := DownloadFile(file_path[i])

		if err == nil && mess_id != 0 {
			response := Response{
				Status:  "success",
				Data:    image,
				Message: fmt.Sprintf("Фото № %d доставленно", attachments_id, "Сообщение № %d доставленно", mess_id),
			}

			request, err := rep.Query(
				ctx,
				`SELECT sent_at FROM Chat.messages WHERE id = $1;`,

				mess_id)
			errorr(err)

			var sent_at time.Time

			for request.Next() {
				err := request.Scan(&sent_at)

				if err != nil {
					fmt.Println(err)

					continue
				}
			}

			json.NewEncoder(rw).Encode(response)

			// Notofication(rep, ctx, rw, id_chat, id_user, text, sent_at) //уведомление пользователя

		} else {

			response := Response{
				Status:  "fatal",
				Message: "Не отправленно",
			}

			rw.WriteHeader(http.StatusOK)
			json.NewEncoder(rw).Encode(response)

		}
	}
	return err
}

func (repo *MyRepository) PrintChatSQL(ctx context.Context, rw http.ResponseWriter, rep *pgxpool.Pool, id_user int) (err error) {
	request, err := rep.Query(
		ctx,
		`
		WITH i AS (
			SELECT id AS chai_id, user_1_id, user_2_id
			FROM chat.chats
			WHERE user_1_id = $1 OR user_2_id = $1
		),
		latest_messages AS (
			SELECT DISTINCT ON (m.chat_id) m.sender_id, m.text, m.sent_at, m.chat_id, attachm.message_id
			FROM chat.messages m
			LEFT JOIN chat.attachments attachm ON attachm.message_id = m.id
			WHERE m.chat_id IN (SELECT chai_id FROM i)
			AND m.sender_id IN (SELECT user_1_id FROM i UNION SELECT user_2_id FROM i)
			ORDER BY m.chat_id, m.sent_at DESC
		),
		buddy AS (
			SELECT id AS chai_id,
				CASE 
					WHEN user_1_id != $1 THEN user_1_id 
					WHEN user_2_id != $1 THEN user_2_id 
				END AS buddy_id
			FROM chat.chats
			WHERE user_1_id = $1 OR user_2_id = $1
		),
		buddy_info AS (
			SELECT buddy.chai_id, name::text AS info
			FROM users.individual_user 
			JOIN buddy ON buddy.buddy_id = users.individual_user.user_id
			UNION
			SELECT buddy.chai_id, name_of_company::text AS info
			FROM users.company_user 
			JOIN buddy ON buddy.buddy_id = users.company_user.user_id
		),
		avatar AS (
			SELECT buddy.chai_id, avatar_path 
			FROM users.users 
			JOIN buddy ON buddy.buddy_id = users.users.id
		)
		SELECT 
			i.chai_id,
			latest_messages.sender_id,
			latest_messages.text,
			latest_messages.sent_at,
			buddy_info.info,
			COALESCE(avatar.avatar_path, '/root/home/beeline_project/media/user/image_10290308543_ava.png') AS avatar_path
		FROM i
		LEFT JOIN latest_messages ON latest_messages.chat_id = i.chai_id
		LEFT JOIN buddy_info ON buddy_info.chai_id = i.chai_id
		LEFT JOIN avatar ON avatar.chai_id = i.chai_id;
		`,

		id_user,
	)

	type Productt struct {
		Chat_id   int        `json:"Chat_id"`
		Sender_id *int       `json:"sender_id"`
		Text      *string    `json:"text"`
		Sent_at   *time.Time `json:"sent_at"`
		Info      *string    `json:"info"`
		Avatar    string     `json:"avatar"`
	}

	products := []Productt{}

	type Response struct {
		Status  string     `json:"status"`
		Data    []Productt `json:"data,omitempty"`
		Message string     `json:"message"`
	}

	var Avatar_path string

	for request.Next() {
		p := Productt{}
		err := request.Scan(
			&p.Chat_id,
			&p.Sender_id,
			&p.Text,
			&p.Sent_at,
			&p.Info,
			&Avatar_path,
		)

		if err != nil {
			response := Response{
				Status:  "fatal",
				Message: "Возникла ошибка" + err.Error(),
			}

			json.NewEncoder(rw).Encode(response)
			return err
		}

		p.Avatar, err = DownloadFile(Avatar_path)
		// if err != nil {
		// 	response := Response{
		// 		Status:  "fatal",
		// 		Message: "Возникла ошибка с изображением " + err.Error() + ". Id чата: " + strconv.Itoa(p.Chat_id),
		// 	}

		// 	json.NewEncoder(rw).Encode(response)
		// }

		products = append(products, Productt{
			Chat_id:   p.Chat_id,
			Sender_id: p.Sender_id,
			Text:      p.Text,
			Sent_at:   p.Sent_at,
			Info:      p.Info,
			Avatar:    p.Avatar,
		})
	}

	if err != nil {
		response := Response{
			Status:  "fatal",
			Message: "Не сгруппировано",
		}

		json.NewEncoder(rw).Encode(response)
		return err
	} else if len(products) == 0 {
		response := Response{
			Status:  "success",
			Data:    products,
			Message: "Чатов не найденно, либо они не созданны",
		}

		json.NewEncoder(rw).Encode(response)

		return
	}

	response := Response{
		Status:  "success",
		Data:    products,
		Message: "Сгруппировано",
	}

	rw.WriteHeader(http.StatusOK)
	json.NewEncoder(rw).Encode(response)

	return
}
