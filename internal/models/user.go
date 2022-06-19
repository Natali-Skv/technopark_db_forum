package models

type User struct {
	Name  string `json:"fullname"`
	Nick  string `json:"nickname"`
	Email string `json:"email"`
	About string `json:"about"`
}
