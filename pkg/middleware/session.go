package middleware

import (
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"

	"github.com/labstack/echo/v4"
)

var Store = sessions.NewCookieStore([]byte("very-secret-key-keep-it-safe"))

func init() {
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   10000000,
		HttpOnly: true,
		Secure:   false,                // ✅ VERY IMPORTANT for localhost
		SameSite: http.SameSiteLaxMode, // Safe default, allows HTMX redirects
	}
}

func IsMaster(c echo.Context) bool {
	session, _ := Store.Get(c.Request(), "session")
	isMaster, ok := session.Values["is_master"].(bool)
	return ok && isMaster
}

func IsAuthorized(c echo.Context) bool {
	session, _ := Store.Get(c.Request(), "session")
	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

// Middleware to check auth
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, _ := Store.Get(c.Request(), "session")
		auth, ok := session.Values["authenticated"].(bool)

		if !ok || !auth {
			redirectTo := url.QueryEscape(c.Request().RequestURI)

			// If HTMX request, use HX-Redirect
			if c.Request().Header.Get("HX-Request") == "true" {
				c.Response().Header().Set("HX-Redirect", "/login?redirect="+redirectTo)
				return c.NoContent(http.StatusUnauthorized)
			}

			// Normal browser redirect
			return c.Redirect(http.StatusSeeOther, "/login?redirect="+redirectTo)
		}

		return next(c)
	}
}
