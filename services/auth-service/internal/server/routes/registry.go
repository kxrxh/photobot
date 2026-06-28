package routes

import "github.com/gofiber/fiber/v3"

type Route struct {
	Method     string
	Path       string
	Handler    fiber.Handler
	Middleware []fiber.Handler
}

type Option func(*Route)

type Registry struct {
	items         *[]Route
	currentPrefix string
	groupOptions  []Option
}

func NewRegistry() *Registry {
	items := []Route{}
	return &Registry{items: &items, groupOptions: nil}
}

func (r *Registry) Group(prefix string, opts ...Option) *Registry {
	combined := make([]Option, len(r.groupOptions)+len(opts))
	copy(combined, r.groupOptions)
	copy(combined[len(r.groupOptions):], opts)
	return &Registry{
		items:         r.items,
		currentPrefix: r.currentPrefix + prefix,
		groupOptions:  combined,
	}
}

func (r *Registry) add(method, path string, handler fiber.Handler, opts ...Option) *Registry {
	rt := Route{Method: method, Path: r.currentPrefix + path, Handler: handler}
	for _, o := range r.groupOptions {
		o(&rt)
	}
	for _, o := range opts {
		o(&rt)
	}
	*r.items = append(*r.items, rt)
	return r
}

func (r *Registry) GET(path string, handler fiber.Handler, opts ...Option) *Registry {
	return r.add(fiber.MethodGet, path, handler, opts...)
}

func (r *Registry) POST(path string, handler fiber.Handler, opts ...Option) *Registry {
	return r.add(fiber.MethodPost, path, handler, opts...)
}

func (r *Registry) PUT(path string, handler fiber.Handler, opts ...Option) *Registry {
	return r.add(fiber.MethodPut, path, handler, opts...)
}

func (r *Registry) DELETE(path string, handler fiber.Handler, opts ...Option) *Registry {
	return r.add(fiber.MethodDelete, path, handler, opts...)
}

func (r *Registry) PATCH(path string, handler fiber.Handler, opts ...Option) *Registry {
	return r.add(fiber.MethodPatch, path, handler, opts...)
}

func (r *Registry) List() []Route {
	return *r.items
}
