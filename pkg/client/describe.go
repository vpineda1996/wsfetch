package client

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/vpnda/wsfetch/pkg/client/generated"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	outflowActities = []generated.ActivityType{
		generated.ActivityTypeWithdrawal,
		generated.ActivityTypeDiyBuy,
		generated.ActivityTypeManagedBuy,
		generated.ActivityTypeP2pPayment,
	}
	inflowActivities = []generated.ActivityType{
		generated.ActivityTypeDeposit,
		generated.ActivityTypeDiySell,
		generated.ActivityTypeManagedSell,
		generated.ActivityTypeInterest,
		generated.ActivityTypeDividend,
		generated.ActivityTypeRefund,
	}
)

// GetFormattedAmount returns a formatted amount for the given activity
// negative are considered income, positive are considered outflows or expenses
func GetFormattedAmount(act *generated.Activity) string {
	prefix := ""
	// We have to introspectively check the type of the account
	// as the hints from wealthsimple aren't great

	if lo.Contains(outflowActities, act.Type) {
		// DO NOTHING
	} else if lo.Contains(inflowActivities, act.Type) ||
		act.AmountSign == generated.AmountSignNegative {
		prefix = "-"
	}
	return fmt.Sprintf("%s%s", prefix, act.Amount)
}

// GetActivityDescription returns a description for the given activity
func GetActivityDescription(ctx context.Context, c Client, act *generated.Activity) (string, error) {
	baseDescription := fmt.Sprintf("%s: %s", capitalizeEnum(string(act.Type)), capitalizeEnum(string(act.SubType)))

	switch act.Type {
	case generated.ActivityTypeInternalTransfer:
		return internalTransferDescription(ctx, c, act)
	case generated.ActivityTypeDiyBuy, generated.ActivityTypeDiySell,
		generated.ActivityTypeManagedBuy, generated.ActivityTypeManagedSell:
		return securityActivityDescription(ctx, c, act)
	case generated.ActivityTypeInterest:
		if act.SubType == generated.ActivitySubtypeFplInterest {
			return "Stock Lending Earnings", nil
		} else {
			return "Interest", nil
		}
	case generated.ActivityTypeDividend:
		security, err := findActivitySymbol(ctx, c, act)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("Dividend: %s", security), nil
	case generated.ActivityTypeFundsConversion:
		fromCurrency := "USD"
		if *act.Currency == "CAD" {
			fromCurrency = "USD"
		} else {
			fromCurrency = "CAD"
		}
		return fmt.Sprintf("Funds converted: %s from %s", *act.Currency, fromCurrency), nil
	case generated.ActivityTypeNonResidentTax:
		return "Non-resident tax", nil
	case generated.ActivityTypeCryptoTransfer:
		return cryptoActivityDescription(act)
	case generated.ActivityTypeWithdrawal, generated.ActivityTypeDeposit:
		return handleDepositWithdrawal(act)
	default:
	}

	if act.Type == "REFUND" && act.SubType == "TRANSFER_FEE_REFUND" {
		return "Reimbursement: account transfer fee", nil
	} else if act.Type == "INSTITUTIONAL_TRANSFER_INTENT" && act.SubType == "TRANSFER_IN" {
		return "Institutional transfer in", nil
	} else if act.Type == "P2P_PAYMENT" && (act.SubType == "SEND" || act.SubType == "SEND_RECEIVED") {
		direction := "sent to"
		if act.SubType == "SEND_RECEIVED" {
			direction = "received from"
		}
		return fmt.Sprintf("Cash %s %s", direction, *act.P2pHandle), nil
	}

	return baseDescription, nil
}

