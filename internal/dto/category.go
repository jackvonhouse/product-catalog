package dto

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
