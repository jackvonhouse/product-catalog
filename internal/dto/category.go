package dto

type Category struct {
	ID   int    `json:"id" default:"1"`
	Name string `json:"name" default:"Категория"`
}

type CreateCategory struct {
	Name string
}

type GetCategory struct {
	Limit  int
	Offset int
}

type UpdateCategory struct {
	ID   int
	Name string
}
