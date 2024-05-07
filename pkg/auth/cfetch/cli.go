package cfetch

import (
	"fmt"
	"io"
	"os"

	"github.com/vpineda1996/wsfetch/pkg/auth/types"
)

type cli struct {
	out io.Writer
	in  io.Reader
}

func NewCli() *cli {
	return &cli{
		out: os.Stdout,
		in:  os.Stdin,
	}
}

func (c *cli) Fetch(twoFaInfo types.TwoFactorAuthRequest) (string, error) {
	fmt.Fprintf(c.out, "Wealthsimple requesting code through %s for claim %s...\n", twoFaInfo.Method, twoFaInfo.AuthenticatedClaim[:10])
	fmt.Fprintf(c.out, "Code: ")
	var result string
	_, err := fmt.Fscanf(c.in, "%s", &result)
	if err != nil {
		return "", err
	}
	return result, nil
}
