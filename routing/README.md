# Routing System

## - Aliasing with Type Instantiation

```
type Router = routing.BaseRouter[Env]

type RouteGroup = routing.RouteGroup[Env]
```

where `Env` is the Concrete Type in your application.

## - Overriding Methods by Embedding

```
type Router struct {
	routing.BaseRouter[Env]
}
```

still, your RouteGroup can be an alias with type instantiation:

```
type RouteGroup = routing.RouteGroup[Env]
```

