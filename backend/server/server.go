package backend

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"agro.store/backend/db"
	"agro.store/backend/pgstore"
	"agro.store/frontend/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/joho/godotenv"
)

var sessionStore *pgstore.PGStore

var DefaultSessionName = "session-name"
var dbQueries *db.Queries

// authMiddleware checks for a valid session stored in our PostgreSQL using the modernized pgstore.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil || session.IsNew {
			//c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if userID, ok := session.Values["userID"]; ok {
			c.Set("userID", userID)
		} else {
			//c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// adminMiddleware is a stub for routes restricted to administrators.
func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		unparsedUserID, ok := c.Get("userID")
		if !ok {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		userID, ok := unparsedUserID.(pgtype.UUID)
		if !ok {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		u, err := dbQueries.GetUserById(c, userID)
		if err != nil {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if u.Role != "admin" {
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

// orderOwnerOrAdminMiddleware restricts order routes to the order owner or admins.
func orderOwnerOrAdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		unparsedUserID, ok := c.Get("userID")
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID, ok := unparsedUserID.(pgtype.UUID)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		u, err := dbQueries.GetUserById(c, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		var orderId pgtype.UUID
		err = orderId.Scan(c.Param("id"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		o, err := dbQueries.GetOrderById(c, orderId)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Server error"})
			return
		}

		if u.ID != o.UserID {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}
		c.Next()
	}
}
func init() {
	_ = godotenv.Load()
}

func StartServer() {
	dbURL := os.Getenv("DB_URI")
	var err error
	sessionStore, err = pgstore.NewPGStore(dbURL, []byte("your-secret-key"))
	if err != nil {
		log.Fatalf("failed to initialize pgstore: %v", err)
	}
	defer sessionStore.Close()

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to initialize")
	}
	defer conn.Close(ctx)

	dbQueries = db.New(conn)

	// Initialize Gin router and load HTML templates.
	router := gin.Default()
	router.Static("/public", "./public")
	// --- Route definitions ---

	// GET "/" redirects to /products.
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/products")
	})

	// GET /products with optional filters: ?tag=...&order=asc|desc|newest
	router.GET("/products", authMiddleware(), func(c *gin.Context) {
		//tag := c.Query("tag")
		//order := c.Query("order")
		// TODO: Query your product database applying optional filters.
		var products []db.Product
		_ = views.ProductsPage(products).Render(c.Request.Context(), c.Writer)
	})

	// GET & POST /products/create.
	router.GET("/products/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create_product.tmpl", nil)
	})

	router.POST("/products/create", func(c *gin.Context) {
		// TODO: Insert logic to create a new product.
		c.Redirect(http.StatusFound, "/products")
	})

	// GET & DELETE /products/:id.
	router.GET("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Retrieve product details.
		c.HTML(http.StatusOK, "product_detail.tmpl", gin.H{"id": id})
	})
	router.DELETE("/products/:id", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Delete the product with the given id.
		c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
	})

	// POST /products/:id/buy saves the current shopping list in the session.
	router.POST("/products/:id/buy", func(c *gin.Context) {
		id := c.Param("id")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session error"})
			return
		}
		// Retrieve or initialize the shopping list.
		shoppingList, ok := session.Values["shoppingList"].([]string)
		if !ok {
			shoppingList = []string{}
		}
		shoppingList = append(shoppingList, id)
		session.Values["shoppingList"] = shoppingList
		if err := sessionStore.Save(c.Request, c.Writer, session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update shopping list"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "added", "shoppingList": shoppingList})
	})

	// GET & POST /products/:id/edit.
	router.GET("/products/:id/edit", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Retrieve product for editing.
		c.HTML(http.StatusOK, "edit_product.tmpl", gin.H{"id": id})
	})
	router.POST("/products/:id/edit", func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Update product details.
		c.Redirect(http.StatusFound, fmt.Sprintf("/products/%s", id))
	})

	// GET /profile redirects to /users/:id based on session information.
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
		// TODO: Update user profile.
		c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v", id))
	})

	// DELETE /user/:id.
	router.DELETE("/user/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Delete the user.
		c.JSON(http.StatusOK, gin.H{"status": "user deleted", "id": id})
	})

	// GET & POST /login.
	router.GET("/login", func(c *gin.Context) {
		err = views.LoginPage().Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatal(err)
		}
	})
	router.POST("/login", func(c *gin.Context) {
		// For demonstration, assume a "userID" is provided.
		userID := c.PostForm("userID")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Session error"})
			return
		}
		session.Values["userID"] = userID
		if err := sessionStore.Save(c.Request, c.Writer, session); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not save session"})
			return
		}
		c.Redirect(http.StatusFound, "/profile")
	})

	// GET & POST /register.
	router.GET("/register", func(c *gin.Context) {
		err = views.RegisterPage().Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatal(err)
		}
	})
	router.POST("/register", func(c *gin.Context) {
		// TODO: Add user registration logic.
		c.Redirect(http.StatusFound, "/login")
	})

	// Admin-restricted orders routes.
	ordersGroup := router.Group("/orders", authMiddleware(), adminMiddleware())
	{
		ordersGroup.GET("", func(c *gin.Context) {
			// TODO: List all orders.
			c.HTML(http.StatusOK, "orders.tmpl", nil)
		})
		ordersGroup.POST("", func(c *gin.Context) {
			// TODO: Create a new order.
			c.Redirect(http.StatusFound, "/orders")
		})
	}

	// GET & POST /orders/:id restricted to order owner and admins.
	router.GET("/orders/:id", authMiddleware(), orderOwnerOrAdminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Show order details.
		c.HTML(http.StatusOK, "order_detail.tmpl", gin.H{"id": id})
	})
	router.POST("/orders/:id", authMiddleware(), orderOwnerOrAdminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Update order.
		c.Redirect(http.StatusFound, fmt.Sprintf("/orders/%s", id))
	})

	// GET /chat redirects to /chats/:id for the current user.
	router.GET("/chat", authMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID")
		c.Redirect(http.StatusFound, fmt.Sprintf("/chats/%v", userID))
	})

	// GET & POST /chats.
	router.GET("/chats", authMiddleware(), func(c *gin.Context) {
		// TODO: List all chats.
		c.HTML(http.StatusOK, "chats.tmpl", nil)
	})
	router.POST("/chats", authMiddleware(), func(c *gin.Context) {
		// TODO: Create a new chat.
		c.Redirect(http.StatusFound, "/chats")
	})

	// GET & POST /chats/:id/messages.
	router.GET("/chats/:id/messages", authMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")
		// TODO: Retrieve chat messages.
		c.HTML(http.StatusOK, "chat_messages.tmpl", gin.H{"chatID": chatID})
	})
	router.POST("/chats/:id/messages", authMiddleware(), func(c *gin.Context) {
		chatID := c.Param("id")
		// TODO: Post a new message.
		c.Redirect(http.StatusFound, fmt.Sprintf("/chats/%s/messages", chatID))
	})

	// Start the server.
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
