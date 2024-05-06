package endpoints

type Route struct {
	Method string
	Host   string
	Path   string
}

func (r Route) String() string {
	return r.Host + r.Path
}

var (
	// Auth Paths
	LoginSplash = Route{"GET", MyWeathSimple, "/app/login"}
	AuthToken   = Route{"POST", Api, "/v1/oauth/v2/token"}
)
