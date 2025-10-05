package routing

import "net/http"

type Router[T any] struct {
	*http.ServeMux
	Env *T
}

// Handle registers a route pattern
func (router *Router[T]) Handle(pattern string, handler http.Handler, handlerWrappers ...HandlerWrapper) {
	wrappedHandler := handler
	for i := len(handlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = handlerWrappers[i].Wrap(wrappedHandler)
	}
	router.ServeMux.Handle(pattern, wrappedHandler)
}

func (router *Router[T]) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper) {
	router.Handle(pattern, http.HandlerFunc(handleFunc), handlerWrappers...)
}

// Group lets you register routes under a common Prefix + middleware.
func (router *Router[T]) Group(prefix string, batch func(*RouteGroup[T]), handlerWrappers ...HandlerWrapper) *RouteGroup[T] {
	rg := &RouteGroup[T]{
		Prefix:          prefix,
		Router:          router,
		HandlerWrappers: handlerWrappers,
	}

	batch(rg)

	return rg // to do more with this routegroup if any
}
