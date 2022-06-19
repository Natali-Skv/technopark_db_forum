package models

type Post struct {
	Id         int    `json:"id"`
	AuthorNick string `json:"author"`
	ParentId   int    `json:"parent"`
	Message    string `json:"message"`
	IsEdited   bool   `json:"isEdited"`
	ForumSlug  string `json:"forum"`
	ThreadId   int    `json:"thread"`
	ThreadSlug string `json:"-"`
	Created    string `json:"created"`
}
