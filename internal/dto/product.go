package dto

type Product struct {
	ID   int    `json:"id" db:"id" default:"1"`
	Name string `json:"name" db:"name" default:"Товар"`
}

type CreateProduct struct {
	Name       string `json:"name" default:"Товар"`
	CategoryId int    `json:"category_id" default:"1"`
}

type GetProduct struct {
	Limit  int
	Offset int
}

type UpdateProduct struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	OldCategoryId int    `json:"old_category_id"`
	NewCategoryId int    `json:"new_category_id"`
}
