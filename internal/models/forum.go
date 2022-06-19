package models

type Forum struct {
	Slug     string `json:"slug" db:"slug"`
	Title    string `json:"title" db:"title"`
	UserNick string `json:"user" db:"author_nick"`
	Posts    int    `json:"posts" db:"posts"`
	Threads  int    `json:"threads" db:"threads"`
}
