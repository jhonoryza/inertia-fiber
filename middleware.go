package inertia

import "github.com/gofiber/fiber/v2"

const (
	key = "__inertia__"
)

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
