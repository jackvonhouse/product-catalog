// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/product": {
            "get": {
                "description": "Получение товаров",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Получить товары",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Лимит",
                        "name": "limit",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "Смещение",
                        "name": "offset",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/dto.Product"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Создание товара с определённой категорией",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "summary": "Создать товар",
                "parameters": [
                    {
                        "description": "Данные о товаре",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/dto.CreateProduct"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "id": {
                                    "type": "integer"
                                }
                            }
                        }
                    },
                    "409": {
                        "description": "Товар уже существует",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "error": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Неизвестная ошибка",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "error": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "dto.CreateProduct": {
            "type": "object",
            "properties": {
                "category_id": {
                    "type": "integer",
                    "default": 1
                },
                "name": {
                    "type": "string",
                    "default": "Товар"
                }
            }
        },
        "dto.Product": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer",
                    "default": 1
                },
                "name": {
                    "type": "string",
                    "default": "Товар"
                }
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "description": "Авторизация при помощи JWT-токена",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.1",
	Host:             "localhost:8081",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Каталог товаров",
	Description:      "Простейшее API для каталога товаров",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
