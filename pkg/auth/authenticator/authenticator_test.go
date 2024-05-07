package authenticator

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func Test_Client_Authenticate(t *testing.T) {

	// g.Expect(c.Authenticate(ctx, PasswordCredentials{})).ToNot(HaveOccurred())
}

func Test_ParseSessionFromBody(t *testing.T) {
	g := NewWithT(t)
	// setup
	token := "{\"access_token\":\"eyJhbGciOiJSUzI1NiJ9.eyJzd\"," +
		"\"token_type\":\"Bearer\",\"expires_in\":3600," +
		"\"refresh_token\":\"8lahDQ_5ock11Tjvf6tR_pfEEDbddPeaMtTHAo9S4Ds\"," +
		"\"scope\":\"invest.read trade.read tax.read\",\"created_at\":1715053576}"

	sess, err := ParseSessionFromBody([]byte(token))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(sess.BearerToken).ToNot(BeEmpty())
	g.Expect(time.Until(*sess.Expiry)).Should(BeNumerically("~", 3500*time.Second, 3700*time.Second))
}
