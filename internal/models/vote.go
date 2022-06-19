package models

type Vote struct {
	Nick       string `json:"nickname"`
	Voice      int    `json:"voice"`
	ThreadId   int    `json:"thread"`
	ThreadSlug string `json:"-"`
}

type Status struct {
	Users   int `json:"user"`
	Forums  int `json:"forum"`
	Threads int `json:"thread"`
	Posts   int `json:"post"`
}
