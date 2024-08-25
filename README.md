# inertia-fiber

inertia fiber adapter

- [x] dev mode
- [x] build mode
- [x] spa mode
- [ ] ssr mode

## Usage

file `main.go`

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jhonoryza/inertia-fiber"
)

func main {
	router := fiber.New()

	r := inertia.NewRenderer("app")

	r.MustParseGlob("resources/views/*.html")
	r.ViteBasePath = "/build/"
	r.AddViteEntryPoint("resources/js/app.js")
	r.MustParseViteManifestFile("public/build/manifest.json")

	router.Use(inertia.Middleware(r))

	router.Static("/build/assets", "public/build/assets")

	router.Get("home", func(c *fiber.Ctx) error {
			return inertia.Render(c, 200, "Home", fiber.Map{
				"data": "ok",
			})
	})
}
```

file `resources/views/app.html`

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />

    <title inertia>Artefak</title>
    {{- .inertiaHead -}}
  </head>
  <body>
    {{- .inertia -}} {{- vite -}}
  </body>
</html>
```

- run `npm run dev` first
- then `go run main.go`

## License

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.
