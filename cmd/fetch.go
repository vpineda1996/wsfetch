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
	"github.com/vpnda/wsfetch/pkg/client/generated"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
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
			c = lo.Must(client.NewClient(ctx, authClient))
		} else {
			fmt.Println("Loaded session from file")
			authClient := base.AuthClientFromSession(session)
			serializeSession(lo.Must(authClient.Fetcher.GetSession(ctx)))
			c = lo.Must(client.NewClient(ctx, authClient))
		}

		accounts := lo.Must(c.GetAccounts(ctx))
		accountIds := lo.Map(accounts, func(a generated.Account, _ int) client.AccountId {
			return client.AccountId(a.Id)
		})
		for _, account := range accounts {
			fmt.Printf("Account: %s, ID: %s\n", *account.Type, account.Id)
		}
		trns := lo.Must(c.Transactions(ctx, accountIds, time.Now().Add(-30*time.Hour*24), lo.ToPtr(time.Now())))
		prettyPrint(trns)
	},
}

const (
	sessionFile = "session.json"
)

func prettyPrint(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Println("Failed to marshal JSON:", err)
		return
	}
	fmt.Println(string(b))
}

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
