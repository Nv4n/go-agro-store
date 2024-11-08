package main

import (
	e "agro.store/example"
	t "agro.store/types"
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5"
	"log"
	"sync"
	"time"
)

func main() {
	engine := gin.Default()

	// Disable trusted proxy warning.
	_ = engine.SetTrustedProxies(nil)

	engine.GET("/", func(c *gin.Context) {
		// Create a channel to send deferred component renders to the template.
		data := make(chan t.SlotContents)

		// We know there are 3 slots, so start a WaitGround.
		var wg sync.WaitGroup
		wg.Add(3)

		// Start the async processes.
		// Sidebar.
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * 3)
			data <- t.SlotContents{
				Name:     "a",
				Contents: e.A(),
			}
		}()

		// Content.
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * 2)
			data <- t.SlotContents{
				Name:     "b",
				Contents: e.B(),
			}
		}()

		// Footer.
		go func() {
			defer wg.Done()
			time.Sleep(time.Second * 1)
			data <- t.SlotContents{
				Name:     "c",
				Contents: e.C(),
			}
		}()

		// Close the channel when all processes are done.
		go func() {
			wg.Wait()
			close(data)
		}()

		// Pass the channel to the template.
		component := e.Page(data)

		// Serve using the streaming mode of the handler.
		templ.Handler(component, templ.WithStreaming()).ServeHTTP(c.Writer, c.Request)

	})
	log.Fatal(engine.Run("localhost:8080"))
}
