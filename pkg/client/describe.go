package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vpineda1996/wsfetch/pkg/client/generated"
)

// GetActivityDescription returns a description for the given activity
func GetActivityDescription(ctx context.Context, c Client, act *generated.Activity) (string, error) {
	baseDescription := fmt.Sprintf("%s: %s", act.Type, act.SubType)

	if act.Type == generated.ActivityTypeInternalTransfer {
		accounts, err := c.GetAccounts(ctx)
		if err != nil {
			return "", err
		}
		var targetAccount *generated.Account
		for _, acc := range accounts {
			if acc.Id == *act.OpposingAccountId {
				targetAccount = &acc
				break
			}
		}

		accountDescription := *act.OpposingAccountId
		if targetAccount != nil {
			accountDescription = targetAccount.GetId()
		}

		if act.SubType == "SOURCE" {
			return fmt.Sprintf("Transfer out: Transfer to Wealthsimple %s", accountDescription), nil
		} else {
			return fmt.Sprintf("Transfer in: Transfer from Wealthsimple %s", accountDescription), nil
		}
	} else if act.Type == "DIY_BUY" || act.Type == "DIY_SELL" || act.Type == "MANAGED_BUY" || act.Type == "MANAGED_SELL" {
		verb := strings.ReplaceAll(string(act.Type), "_", " ")
		action := "buy"
		if act.Type == "DIY_SELL" || act.Type == "MANAGED_SELL" {
			action = "sell"
		}

		security, err := c.SecurityIDToSymbol(ctx, *act.SecurityId)
		if err != nil {
			return "", err
		}
		assetQuantity, _ := strconv.ParseFloat(act.AssetQuantity, 64)
		amount, _ := strconv.ParseFloat(act.Amount, 64)
		price := amount / assetQuantity

		return fmt.Sprintf("%s: %s %g x %s @ %g", verb, action, assetQuantity, security, price), nil
	} else if (act.Type == "DEPOSIT" || act.Type == "WITHDRAWAL") && (act.SubType == "E_TRANSFER" || act.SubType == "E_TRANSFER_FUNDING") {
		direction := "from"
		if act.Type == "WITHDRAWAL" {
			direction = "to"
		}

		return fmt.Sprintf("Deposit: Interac e-transfer %s %s %s", direction, *act.ETransferName, *act.ETransferEmail), nil
	} else if act.Type == "DEPOSIT" && act.SubType == "PAYMENT_CARD_TRANSACTION" {
		return fmt.Sprintf("%s: Debit card funding", act.Type), nil
	} else if act.SubType == "EFT" {
		direction := "from"

		if act.Type == "WITHDRAWAL" {
			direction = "to"
		}

		return fmt.Sprintf("%s: EFT %s", act.Type, direction), nil
	} else if act.Type == "REFUND" && act.SubType == "TRANSFER_FEE_REFUND" {
		return "Reimbursement: account transfer fee", nil
	} else if act.Type == "INSTITUTIONAL_TRANSFER_INTENT" && act.SubType == "TRANSFER_IN" {
		return "Institutional transfer in", nil
	} else if act.Type == "INTEREST" {
		if act.SubType == "FPL_INTEREST" {
			return "Stock Lending Earnings", nil
		} else {
			return "Interest", nil
		}
	} else if act.Type == "DIVIDEND" {
		security, err := c.SecurityIDToSymbol(ctx, *act.SecurityId)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Dividend: %s", security), nil
	} else if act.Type == "FUNDS_CONVERSION" {
		fromCurrency := "USD"
		if *act.Currency == "CAD" {
			fromCurrency = "USD"
		} else {
			fromCurrency = "CAD"
		}

		return fmt.Sprintf("Funds converted: %s from %s", *act.Currency, fromCurrency), nil
	} else if act.Type == "NON_RESIDENT_TAX" {
		return "Non-resident tax", nil
	} else if (act.Type == "DEPOSIT" || act.Type == "WITHDRAWAL") && act.SubType == "AFT" {
		typeStr := "Direct deposit"
		direction := "from"

		if act.Type == "WITHDRAWAL" {
			typeStr = "Pre-authorized debit"
			direction = "to"
		}

		institution := act.ExternalCanonicalId
		if *act.AftOriginatorName != "" {
			institution = act.AftOriginatorName
		}

		return fmt.Sprintf("%s: %s %s", typeStr, direction, *institution), nil
	} else if act.Type == "WITHDRAWAL" && act.SubType == "BILL_PAY" {
		name := act.BillPayCompanyName

		if act.BillPayPayeeNickname != nil && *act.BillPayPayeeNickname != "" {
			name = act.BillPayPayeeNickname
		}

		number := act.RedactedExternalAccountNumber
		return fmt.Sprintf("%s: Bill pay %s %s", act.Type, *name, *number), nil
	} else if act.Type == "P2P_PAYMENT" && (act.SubType == "SEND" || act.SubType == "SEND_RECEIVED") {
		direction := "sent to"
		if act.SubType == "SEND_RECEIVED" {
			direction = "received from"
		}

		return fmt.Sprintf("Cash %s %s", direction, *act.P2pHandle), nil
	}

	return baseDescription, nil
}
