package cfetch

import (
	"bytes"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/vpnda/wsfetch/pkg/auth/types"
)

func TestCliFetch(t *testing.T) {
	g := NewWithT(t)
	c := &cli{
		out: bytes.NewBuffer([]byte{}),
		in:  bytes.NewBufferString("100\n5\n"),
	}
	result, err := c.Fetch(types.TwoFactorAuthRequest{
		AuthenticatedClaim: "Some Authentication Claim",
	})

	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal("100"))

}
