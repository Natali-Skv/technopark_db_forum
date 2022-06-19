package models

type Thread struct {
	Id         int    `json:"id"`
	Slug       string `json:"slug" db:"slug"`
	Title      string `json:"title" db:"title"`
	AuthorNick string `json:"author" db:"author_nick"`
	ForumSlug  string `json:"forum"`
	Message    string `json:"message"`
	Votes      int    `json:"votes"`
	Created    string `json:"created"`
}
