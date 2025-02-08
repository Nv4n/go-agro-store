package main

import (
	"agro.store/backend/server"
	_ "github.com/a-h/templ"
	_ "github.com/gin-gonic/gin"
	_ "github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/joho/godotenv"
)

func main() {
	server.StartServer()
}
