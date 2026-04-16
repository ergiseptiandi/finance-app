package routeinfo

type RouteInfo struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Summary   string `json:"summary"`
	Protected bool   `json:"protected"`
}
