package inertia

import (
	"bytes"
	"sync"

	"github.com/gofiber/fiber/v2"
)

const (
	HeaderXInertia                 = "X-Inertia"
	HeaderXInertiaVersion          = "X-Inertia-Version"
	HeaderXInertiaLocation         = "X-Inertia-Location"
	HeaderXInertiaPartialData      = "X-Inertia-Partial-Data"
	HeaderXInertiaPartialComponent = "X-Inertia-Partial-Component"
)

type Inertia struct {
	c             *fiber.Ctx
	rootView      string
	sharedProps   map[string]interface{}
	version       VersionFunc
	renderer      Renderer
	isSsrDisabled bool
	mu            sync.RWMutex
}

func (i *Inertia) SetRenderer(r Renderer) {
	i.renderer = r
}

func (i *Inertia) Renderer() Renderer {
	return i.renderer
}

func (i *Inertia) IsSsrDisabled() bool {
	return i.isSsrDisabled
}

func (i *Inertia) IsSsrEnabled() bool {
	return !i.isSsrDisabled
}

func (i *Inertia) EnableSsr() {
	i.isSsrDisabled = false
}

func (i *Inertia) DisableSsr() {
	i.isSsrDisabled = true
}

func (i *Inertia) SetRootView(name string) {
	i.rootView = name
}

func (i *Inertia) RootView() string {
	return i.rootView
}

func (i *Inertia) Share(props map[string]interface{}) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedProps = make(map[string]interface{})

	// merge shared props
	for k, v := range props {
		i.sharedProps[k] = v
	}
}

func (i *Inertia) Shared() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.sharedProps
}

func (i *Inertia) FlushShared() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.sharedProps = map[string]interface{}{}
}

type VersionFunc func() string

func (i *Inertia) SetVersion(version VersionFunc) {
	i.version = version
}

func (i *Inertia) Version() string {
	return i.version()
}

// Location generates 409 response for external redirects
// see https://inertiajs.com/redirects#external-redirects
func (i *Inertia) Location(url string, code int) error {
	if i.c.Get(HeaderXInertia) != "" {
		res := i.c.Response()
		res.Header.Set(HeaderXInertiaLocation, url)
		res.Header.SetStatusCode(fiber.StatusConflict)
		return nil
	}
	return i.c.Redirect(url, code)
}

func (i *Inertia) Render(code int, component string, props map[string]interface{}) error {
	return i.render(code, component, props, map[string]interface{}{})
}

func (i *Inertia) RenderWithViewData(code int, component string, props, viewData map[string]interface{}) error {
	return i.render(code, component, props, viewData)
}

type Page struct {
	Component string                 `json:"component"`
	Props     map[string]interface{} `json:"props"`
	URL       string                 `json:"url"`
	Version   string                 `json:"version"`
}

func (i *Inertia) render(code int, component string, props, viewData map[string]interface{}) error {

	res := i.c.Response()

	props = mergeProps(i.sharedProps, props)

	only := splitAndRemoveEmpty(i.c.Get(HeaderXInertiaPartialData), ",")
	if len(only) > 0 && i.c.Get(HeaderXInertiaPartialComponent) == component {
		filteredProps := map[string]interface{}{}
		for _, key := range only {
			filteredProps[key] = props[key]
		}
		props = filteredProps
	} else {
		filteredProps := map[string]interface{}{}
		for key, prop := range props {
			// LazyProp is only used in partial reloads
			// see https://inertiajs.com/partial-reloads#lazy-data-evaluation
			if _, ok := prop.(*LazyProp); !ok {
				filteredProps[key] = prop
			}
		}
		props = filteredProps
	}

	if err := evaluateProps(props); err != nil {
		return err
	}

	page := &Page{
		Component: component,
		Props:     props,
		URL:       i.c.BaseURL() + i.c.OriginalURL(),
		Version:   i.Version(),
	}

	res.Header.Set("Vary", HeaderXInertia)

	if i.c.Get(HeaderXInertia) != "" {
		res.Header.Set(HeaderXInertia, "true")
		return i.c.Status(code).JSON(page)
	}

	viewData["page"] = page

	return i.renderHTML(code, i.rootView, viewData)
}

// renderHTML renders HTML template with given code, name and data.
func (i *Inertia) renderHTML(code int, name string, data map[string]interface{}) error {
	if i.renderer == nil {
		return ErrRendererNotRegistered
	}
	buf := new(bytes.Buffer)
	if err := i.renderer.Render(buf, name, data, i); err != nil {
		return err
	}

	i.c.Set("Content-type", "text/html; charset=UTF-8")
	return i.c.Status(code).Send(buf.Bytes())
}

type LazyPropFunc func() (interface{}, error)

type LazyProp struct {
	callback LazyPropFunc
}

// Lazy defines a lazy evaluated data.
// see https://inertiajs.com/partial-reloads#lazy-data-evaluation
func Lazy(callback LazyPropFunc) *LazyProp {
	return &LazyProp{
		callback: callback,
	}
}

func SetRootView(c *fiber.Ctx, name string) {
	MustGet(c).SetRootView(name)
}

func RootView(c *fiber.Ctx) string {
	return MustGet(c).RootView()
}

func Share(c *fiber.Ctx, props map[string]interface{}) {
	MustGet(c).Share(props)
}

func Shared(c *fiber.Ctx) map[string]interface{} {
	return MustGet(c).Shared()
}

func FlushShared(c *fiber.Ctx) {
	MustGet(c).FlushShared()
}

func SetVersion(c *fiber.Ctx, version VersionFunc) {
	MustGet(c).SetVersion(version)
}

func Version(c *fiber.Ctx) string {
	return MustGet(c).Version()
}

func Location(c *fiber.Ctx, url string, code int) error {
	return MustGet(c).Location(url, code)
}

func Render(c *fiber.Ctx, code int, component string, props map[string]interface{}) error {
	return MustGet(c).Render(code, component, props)
}

func RenderWithViewData(c *fiber.Ctx, code int, component string, props, viewData map[string]interface{}) error {
	return MustGet(c).RenderWithViewData(code, component, props, viewData)
}

func Redirect(c *fiber.Ctx, url string, props map[string]interface{}) error {
	return MustGet(c).Redirect(url, props)
}

func (i *Inertia) Redirect(url string, props map[string]interface{}) error {
	return i.c.Redirect(url, fiber.StatusFound)
}

func RedirectToRoute(c *fiber.Ctx, routeName string, props map[string]interface{}) error {
	return MustGet(c).RedirectToRoute(routeName, props)
}

func (i *Inertia) RedirectToRoute(routeName string, props map[string]interface{}) error {
	return i.c.RedirectToRoute(routeName, props, fiber.StatusFound)
}
