package routing

import (
	"log"
	"net/http"
	"strings"
)

type RouteGroup[T any] struct {
	*Router[T]      // Embedded
	Prefix          string
	HandlerWrappers []HandlerWrapper // Group Handler Wrappers
}

// Handle registers a route pattern
func (g *RouteGroup[T]) Handle(subpattern string, handler http.Handler, handlerWrappers ...HandlerWrapper) {
	var (
		subPatternParts []string
		subpath         string
		method          string
		fullPattern     string
	)

	subPatternParts = strings.SplitN(subpattern, " ", 2)
	if len(subPatternParts) == 2 {
		// subpattern "<method> <subpath>" -> fullpattern "<method> <groupPrefix><subpath>"
		// method: e.g. GET, POST
		method = subPatternParts[0]
		subpath = subPatternParts[1]
		fullPattern = method + " " + g.Prefix + subpath
	} else {
		fullPattern = g.Prefix + subpattern
	}

	if strings.Contains(fullPattern, "//") {
		log.Fatalf("[ERROR] Can't Register Route Pattern %s", fullPattern)
	}

	// Wrapping the Handler (Nesting) by the HandlerWrappers into the Actual Handler
	// Wrapped Handler = grpHndWrapr1 (
	//						...
	//						grpHndWraprN (
	//							hndWrapr1 (
	//								...
	//								hndWraprN (
	//									handler
	//								)
	// 							)
	//						)
	//					)
	// 1. Pre-action order:
	//		grpHndWrapr1 -> ... -> grpHndWraprN -> hndWrapr1 -> ... -> hndWraprN
	// 2. handler.ServeHTTP(w,r)
	// 3. Post-action order:
	//		grpHndWrapr1 <- ... <- grpHndWraprN <- hndWrapr1 <- ... <- hndWraprN
	wrappedHandler := handler
	for i := len(handlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = handlerWrappers[i].Wrap(wrappedHandler)
	}
	for i := len(g.HandlerWrappers) - 1; i >= 0; i-- {
		wrappedHandler = g.HandlerWrappers[i].Wrap(wrappedHandler)
	}
	// Register the fullPattern with the WrappedHandler
	g.Router.ServeMux.Handle(fullPattern, wrappedHandler)
}

func (g *RouteGroup[T]) HandleFunc(subpattern string, handleFunc func(http.ResponseWriter, *http.Request), handlerWrappers ...HandlerWrapper) {
	g.Handle(subpattern, http.HandlerFunc(handleFunc), handlerWrappers...)
}

// Group on *RouteGroup makes a Subgroup
//
//	router.Group("/foo/", func(foo *RouteGroup) {        // RouteGroup for "/foo/..."
//	  foo.Handle("GET bar", foobarGetHandler)            // "GET /foo/bar"
//
//	  foo.Group("baz/", func(foobaz *RouteGroup) {		 // RouteGroup for "/foo/baz/..." = Subgroup of "/foo/"
//	    foobaz.Handle("GET baas", foobazbaasGetHandler)  // "GET /foo/baz/baas"
//	    foobaz.Handle("POST bam", foobazbamPostHandler)  // "POST /foo/baz/bam"
//	  }
//	}
func (g *RouteGroup[T]) Group(subPrefix string, batch func(*RouteGroup[T]), handlerWrappers ...HandlerWrapper) *RouteGroup[T] {
	rg := &RouteGroup[T]{
		Prefix:          g.Prefix + subPrefix,                          // extended prefix
		Router:          g.Router,                                      // same router
		HandlerWrappers: append(g.HandlerWrappers, handlerWrappers...), // handlerwrappers appended
	}

	batch(rg)

	return rg // to do more with this routegroup if any
}
