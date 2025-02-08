package server

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"agro.store/backend/db"
	"agro.store/backend/pgstore"
	"agro.store/frontend/views"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

var sessionStore *pgstore.PGStore

var DefaultSessionName = "session-name"
var DefaultSecretKey = "your-secret-key"
var dbQueries *db.Queries

func init() {
	_ = godotenv.Load()
}

func StartServer() {
	dbURL := os.Getenv("DB_URI")
	var err error
	sessionStore, err = pgstore.NewPGStore(dbURL, []byte(DefaultSecretKey))
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

	router := gin.Default()
	router.Static("/public", "./public")
	router.Static("/upload", "./upload")

	// --- Route definitions ---

	// GET "/" redirects to /products.
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/products")
	})

	// GET /products with optional filters: ?tag=...&order=asc|desc|newest
	router.GET("/products", func(c *gin.Context) {
		//tag := c.Query("tag")
		//order := c.Query("order")
		// TODO: Query your product database applying optional filters.

		products, err := dbQueries.ListAllProducts(c)
		if err != nil {
			slog.Info("Failed to list all products in /products")
			slog.Warn(fmt.Sprintf("failed to list all products: %v", err))
			return
		}
		err = views.ProductsPage(products).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatalf("failed to render in /products: %v", err)
		}
	})

	// GET & POST /products/create.
	router.GET("/products/create", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		//TODO implement templ
		err = views.CreateProductPage().Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatalf("failed to render in /products/create: %v", err)
		}
	})

	router.POST("/products/create", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		// TODO: Insert logic to create a new product.
		c.Redirect(http.StatusFound, "/products")
	})

	// GET & DELETE /products/:id.
	router.GET("/products/:id", func(c *gin.Context) {
		productId, err := StrToUUID(c.Param("id"))
		if err != nil {
			slog.Warn(fmt.Sprintf(""))
			c.Redirect(http.StatusFound, "/products")
			c.Abort()
			return
		}
		// TODO: Retrieve product details.
		product, err := dbQueries.GetProductById(c, productId)
		if err != nil {
			return
		}
		views.ProductPage(product)
	})
	router.DELETE("/products/:id", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Delete the product with the given id.
		c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
	})

	// POST /products/:id/buy saves the current shopping list in the session.
	router.POST("/products/:id/buy", authMiddleware(), func(c *gin.Context) {
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
	router.GET("/products/:id/edit", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
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
	router.DELETE("/user/:id", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		// TODO: Delete the user.
		c.JSON(http.StatusOK, gin.H{"status": "user deleted", "id": id})
	})

	// GET & POST /login.
	router.GET("/login", func(c *gin.Context) {
		err = views.LoginPage("").Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatal(err)
		}
	})
	router.POST("/login", func(c *gin.Context) {
		// For demonstration, assume a "userID" is provided.
		userID := c.PostForm("userID")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(fmt.Sprintf("From POST /login: %v", err))

			err = views.LoginPage("Couldn't login try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			c.Abort()
			return
		}
		session.Values["userID"] = userID
		if err := sessionStore.Save(c.Request, c.Writer, session); err != nil {
			slog.Warn(fmt.Sprintf("From POST /login: %v", err))
			err = views.LoginPage("Couldn't login try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			c.Abort()
			return
		}

		c.Redirect(http.StatusFound, "/")
		c.Abort()
	})

	// GET & POST /register.
	router.GET("/register", func(c *gin.Context) {
		err = views.RegisterPage("").Render(c.Request.Context(), c.Writer)
		if err != nil {
			slog.Warn("From GET /register:")
			log.Fatal(err)
		}
	})

	router.POST("/register", func(c *gin.Context) {
		validate := validator.New()
		var userForm UserRegister
		err := c.Bind(&userForm)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Wrong fields").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
		}

		err = validate.Struct(userForm)
		formErrMsg := ""
		if err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				curr := fmt.Sprintf("Field: %v, Error: %v. ", err.StructField(), err.Tag())
				formErrMsg += curr
				slog.Warn(curr)
			}
			err = views.RegisterPage(formErrMsg).Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		_, err = dbQueries.GetUserByEmail(c, userForm.Email)
		if err == nil {
			err = views.RegisterPage("Such user already exists").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(userForm.Password), bcrypt.MinCost)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't register try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		user := db.CreateUserParams{
			Email:    userForm.Email,
			Fname:    userForm.FirstName,
			Lname:    userForm.LastName,
			Password: string(hash),
			Role:     "user"}

		createUser, err := dbQueries.CreateUser(c, user)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't register try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't register try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		session.Values["userID"] = createUser.String()
		err = sessionStore.Save(c.Request, c.Writer, session)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't register try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		c.Set("userID", createUser.String())

		c.Redirect(http.StatusFound, "/")
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
