package endpoints

import "net/http"

type Route struct {
	Method string
	Host   string
	Path   string
}

func (r Route) String() string {
	return "https://" + r.Host + r.Path
}

var (
	// Auth Paths
	MyWealthsimpleLoginSplash = Route{http.MethodGet, MyWeathSimple, "/app/login"}
	MyWealthsimpleSession     = Route{http.MethodPost, MyWeathSimple, "/api/sessions"}
	MyWealthsimpleGetGraphQl  = Route{http.MethodGet, MyWeathSimple, "/graphql"}

	AuthToken     = Route{http.MethodPost, Api, "/v1/oauth/v2/token"}
	AuthTokenInfo = Route{http.MethodGet, Api, "/v1/oauth/v2/token/info"}
)
