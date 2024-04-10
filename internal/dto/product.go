package dto

type Product struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type CreateProduct struct {
	Name       string `json:"name"`
	CategoryId int    `json:"category_id"`
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
