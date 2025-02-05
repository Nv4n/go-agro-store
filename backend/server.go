package backend

import (
	"agro.store/backend/pgstore"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

// global session store variable
var sessionStore *pgstore.PGStore

// initSessionStore initializes the session store and creates the sessions table if needed.
func initSessionStore(dsn string, key []byte) {
	// Open the database using pgx as the driver
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	// Create sessions table if it does not exist.
	createSessionsTable(db)

	// Initialize the PGStore. pgstore.NewPGStore uses a *sql.DB so pgx works here.
	store, err := pgstore.NewPGStore(db, key)
	if err != nil {
		log.Fatalf("Failed to create PGStore: %v", err)
	}
	sessionStore = store
}

// createSessionsTable creates the table that will hold session records.
func createSessionsTable(db *sql.DB) {
	// Adjust this SQL statement to match the expected schema by pgstore.
	const query = `
	CREATE TABLE IF NOT EXISTS http_sessions (
		key TEXT PRIMARY KEY,
		data BYTEA NOT NULL,
		expiry TIMESTAMP NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		log.Fatalf("Failed to create sessions table: %v", err)
	}
}

// authMiddleware uses the session store to check if a valid session exists.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Retrieve session using Gorilla sessions.
		session, err := sessionStore.Get(c.Request, "session-name")
		if err != nil || session.IsNew {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Save user id in the context if it exists.
		if userID, ok := session.Values["userID"]; ok {
			c.Set("userID", userID)
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

// Stub middleware: only allows admins.
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Insert your admin-check logic here. This is just a placeholder.
		isAdmin := true // Replace with actual check
		if !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admins only"})
			return
		}
		c.Next()
	}
}

// Stub middleware: only allow order owner or admins.
func orderOwnerOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Insert logic to check if current user is order owner or admin.
		isOwnerOrAdmin := true // Replace with actual check
		if !isOwnerOrAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}
		c.Next()
	}
}

func main() {
	// Replace with your actual PostgreSQL DSN and secret key.
	dsn := "postgres://username:password@localhost:5432/dbname"
	secretKey := []byte("your-secret-key")
	initSessionStore(dsn, secretKey)

	// Initialize Gin and load Go templates (using go-templ files under templates/).
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// --- Route definitions ---

	// GET "/" redirects to /products.
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/products")
	})

	// GET /products with optional filters: ?tag=...&order=asc|desc|newest
	router.GET("/products", func(c *gin.Context) {
		tag := c.Query("tag")
		order := c.Query("order")
		// TODO: query your database for products applying tag and order filters if provided.
		products := []string{"Product1", "Product2"} // dummy data
		c.HTML(http.StatusOK, "products.tmpl", gin.H{
			"products": products,
			"tag":      tag,
			"order":    order,
		})
	})

	// GET & POST /products/create.
	router.GET("/products/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create_product.tmpl", nil)
	})
	router.POST("/products/create", func(c *gin.Context) {
		// TODO: create a new product using form data.
		c.Redirect(http.StatusFound, "/products")
	})

	// GET & DELETE /products/:id.
	router.GET("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: fetch product details by id.
		c.HTML(http.StatusOK, "product_detail.tmpl", gin.H{"id": id})
	})
	router.DELETE("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: delete the product with the given id.
		c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
	})

	// POST /products/:id/buy saves current shopping list in session (simulating localStorage).
	router.POST("/products/:id/buy", func(c *gin.Context) {
		id := c.Param("id")
		session, _ := sessionStore.Get(c.Request, "session-name")
		// Retrieve existing shopping list if available.
		shoppingList, ok := session.Values["shoppingList"].([]string)
		if !ok {
			shoppingList = []string{}
		}
		// Add product id to shopping list.
		shoppingList = append(shoppingList, id)
		session.Values["shoppingList"] = shoppingList
		if err := session.Save(c.Request, c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update shopping list"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "added to shopping list", "shoppingList": shoppingList})
	})

	// GET & POST /products/:id/edit.
	router.GET("/products/:id/edit", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: fetch product for editing.
		c.HTML(http.StatusOK, "edit_product.tmpl", gin.H{"id": id})
	})
	router.POST("/products/:id/edit", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: update product details using posted form data.
		c.Redirect(http.StatusFound, fmt.Sprintf("/products/%s", id))
	})

	// GET /profile redirects to /users/:id based on the session and shows user dashboard.
	router.GET("/profile", authMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID")
		c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v", userID))
	})

	// GET & POST /user/:id/edit.
	router.GET("/user/:id/edit", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		c.HTML(http.StatusOK, "edit_user.tmpl", gin.H{"id": id})
	})
	router.POST("/user/:id/edit", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: update user profile using form data.
		c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v", id))
	})

	// DELETE /user/:id.
	router.DELETE("/user/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: delete the user.
		c.JSON(http.StatusOK, gin.H{"status": "user deleted", "id": id})
	})

	// GET & POST /login.
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", nil)
	})
	router.POST("/login", func(c *gin.Context) {
		// TODO: verify user credentials.
		// For demonstration, assume a field "userID" is posted.
		userID := c.PostForm("userID")
		session, _ := sessionStore.Get(c.Request, "session-name")
		session.Values["userID"] = userID
		if err := session.Save(c.Request, c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
			return
		}
		c.Redirect(http.StatusFound, "/profile")
	})

	// GET & POST /register.
	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.tmpl", nil)
	})
	router.POST("/register", func(c *gin.Context) {
		// TODO: register the user.
		c.Redirect(http.StatusFound, "/login")
	})

	// Routes restricted to admins for /orders.
	ordersGroup := router.Group("/orders", authMiddleware(), adminMiddleware())
	{
		ordersGroup.GET("", func(c *gin.Context) {
			// TODO: list all orders.
			c.HTML(http.StatusOK, "orders.tmpl", nil)
		})
		ordersGroup.POST("", func(c *gin.Context) {
			// TODO: create a new order.
			c.Redirect(http.StatusFound, "/orders")
		})
	}

	// GET & POST /orders/:id restricted to order owner and admins.
	router.GET("/orders/:id", authMiddleware(), orderOwnerOrAdminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: show order details.
		c.HTML(http.StatusOK, "order_detail.tmpl", gin.H{"id": id})
	})
	router.POST("/orders/:id", authMiddleware(), orderOwnerOrAdminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: update order.
		c.Redirect(http.StatusFound, fmt.Sprintf("/orders/%s", id))
	})

	// GET /chat redirects to GET /chats/:id (using the current user id), restricted.
	router.GET("/chat", authMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID")
		c.Redirect(http.StatusFound, fmt.Sprintf("/chats/%v", userID))
	})

	// GET & POST /chats.
	router.GET("/chats", authMiddleware(), func(c *gin.Context) {
		// TODO: list all chats.
		c.HTML(http.StatusOK, "chats.tmpl", nil)
	})
	router.POST("/chats", authMiddleware(), func(c *gin.Context) {
		// TODO: create a new chat.
		c.Redirect(http.StatusFound, "/chats")
	})

	// GET & POST /chats/:id/messages.
	router.GET("/chats/:id/messages", authMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")
		// TODO: fetch messages for chat chatID.
		c.HTML(http.StatusOK, "chat_messages.tmpl", gin.H{"chatID": chatID})
	})
	router.POST("/chats/:id/messages", authMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")
		// TODO: post a new message to chat chatID.
		c.Redirect(http.StatusFound, fmt.Sprintf("/chats/%s/messages", chatID))
	})

	// Start the server.
	router.Run(":8080")
}
