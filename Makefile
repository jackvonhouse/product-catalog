mockgen:
	mockgen -source=internal/service/product/product.go -destination=internal/service/product/product.mock.go -package=product
