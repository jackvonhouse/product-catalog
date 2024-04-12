COVER = coverage.out

mockgen:
	mockgen -source=internal/service/product/product.go -destination=internal/service/product/product.mock.go -package=product
	mockgen -source=internal/service/category/category.go -destination=internal/service/category/category.mock.go -package=category
	mockgen -source=internal/usecase/product/product.go -destination=internal/usecase/product/product.mock.go -package=product
	mockgen -source=internal/usecase/category/category.go -destination=internal/usecase/category/category.mock.go -package=category
	mockgen -source=internal/transport/product/product.go -destination=internal/transport/product/product.mock.go -package=product
	mockgen -source=internal/transport/category/category.go -destination=internal/transport/category/category.mock.go -package=category

cover:
	go test ./... -short -count=100 -race -coverprofile=$(COVER) -v -cover
	go tool cover -html=$(COVER)
	rm -f $(COVER)
