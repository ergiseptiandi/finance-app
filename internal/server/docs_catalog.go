package server

import (
	"sort"

	"finance-backend/internal/server/routeinfo"
)

type routeCatalog map[string]routeinfo.RouteInfo

func newRouteCatalog() routeCatalog {
	return routeCatalog{}
}

func (c routeCatalog) Add(route routeinfo.RouteInfo) {
	c[route.Method+" "+route.Path] = route
}

func (c routeCatalog) List() []routeinfo.RouteInfo {
	routes := make([]routeinfo.RouteInfo, 0, len(c))
	for _, route := range c {
		routes = append(routes, route)
	}

	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Path == routes[j].Path {
			return routes[i].Method < routes[j].Method
		}

		return routes[i].Path < routes[j].Path
	})

	return routes
}
