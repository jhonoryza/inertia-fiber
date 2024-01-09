package inertia

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	key = "__inertia__"
)

type MiddlewareConfig struct {
	Next func(c *fiber.Ctx) bool
	// The root template that's loaded on the first page visit.
	// see https://inertiajs.com/server-side-setup#root-template
	RootView string
	// Determines the current asset version.
	// see https://inertiajs.com/asset-versioning
	VersionFunc func() string
	// Defines the props that are shared by default.
	// see https://inertiajs.com/shared-data
	Share SharedDataFunc
	// Renderer is a renderer that is used for rendering the root view.
	Renderer Renderer
	// IsSsrDisabled is a flag that determines whether server-side rendering is disabled.
	IsSsrDisabled bool
}

type SharedDataFunc func(c *fiber.Ctx) (map[string]interface{}, error)

var DefaultMiddlewareConfig = MiddlewareConfig{
	Next:          nil,
	RootView:      "app.html",
	VersionFunc:   defaultVersionFunc(),
	Share:         nil,
	Renderer:      nil,
	IsSsrDisabled: false,
}

func defaultVersionFunc() VersionFunc {
	var v string

	// It is for Google App Engine.
	// see https://cloud.google.com/appengine/docs/standard/go/runtime#environment_variables
	if v = os.Getenv("GAE_VERSION"); v == "" {
		// The fallback version value that imitates the default GAE version format.
		// It assumes to be used for development.
		v = time.Now().Format("20060102t150405")
	}

	return func() string {
		return v
	}
}

func MustGet(c *fiber.Ctx) *Inertia {
	in, err := Get(c)
	if err != nil {
		panic(err)
	}
	return in
}

func Get(c *fiber.Ctx) (*Inertia, error) {
	in, ok := c.Locals(key).(*Inertia)
	if !ok {
		return nil, ErrNotFound
	}
	return in, nil
}

func Middleware(r Renderer) fiber.Handler {
	return MiddlewareWithConfig(MiddlewareConfig{
		Renderer: r,
	})
}

func MiddlewareWithConfig(config MiddlewareConfig) fiber.Handler {
	// Defaults
	if config.Next == nil {
		config.Next = DefaultMiddlewareConfig.Next
	}
	if config.RootView == "" {
		config.RootView = DefaultMiddlewareConfig.RootView
	}
	if config.VersionFunc == nil {
		config.VersionFunc = DefaultMiddlewareConfig.VersionFunc
	}

	return func(c *fiber.Ctx) error {
		if config.Next != nil && config.Next(c) {
			return c.Next()
		}

		var sharedProps map[string]interface{}
		in := &Inertia{
			c:             c,
			rootView:      config.RootView,
			sharedProps:   sharedProps,
			version:       config.VersionFunc,
			renderer:      config.Renderer,
			isSsrDisabled: config.IsSsrDisabled,
		}

		c.Locals(key, in)

		return c.Next()
	}

}
