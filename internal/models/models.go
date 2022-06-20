package models

//easyjson:json
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

//easyjson:json
type Forum struct {
	Slug     string `json:"slug" db:"slug"`
	Title    string `json:"title" db:"title"`
	UserNick string `json:"user" db:"author_nick"`
	Posts    int    `json:"posts" db:"posts"`
	Threads  int    `json:"threads" db:"threads"`
}

//easyjson:json
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

//easyjson:json
type User struct {
	Name  string `json:"fullname"`
	Nick  string `json:"nickname"`
	Email string `json:"email"`
	About string `json:"about"`
}

//easyjson:json
type Vote struct {
	Nick       string `json:"nickname"`
	Voice      int    `json:"voice"`
	ThreadId   int    `json:"thread"`
	ThreadSlug string `json:"-"`
}

//easyjson:json
type Status struct {
	Users   int `json:"user"`
	Forums  int `json:"forum"`
	Threads int `json:"thread"`
	Posts   int `json:"post"`
}
