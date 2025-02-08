package main

import (
	"agro.store/backend/server"
	_ "github.com/a-h/templ"
	_ "github.com/gin-gonic/gin"
	_ "github.com/gorilla/sessions"
	_ "github.com/jackc/pgx/v5"
)

func main() {
	server.StartServer()
}
