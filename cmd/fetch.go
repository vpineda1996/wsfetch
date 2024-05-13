/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/vpineda1996/wsfetch/pkg/auth/types"
	"github.com/vpineda1996/wsfetch/pkg/base"
	"github.com/vpineda1996/wsfetch/pkg/services/cash"
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

		cl := lo.Must(cash.NewClient(ctx, base.DefaultAuthClient(types.PasswordCredentials{})))

		trns, err := cl.Transactions(ctx, time.Time{}, time.Time{})
		if err != nil {
			panic(err)
		}

		fmt.Println(trns)
	},
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
