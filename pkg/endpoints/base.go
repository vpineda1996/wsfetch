package endpoints

const (
	RootHostname  = "wealthsimple.com"
	MyWeathSimple = "https://my." + RootHostname

	Api = "https://api.production." + RootHostname

	Trade       = "https://trade-service." + RootHostname
	TradePublic = Trade + "/public"

	Status = "https://status." + RootHostname
)
