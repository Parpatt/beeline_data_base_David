package models

import (
	"time"
)

type User struct {
	Id            int       `json:"id" db:"id"`
	Password_hash string    `json:"password_hash" db:"password_hash"`
	Email         string    `json:"email" db:"email"`
	Phone_number  string    `json:"phone_number" db:"phone_number"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
	Updated_at    time.Time `json:"updated_at" db:"updated_at"`
	Avatar_path   string    `json:"avatar_path" db:"avatar_path"`
	User_type     int       `json:"user_type" db:"user_type"`
	User_role     int       `json:"user_role" db:"user_role"`
}

type Review struct {
	Id          int       `json:"id" db:"id"`
	Ad_id       int       `json:"ad_id" db:"ad_id"`
	Reviewer_id int       `json:"reviewer_id" db:"reviewer_id"`
	Rating      int       `json:"rating" db:"rating"`
	Comment     string    `json:"comment" db:"comment"`
	Created_at  time.Time `json:"created_at" db:"created_at"`
	Updated_at  time.Time `json:"updated_at" db:"updated_at"`
}

type Order struct {
	Id          int       `json:"id" db:"id"`
	Ad_id       int       `json:"ad_id" db:"ad_id"`
	Renter_id   int       `json:"renter_id" db:"renter_id"`
	Total_price int       `json:"total_price" db:"total_price"`
	Created_at  time.Time `json:"created_at" db:"created_at"`
	Status      int       `json:"status" db:"status"`
}

type Bookings struct {
	Id             int       `json:"id" db:"id"`
	Order_id       int       `json:"order_id" db:"order_id"`
	Starts_at      time.Time `json:"starts_at" db:"starts_at"`
	Snds_at        time.Time `json:"ends_at" db:"ends_at"`
	Created_at     time.Time `json:"created_at" db:"created_at"`
	Amount         int       `json:"amount" db:"amount"`
	Transaction_id int       `json:"transaction_id" db:"transaction_id"`
	Type           int       `json:"type" db:"type"`
}

type Disputes struct {
	Id           int       `json:"id" db:"id"`
	Id_orders    int       `json:"id_orders" db:"id_orders"`
	Id_winner    int       `json:"id_winner" db:"id_winner"`
	Date_of_win  time.Time `json:"date_of_win" db:"date_of_win"`
	Date_of_call time.Time `json:"date_of_call" db:"date_of_call"`
	Comments     string    `json:"comments" db:"comments"`
}

// type Report struct {
// 	Id         int       `json:"id" db:"id"`
// 	Order_id   int       `json:"order_id" db:"order_id"`
// 	Created_at time.Time `json:"created_at" db:"created_at"`
// 	File_path  string    `json:"file_path" db:"file_path"`
// }

type Stories struct {
	Id         int       `json:"id" db:"id"`
	Photo_path string    `json:"photo_path" db:"photo_path"`
	Date_start time.Time `json:"date_start" db:"date_start"`
	Date_end   time.Time `json:"date_end" db:"date_end"`
}

type Transactions struct {
	Id         int       `json:"id" db:"id"`
	Wallet_id  int       `json:"wallet_id" db:"wallet_id"`
	Amount     int       `json:"amount" db:"amount"`
	Created_at time.Time `json:"created_at" db:"created_at"`
	User_2     int       `json:"user_2" db:"user_2"`
	Typee      int       `json:"typee" db:"typee"`
}

type Wallets struct {
	Id            int       `json:"id" db:"id"`
	User_id       int       `json:"user_id" db:"user_id"`
	Total_balance int       `json:"total_balance" db:"total_balance"`
	Frozen_funds  int       `json:"frozen_funds" db:"frozen_funds"`
	Created_at    time.Time `json:"created_at" db:"created_at"`
}

type Chat struct {
	Id          int  `json:"id" db:"id"`
	User_1_id   int  `json:"user_1_id" db:"user_1_id"`
	User_2_id   int  `json:"user_2_id" db:"user_2_id"`
	Have_disput bool `json:"have_disput" db:"have_disput"`
	Mediator_id int  `json:"mediator_id" db:"mediator_id"`
	State       bool `json:"state" db:"state"`
}

type Message struct {
	Id           int       `json:"id" db:"id"`
	Chat_id      int       `json:"chat_id" db:"chat_id"`
	Sender_id    int       `json:"sender_id" db:"sender_id"`
	Text         string    `json:"text" db:"text"`
	Sent_at      time.Time `json:"v" db:"sent_at"`
	Path_to_file string    `json:"path_to_file" db:"path_to_file"`
}
