package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
)

func notAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err == nil && !session.IsNew {
			DefaultMiddlewareLog("From notAuthMiddleware()", "Authorized", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if _, ok := session.Values["userID"]; ok {
			DefaultMiddlewareLog("From notAuthMiddleware()", "Authorized", c, nil)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		slog.Info(session.ID)
		c.Next()
	}
}

// authMiddleware checks for a valid session stored in our PostgreSQL using the modernized pgstore.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil || session.IsNew {
			DefaultMiddlewareLog("From authMiddleware()", "Session error", c, err)
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		if userID, ok := session.Values["userID"]; ok {
			c.Set("userID", userID)
		} else {
			DefaultMiddlewareLog("From authMiddleware()", "userID not in Session Values", c, nil)
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
		userID, err := StrToUUID(c.GetString("userID"))
		if err != nil {
			DefaultMiddlewareLog("From adminMiddleware()", "userID is not UUID", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		u, err := dbQueries.GetUserById(c, userID)
		if err != nil {
			DefaultMiddlewareLog("From adminMiddleware()", "Can't query by userID", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if u.Role != "admin" {
			slog.Info("From adminMiddleware(): user is not admin")
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
		userID, err := StrToUUID(c.GetString("userID"))
		if err != nil {
			DefaultMiddlewareLog("From orderOwnerOrAdminMiddleware()", "userID is not UUID", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		u, err := dbQueries.GetUserById(c, userID)
		if err != nil {
			DefaultMiddlewareLog("From orderOwnerOrAdminMiddleware()", "Can't get user by userID", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		orderId, err := StrToUUID(c.Param("id"))
		if err != nil {
			DefaultMiddlewareLog("From orderOwnerOrAdminMiddleware()", "Param id is not UUID", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		o, err := dbQueries.GetOrderById(c, orderId)
		if err != nil {
			DefaultMiddlewareLog("From orderOwnerOrAdminMiddleware()", "Error querying order by id", c, err)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}

		if u.ID != o.UserID {
			slog.Info("From orderOwnerOrAdminMiddleware(): User is not Order User")
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		c.Next()
	}
}

// csrfMiddleware is the middleware function for CSRF protection
func csrfMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := sessionStore.Get(c.Request, DefaultSessionName)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		csrfToken, ok := session.Values[csrfTokenKey].(string)
		if !ok {
			csrfToken, err = GenerateCSRFToken()
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			session.Values[csrfTokenKey] = csrfToken
			if err := session.Save(c.Request, c.Writer); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		}

		if c.Request.Method != http.MethodGet {
			requestToken := c.Request.Header.Get("X-CSRF-Token")
			if requestToken == "" {
				requestToken = c.PostForm("csrf_token")
			}
			if csrfToken != requestToken {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Set the CSRF token in the context for access in handlers
		c.Set(csrfTokenKey, csrfToken)
		c.Next()
	}
}

func DefaultMiddlewareLog(from string, msg string, c *gin.Context, err error) {
	slog.Info(fmt.Sprintf("%s: %v", from, err))
	slog.Warn(fmt.Sprintf("%s: %v", msg, c.Request.Pattern))
}
