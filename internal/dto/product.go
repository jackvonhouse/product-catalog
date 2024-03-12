package dto

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CreateProduct struct {
	Name string
}

type GetProduct struct {
	Limit  int
	Offset int
}

type UpdateProduct struct {
	ID   int
	Name string
}