func handleDepositWithdrawal(act *generated.Activity) (string, error) {
	verb := capitalizeEnum(string(act.Type))
	action := capitalizeEnum(string(act.SubType))
	description := ""

	direction := "from"
	if act.Type == generated.ActivityTypeWithdrawal {
		direction = "to"
	}

	switch act.SubType {
	case generated.ActivitySubtypeBillPay:
		payeeName := *act.BillPayCompanyName
		if act.BillPayPayeeNickname != nil && *act.BillPayPayeeNickname != "" {
			payeeName = *act.BillPayPayeeNickname
		}
		var accountNumber string
		if act.RedactedExternalAccountNumber != nil {
			accountNumber = fmt.Sprintf(" (%s)", *act.RedactedExternalAccountNumber)
		}
		description = fmt.Sprintf("%s %s%s", direction, payeeName, accountNumber)
	case generated.ActivitySubtypeAft:
		action = "Direct deposit"
		if act.Type == generated.ActivityTypeWithdrawal {
			action = "Pre-authorized debit"
		}
		description = fmt.Sprintf("%s %s", direction, *act.AftOriginatorName)
	case generated.ActivitySubtypeETransfer, generated.ActivitySubtypeETransferFunding:
		action = "Interac e-transfer"
		description = fmt.Sprintf("%s %s (%s)", direction, *act.ETransferName, *act.ETransferEmail)
	case generated.ActivitySubtypePaymentCardTransaction:
		action = "Debit card transaction"
	case generated.ActivitySubtypeEft:
		action = "EFT"
	}
	return fmt.Sprintf("%s: %s %s", verb, action, description), nil

}

func cryptoActivityDescription(act *generated.Activity) (string, error) {
	symbol := SecuritySymbol(*act.AssetSymbol)
	if act.SubType == generated.ActivitySubtypeTransferIn {
		return fmt.Sprintf("Transfer in: Crypto transfer in: %s %s", act.AssetQuantity, symbol), nil
	} else {
		return fmt.Sprintf("Transfer oun: Crypto transfer in: %s %s", act.AssetQuantity, symbol), nil
	}
}

func findActivitySymbol(ctx context.Context, c Client, act *generated.Activity) (SecuritySymbol, error) {
	if act.AssetSymbol != nil && *act.AssetSymbol != "" {
		return SecuritySymbol(*act.AssetSymbol), nil
	}
	securityInfo, err := c.GetSecurityMarketData(ctx, *act.SecurityId)
	if err != nil {
		return "", err

	}
	return SecuritySymbolFromMarketData(securityInfo)
}

func securityActivityDescription(ctx context.Context, c Client, act *generated.Activity) (string, error) {
	verb := capitalizeEnum(string(act.SubType))
	if verb == "" {
		verb = capitalizeEnum(string(act.Type))
	}
	action := "buy"
	if act.Type == generated.ActivityTypeDiySell || act.Type == generated.ActivityTypeManagedSell {
		action = "sell"
	}

	security, err := findActivitySymbol(ctx, c, act)
	if err != nil {
		return "", err
	}

	assetQuantity, _ := strconv.ParseFloat(act.AssetQuantity, 64)
	amount, _ := strconv.ParseFloat(act.Amount, 64)
	price := amount / assetQuantity

	return fmt.Sprintf("%s: %s %g x %s @ %0.2f", verb, action, assetQuantity, security, price), nil
}

func internalTransferDescription(ctx context.Context, c Client, act *generated.Activity) (string, error) {
	targetAccount, err := c.GetAccount(ctx, *act.OpposingAccountId)
	if err != nil {
		return "", err
	}

	accountDescription := *act.OpposingAccountId
	if targetAccount != nil {
		// try to get the description from nickname first
		if targetAccount.Nickname != nil && *targetAccount.Nickname != "" {
			accountDescription = *targetAccount.Nickname
		} else if targetAccount.UnifiedAccountType != nil {
			// use the formattied name of the account type
			accountDescription = capitalizeEnum(string(*targetAccount.UnifiedAccountType))
		}
	}

	if act.SubType == generated.ActivitySubtypeSource {
		return fmt.Sprintf("Transfer out: Transfer to Wealthsimple %s", accountDescription), nil
	} else {
		return fmt.Sprintf("Transfer in: Transfer from Wealthsimple %s", accountDescription), nil
	}
}

func capitalizeEnum(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.English).String(s)
	return s
}
