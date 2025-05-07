/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/vpnda/wsfetch/pkg/auth/types"
	"github.com/vpnda/wsfetch/pkg/base"
	"github.com/vpnda/wsfetch/pkg/client"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches financial data from your brokerage account.",
	Long: `Fetches and displays financial data such as account balances and recent activity
from your brokerage account.

It can authenticate using a username/password or a previously saved session.
If a session exists, it will be used; otherwise, you will be prompted for credentials.
The session will be refreshed and saved for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		fmt.Println("fetch called")

		var c client.Client
		session, err := loadSession(ctx)
		if err != nil {
			fmt.Println("Failed to load session, using password method:", err)
			var username, password string
			fmt.Println("Enter your username:")
			fmt.Scanln(&username)
			fmt.Println("Enter your password:")
			fmt.Scanln(&password)
			authClient := base.DefaultAuthClient(types.PasswordCredentials{
				Username: username,
				Password: password,
			})
			serializeSession(lo.Must(authClient.Fetcher.GetSession(ctx)))
			c = client.NewCachingClient(lo.Must(client.NewClient(ctx, authClient)))
		} else {
			fmt.Println("Loaded session from file")
			authClient := base.AuthClientFromSession(session)
			serializeSession(lo.Must(authClient.Fetcher.GetSession(ctx)))
			c = client.NewCachingClient(lo.Must(client.NewClient(ctx, authClient)))
		}

		accounts := lo.Must(c.GetAccounts(ctx))
		for _, account := range accounts {
			fmt.Printf("Account: %s -- %s\n", account.Id, account.Financials.CurrentCombined.NetLiquidationValueV2.Amount)
			activites := lo.Must(c.GetActivities(ctx,
				[]client.AccountId{client.AccountId(account.Id)}, lo.ToPtr(time.Now().Add(-30*24*time.Hour)), lo.ToPtr(time.Now())))
			for _, activity := range activites[client.AccountId(account.Id)] {
				desc := lo.Must(client.GetActivityDescription(ctx, c, &activity))
				fmt.Printf("%15s $%10s: %s\n", activity.OccurredAt.Format(time.DateOnly), client.GetFormattedAmount(&activity), desc)
			}
		}
	},
}

const (
	sessionFile = "session.json"
)

func serializeSession(sess *types.Session) {
	sessFile, err := os.Create(sessionFile)
	if err != nil {
		fmt.Println("Failed to create session file:", err)
		return
	}
	defer sessFile.Close()
	if err := json.NewEncoder(sessFile).Encode(sess); err != nil {
		fmt.Println("Failed to encode session file:", err)
	}
}

func loadSession(ctx context.Context) (*types.Session, error) {
	var sess *types.Session
	sessFile, err := os.Open(sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open session file: %w", err)
	}
	defer sessFile.Close()
	if err := json.NewDecoder(sessFile).Decode(&sess); err != nil {
		return nil, fmt.Errorf("failed to decode session file: %w", err)
	}
	return sess, nil
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// fetchCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// fetchCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
