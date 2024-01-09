package inertia

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/utils"
)

type CSRFConfig csrf.Config

var DefaultCSRFConfig = CSRFConfig{
	Next:           nil,
	KeyLookup:      "header:X-XSRF-TOKEN",
	CookieName:     "XSRF-TOKEN",
	CookieSameSite: "samesite",
	CookiePath:     "/",
	Expiration:     1 * time.Hour,
	KeyGenerator:   utils.UUIDv4,
}

func CSRF() fiber.Handler {
	c := DefaultCSRFConfig
	return CSRFWithConfig(c)
}

func CSRFWithConfig(config CSRFConfig) fiber.Handler {
	if config.Next == nil {
		config.Next = DefaultCSRFConfig.Next
	}

	if config.ContextKey == "" {
		config.ContextKey = DefaultCSRFConfig.ContextKey
	}
	if config.CookieName == "" {
		config.CookieName = DefaultCSRFConfig.CookieName
	}

	if config.CookieSameSite == "samesite" {
		config.CookieSecure = true
	}
	if config.CookiePath == "" {
		config.CookiePath = DefaultCSRFConfig.CookiePath
	}

	return func(c *fiber.Ctx) error {
		if config.Next != nil && config.Next(c) {
			return c.Next()
		}

		csrf.New(csrf.Config(config))

		return c.Next()
	}
}
