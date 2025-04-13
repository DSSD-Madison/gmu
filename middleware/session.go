package middleware

import (
	"github.com/gorilla/sessions"
	"net/http"

	"github.com/labstack/echo/v4"
)

var Store = sessions.NewCookieStore([]byte("very-secret-key-keep-it-safe"))

func init() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   60,         
		HttpOnly: true,
		Secure:   false,              // ✅ VERY IMPORTANT for localhost
		SameSite: http.SameSiteLaxMode, // Safe default, allows HTMX redirects
	}
}

func IsMaster(c echo.Context) bool {
	session, _ := Store.Get(c.Request(), "session")
	isMaster, ok := session.Values["is_master"].(bool)
	return ok && isMaster
}

// Middleware to check auth
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, _ := Store.Get(c.Request(), "session")
		auth, ok := session.Values["authenticated"].(bool)

		if !ok || !auth {
			redirectTo := c.Request().RequestURI
			// If HTMX, use HX-Redirect
			if c.Request().Header.Get("HX-Request") == "true" {
				c.Response().Header().Set("HX-Redirect", "/login?redirect="+redirectTo)
				return c.NoContent(http.StatusUnauthorized)
			}
			return c.Redirect(http.StatusSeeOther, "/login?redirect="+redirectTo)
		}

		return next(c)
	}
}

