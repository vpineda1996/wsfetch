package endpoints

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
	MyWealthsimpleLoginSplash = Route{"GET", MyWeathSimple, "/app/login"}
	MyWealthsimpleSession     = Route{"GET", MyWeathSimple, "/api/sessions"}

	AuthToken = Route{"POST", Api, "/v1/oauth/v2/token"}
)
