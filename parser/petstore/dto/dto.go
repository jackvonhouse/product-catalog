package dto

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Pet struct {
	ID       int64    `json:"id"`
	Category Category `json:"category"`
	Name     string   `json:"name"`
}

type Result struct {
	ID      int64
	Success bool
	Error   error
	Source  string
}
