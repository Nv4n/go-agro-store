package server

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
var validate *validator.Validate

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

	validate, err = NewValidator()
	if err != nil {
		log.Fatalf("failed to initialize validator: %v", err)
	}

	router := gin.Default()
	router.Static("/public", "./public")
	router.Static("/upload", "./upload")
	router.MaxMultipartMemory = 8 << 20

	// --- Route definitions ---

	// GET "/" redirects to /products.
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/products")
	})

	// GET /products with optional filters: ?tag=...&order=asc|desc|newest
	router.GET("/products", func(c *gin.Context) {
		productName := c.Query("name")
		productType := c.Query("type")
		var products []db.ListAllProductsRow

		if productName != "" {
			p, err := dbQueries.GetProductByName(c, productName)
			if err != nil {
				slog.Warn(fmt.Sprintf("failed to get product by name: %s", productName))
				slog.Info(err.Error())
				c.Redirect(http.StatusFound, "/products")
				c.Abort()
				return
			}
			products = append(products, db.ListAllProductsRow{ID: p.ID,
				Name:        p.Name,
				Price:       p.Price,
				Discount:    p.Discount,
				Description: p.Description,
				CreatedAt:   p.CreatedAt,
				UpdatedAt:   p.UpdatedAt,
				Type:        p.Type,
				Category:    p.Category})
		} else if productType != "" {
			prods, err := dbQueries.ListAllProductsByType(c, productName)
			if err != nil {
				slog.Warn(fmt.Sprintf("failed to list products by type: %s", productType))
				c.Redirect(http.StatusFound, "/products")
				c.Abort()
				return
			}
			for _, p := range prods {
				products = append(products, db.ListAllProductsRow{ID: p.ID,
					Name:        p.Name,
					Price:       p.Price,
					Discount:    p.Discount,
					Description: p.Description,
					CreatedAt:   p.CreatedAt,
					UpdatedAt:   p.UpdatedAt,
					Type:        p.Type,
					Category:    p.Category})
			}
		} else {
			products, err = dbQueries.ListAllProducts(c)
			if err != nil {
				slog.Info("Failed to list all products in /products")
				slog.Warn(fmt.Sprintf("failed to list all products: %v", err))
				return
			}
		}

		err = views.ProductsPage(products).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatalf("failed to render in /products: %v", err)
		}
	})

	// GET & POST /products/create.
	router.GET("/products/create", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		categories, err := dbQueries.ListAllCategoryTags(c)
		if err != nil {
			slog.Warn(err.Error())
			categories = []db.ListAllCategoryTagsRow{}
		}
		err = views.CreateProductPage(categories, "").Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatalf("failed to render in /products/create: %v", err)
		}
	})

	router.POST("/products/create", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		categories, err := dbQueries.ListAllCategoryTags(c)
		if err != nil {
			categories = []db.ListAllCategoryTagsRow{}
		}

		var productForm ProductCreateEdit
		err = c.ShouldBind(&productForm)
		if err != nil {
			slog.Warn(err.Error())
			err = views.CreateProductPage(categories, "wrong fields").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}
		err = validate.Struct(productForm)
		formErrMsg := ""
		if err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				curr := fmt.Sprintf("Field: %v, Error: %v. ", err.StructField(), err.Tag())
				formErrMsg += curr
				slog.Warn(curr)
			}
			err = views.CreateProductPage(categories, formErrMsg).Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}

		file, err := c.FormFile("file")
		if err != nil {
			slog.Warn(err.Error())
			err = views.CreateProductPage(categories, err.Error()).Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}

		uploadDir := "./upload"

		ext := filepath.Ext(file.Filename)
		if ext != ".svg" && ext != ".jpeg" && ext != ".jpg" && ext != ".png" {
			err = views.CreateProductPage(categories, "File must be an image").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}

		newFileName := fmt.Sprintf("IMG-%d%s", time.Now().Unix(), ext)
		dst := filepath.Join(uploadDir, newFileName)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			slog.Warn(err.Error())
			err = views.CreateProductPage(categories, "failed to save file").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}
		log.Println("Uploaded:", dst)

		priceNumeric := pgtype.Numeric{}
		err = priceNumeric.Scan(productForm.Price)
		if err != nil {
			slog.Warn(err.Error())
			_ = os.Remove(dst)
			err = views.CreateProductPage(categories, "Failed to get price").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}
		typeTag, err := dbQueries.GetTagByName(c, productForm.Type)
		if err != nil {
			typeTag, err = dbQueries.CreateTag(c, productForm.Type)
			if err != nil {
				slog.Warn(err.Error())
				_ = os.Remove(dst)
				err = views.CreateProductPage(categories, "Failed create type").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}
		}

		categoryTag, err := dbQueries.GetTagByName(c, productForm.Category)
		if err != nil {
			categoryTag, err = dbQueries.CreateTag(c, productForm.Category)
			if err != nil {
				slog.Warn(err.Error())
				_ = os.Remove(dst)
				err = views.CreateProductPage(categories, "Failed create category").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}
		}

		dbProduct := db.CreateProductParams{Name: productForm.Name,
			Price:       priceNumeric,
			Description: pgtype.Text{String: productForm.Description, Valid: true},
			Type:        typeTag.ID,
			Category:    categoryTag.ID,
			Img:         newFileName,
		}
		err = dbQueries.CreateProduct(c, dbProduct)
		if err != nil {
			slog.Warn(err.Error())
			_ = os.Remove(dst)
			err = views.CreateProductPage(categories, "Failed to create product try again!").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}

		c.Redirect(http.StatusFound, "/profile")
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

		product, err := dbQueries.GetProductById(c, productId)
		if err != nil {
			return
		}
		views.ProductPage(product)
	})
	router.GET("/products/:id/delete", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		productId, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in /products/:id/delete : %v", err))
			c.Redirect(http.StatusFound, "/profile")
			c.Abort()
			return
		}
		err = dbQueries.DeleteProduct(c, productId)
		if err != nil {
			slog.Warn(err.Error())
		}
		c.Redirect(http.StatusFound, "/profile")
		c.Abort()
	})

	// POST /products/:id/buy saves the current shopping list in the session.
	router.POST("/products/:id/buy", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(fmt.Sprintf("sessionStore.Get error: %v", err))
			c.Redirect(http.StatusFound, fmt.Sprintf("/products/%s", id))
			return
		}

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
		categories, err := dbQueries.ListAllCategoryTags(c)
		if err != nil {
			categories = []db.ListAllCategoryTagsRow{}
		}
		pid, err := StrToUUID(id)
		if err != nil {
			return
		}
		product, err := dbQueries.GetProductById(c, pid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Product not found /products/%s/edit: %v", id, err))
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		views.EditProductPage(product, categories, "")
	})

	router.POST("/products/:id/edit", func(c *gin.Context) {
		var categories []db.ListAllCategoryTagsRow
		var productForm ProductCreateEdit
		id := c.Param("id")
		pid, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in /products/:id/edit : %v", err))
			c.Redirect(http.StatusFound, "/profile")
			c.Abort()
			return
		}
		product, err := dbQueries.GetProductById(c, pid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Such product doesn't exist /products/:id/edit : %v", err))
			c.Redirect(http.StatusFound, "/profile")
			c.Abort()
			return
		}

		err = c.ShouldBind(&productForm)
		if err != nil {
			slog.Warn(err.Error())
			err = views.EditProductPage(product, categories, "wrong fields").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/:id/create : %v", err)
			}
			return
		}
		err = validate.Struct(productForm)
		formErrMsg := ""
		if err != nil {
			for _, err := range err.(validator.ValidationErrors) {
				curr := fmt.Sprintf("Field: %v, Error: %v. ", err.StructField(), err.Tag())
				formErrMsg += curr
				slog.Warn(curr)
			}
			err = views.EditProductPage(product, categories, formErrMsg).Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/:id/edit : %v", err)
			}
			return
		}

		file, err := c.FormFile("file")
		var dst string
		var newFileName string
		if err == nil {
			uploadDir := "./upload"

			ext := filepath.Ext(file.Filename)
			if ext != ".svg" && ext != ".jpeg" && ext != ".jpg" && ext != ".png" {
				err = views.EditProductPage(product, categories, "File must be an image").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}

			newFileName = fmt.Sprintf("IMG-%d%s", time.Now().Unix(), ext)
			dst = filepath.Join(uploadDir, newFileName)

			if err := c.SaveUploadedFile(file, dst); err != nil {
				slog.Warn(err.Error())
				err = views.EditProductPage(product, categories, "failed to save file").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}
			log.Println("Uploaded:", dst)

		}

		priceNumeric := pgtype.Numeric{}
		err = priceNumeric.Scan(productForm.Price)
		if err != nil {
			slog.Warn(err.Error())
			_ = os.Remove(dst)
			err = views.EditProductPage(product, categories, "Failed to get price").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}
		typeTag, err := dbQueries.GetTagByName(c, productForm.Type)
		if err != nil {
			typeTag, err = dbQueries.CreateTag(c, productForm.Type)
			if err != nil {
				slog.Warn(err.Error())
				_ = os.Remove(dst)
				err = views.EditProductPage(product, categories, "Failed create type").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}
		}

		categoryTag, err := dbQueries.GetTagByName(c, productForm.Category)
		if err != nil {
			categoryTag, err = dbQueries.CreateTag(c, productForm.Category)
			if err != nil {
				slog.Warn(err.Error())
				_ = os.Remove(dst)
				err = views.EditProductPage(product, categories, "Failed create category").Render(c.Request.Context(), c.Writer)
				if err != nil {
					log.Fatalf("failed to render in /products/create: %v", err)
				}
				return
			}
		}

		dbProduct := db.UpdateProductParams{Name: productForm.Name,
			Price:       priceNumeric,
			Description: pgtype.Text{String: productForm.Description, Valid: true},
			Type:        typeTag.ID,
			Category:    categoryTag.ID,
			Img:         newFileName,
		}
		err = dbQueries.UpdateProduct(c, dbProduct)
		if err != nil {
			slog.Warn(err.Error())
			_ = os.Remove(dst)
			err = views.EditProductPage(product, categories, "Failed to update product try again!").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatalf("failed to render in /products/create: %v", err)
			}
			return
		}

		c.Redirect(http.StatusFound, "/profile")
	})

	// GET /profile redirects to /users/:id based on session information.
	router.GET("/profile", authMiddleware(), func(c *gin.Context) {
		userID := c.MustGet("userID")
		c.Redirect(http.StatusFound, fmt.Sprintf("/users/%v", userID))
	})

	router.GET("/users/:id", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		uid, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in /users/:id : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/"))
			c.Abort()
			return
		}
		user, err := dbQueries.GetUserById(c, uid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error on finding user /users/:id : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/"))
			c.Abort()
			return
		}
		var products []db.ListAllProductsRow
		var orders []db.Order
		var users []db.ListAllUsersRow

		products, err = dbQueries.ListAllProducts(c)
		if err != nil {
			products = []db.ListAllProductsRow{}
		}
		orders, err = dbQueries.ListAllOrders(c)
		if err != nil {
			orders = []db.Order{}
		}
		users, err = dbQueries.ListAllUsers(c)
		if err != nil {
			users = []db.ListAllUsersRow{}
		}

		chats, err := dbQueries.ListAllChats(c)
		if err != nil {
			chats = []db.Chat{}
		}

		err = views.UserPage(user, true, products, orders, users, chats).Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatalf("Can't render /users/:id : %v", err)
		}
	})

	// GET & POST /users/:id/edit.
	router.GET("/users/:id/edit", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(fmt.Sprintf("Can't get session in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		suid, ok := session.Values["userID"].(string)
		if !ok {
			slog.Warn(fmt.Sprintf("Session uid is not string in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		requestUid, err := StrToUUID(suid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Session uid is not UUID in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		requestUser, err := dbQueries.GetUserById(c, requestUid)
		if err != nil {
			slog.Warn(fmt.Sprintf("No such user as requester in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}

		if id != suid && requestUser.Role != "admin" {
			slog.Warn(fmt.Sprintf("Forbidden in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		uid, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}

		user, err := dbQueries.GetUserById(c, uid)
		if err != nil {
			slog.Warn(fmt.Sprintf("No such user in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		err = views.UserEditPage("", user, requestUser.Role == "admin").Render(c, c.Writer)
		if err != nil {
			slog.Warn(fmt.Sprintf("Can't render /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, "/profile")
			return
		}
	})
	router.POST("/users/:id/edit", authMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(fmt.Sprintf("Can't get session in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		suid, ok := session.Values["userID"].(string)
		if !ok {
			slog.Warn(fmt.Sprintf("Session uid is not string in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		requestUid, err := StrToUUID(suid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Session uid is not UUID in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		requestUser, err := dbQueries.GetUserById(c, requestUid)
		if err != nil {
			slog.Warn(fmt.Sprintf("No such user as requester in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}

		if id != suid && requestUser.Role != "admin" {
			slog.Warn(fmt.Sprintf("Forbidden in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}
		uid, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}

		user, err := dbQueries.GetUserById(c, uid)
		if err != nil {
			slog.Warn(fmt.Sprintf("No such user in /users/:id/edit : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			return
		}

		var userForm UserEdit
		err = c.ShouldBind(&userForm)
		if err != nil {
			slog.Warn(fmt.Sprintf("Can't render /users/:id/edit : %v", err.Error()))
			views.UserEditPage("can't get fields", user, requestUser.Role == "admin")
			return
		}

		_, err = dbQueries.UpdateUserNames(c, db.UpdateUserNamesParams{ID: uid,
			Fname: userForm.FirstName,
			Lname: userForm.LastName})
		if err != nil {
			slog.Warn(fmt.Sprintf("Can't update user names /users/:id/edit : %v", err.Error()))
			views.UserEditPage("can't update user names try again", user, requestUser.Role == "admin")
			return
		}
		if requestUser.Role == "admin" {
			roleForm := c.GetString("role")
			if roleForm != "admin" && roleForm != "user" {
				slog.Warn(fmt.Sprintf("Can't update role /users/:id/edit : %v", err.Error()))
				views.UserEditPage("User role is invalid", user, requestUser.Role == "admin")
				return
			}
			if suid != id {
				_, err = dbQueries.UpdateUserRole(c, db.UpdateUserRoleParams{ID: uid, Role: db.UserRole(roleForm)})
				if err != nil {
					slog.Warn(fmt.Sprintf("Can't update role /users/:id/edit : %v", err.Error()))
					views.UserEditPage("Can't update role try again", user, requestUser.Role == "admin")
					return
				}
			}
		}

		c.Redirect(http.StatusFound, "/profile")
		c.Abort()
	})

	// DELETE /users/:id.
	router.DELETE("/users/:id", authMiddleware(), adminMiddleware(), func(c *gin.Context) {
		id := c.Param("id")
		uuid, err := StrToUUID(id)
		if err != nil {
			slog.Warn(fmt.Sprintf("Id is not UUID in DELETE /users/:id/ : %v", err.Error()))
			c.Redirect(http.StatusFound, fmt.Sprintf("/profile"))
			c.Abort()
			return
		}
		err = dbQueries.DeleteUser(c, uuid)
		if err != nil {
			slog.Warn(fmt.Sprintf("Error delete user: %v", err.Error()))
		}
	})

	// GET & POST /login.
	router.GET("/login", notAuthMiddleware(), func(c *gin.Context) {
		err = views.LoginPage("").Render(c.Request.Context(), c.Writer)
		if err != nil {
			log.Fatal(err)
		}
	})
	router.POST("/login", notAuthMiddleware(), func(c *gin.Context) {
		var userForm UserLogin
		err := c.ShouldBind(&userForm)
		if err != nil {
			slog.Warn(err.Error())
			err = views.LoginPage("Wrong fields").Render(c.Request.Context(), c.Writer)
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
			err = views.LoginPage(formErrMsg).Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		user, err := dbQueries.GetUserByEmail(c, userForm.Email)
		if err != nil {
			slog.Warn(err.Error())
			err = views.LoginPage("Email or password are wrong").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(userForm.Password))
		if err != nil {
			slog.Warn(err.Error())
			err = views.LoginPage("Email or password are wrong").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		session, err := sessionStore.New(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't login try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		session.Values["userID"] = user.ID.String()
		log.Println("SID:")
		log.Println(session.ID)
		err = sessionStore.Save(c.Request, c.Writer, session)
		if err != nil {
			slog.Warn(err.Error())
			err = views.RegisterPage("Couldn't login try again").Render(c.Request.Context(), c.Writer)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		c.Set("userID", user.ID.String())
		c.Redirect(http.StatusFound, "/")
	})

	// GET & POST /register.
	router.GET("/register", notAuthMiddleware(), func(c *gin.Context) {
		err = views.RegisterPage("").Render(c.Request.Context(), c.Writer)
		if err != nil {
			slog.Warn("From GET /register:")
			log.Fatal(err)
		}
	})

	router.POST("/register", notAuthMiddleware(), func(c *gin.Context) {
		var userForm UserRegister
		err := c.ShouldBind(&userForm)
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

	router.GET("/logout", authMiddleware(), func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			slog.Warn(err.Error())
			c.Redirect(http.StatusFound, "/")
			return
		}
		session.Options.MaxAge = -1
		err = sessionStore.Save(c.Request, c.Writer, session)
		if err != nil {
			slog.Warn(err.Error())
		}
		c.Redirect(http.StatusFound, "/")
	})

	router.POST("/orders/create", authMiddleware(), func(c *gin.Context) {
		// TODO: Create a new order.
		c.Redirect(http.StatusFound, "/orders")
	})

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
	router.GET("/chats/:id", authMiddleware(), func(c *gin.Context) {
		// TODO: List all chats.
		c.HTML(http.StatusOK, "chats.tmpl", nil)
	})

	// Start the server.
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
