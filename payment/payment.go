package payment

import (
	"net/http"
	"strings"

	"github.com/Mr-Comand/qrCodeGeneratorAPI/log"
)

// Allowed currencies as a slice (initially populated)
var allowedCurrencies = []string{"EUR", "USD", "GBP", "JPY", "AUD", "CAD"}

// Default currency (initially set to EUR)
var defaultCurrency = "EUR"

// SetAllowedCurrencies allows you to set the list of allowed currencies
func SetAllowedCurrencies(currencies []string) {
	allowedCurrencies = currencies
}

// SetDefaultCurrency allows you to set the default currency
func SetDefaultCurrency(currency string) {
	// Check if the currency is valid, else set to a fallback (e.g., EUR)
	isValidCurrency := false
	for _, allowedCurrency := range allowedCurrencies {
		if currency == allowedCurrency {
			isValidCurrency = true
			break
		}
	}

	if isValidCurrency {
		defaultCurrency = currency
	} else {
		log.LogErrorString("SetDefaultCurrency", "Invalid default currency: "+currency+". Defaulting to EUR.")
		defaultCurrency = "EUR"
	}
}

// generatePaymentQRCode handles the SEPA QR code generation with validation.
func GeneratePaymentQRCode(r *http.Request) (string, string) {
	method := r.URL.Query().Get("method")
	if method != "sepa" {
		log.LogInfo("GeneratePaymentQRCode", "Invalid method: <"+method+">. Method should be 'sepa'.")
		return "", "Only SEPA payment method is supported"
	}

	// Required fields with validation
	iban := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("iban")))
	if !isValidIBAN(iban) {
		log.LogInfo("GeneratePaymentQRCode", "Invalid IBAN format: <"+iban+">")
		return "", "Invalid IBAN format"
	}

	bic := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("bic")))
	if bic != "" && !isValidBIC(bic) {
		log.LogInfo("GeneratePaymentQRCode", "Invalid BIC format: <"+bic+">")
		return "", "Invalid BIC format"
	}

	name := strings.TrimSpace(r.URL.Query().Get("name"))
	if len(name) == 0 || len(name) > 70 {
		log.LogInfo("GeneratePaymentQRCode", "Invalid or missing 'name' parameter (max 70 characters): <"+name+">")
		return "", "Invalid or missing 'name' parameter (max 70 characters)"
	}

	amountParam := r.URL.Query().Get("amount")
	currencyParam := r.URL.Query().Get("currency")
	// If no currency is provided, default to "EUR"
	if currencyParam == "" {
		currencyParam = defaultCurrency
		log.LogDebug("GeneratePaymentQRCode", "'currency' parameter not provieded fallback to "+defaultCurrency+".")
	}

	// Check if the currency is in the allowed list
	isValidCurrency := false
	for _, allowedCurrency := range allowedCurrencies {
		if currencyParam == allowedCurrency {
			isValidCurrency = true
			break
		}
	}

	if !isValidCurrency {
		log.LogInfo("GeneratePaymentQRCode", "Invalid 'currency' parameter: <"+currencyParam+">")
		return "", "Invalid 'currency' parameter"
	}

	amount := formatAmount(amountParam, currencyParam)
	if amount == "" {
		log.LogInfo("GeneratePaymentQRCode", "Invalid 'amount' parameter: <"+amountParam+">")
		return "", "Invalid 'amount' parameter"
	}

	purpose := strings.TrimSpace(r.URL.Query().Get("purpose"))
	if len(purpose) > 4 {
		log.LogInfo("GeneratePaymentQRCode", "Invalid 'purpose' parameter (max 4 characters): <"+purpose+">")
		return "", "Invalid 'purpose' parameter (max 4 characters)"
	}

	reference := strings.TrimSpace(r.URL.Query().Get("reference"))
	remittanceText := strings.TrimSpace(r.URL.Query().Get("remittance"))
	information := strings.TrimSpace(r.URL.Query().Get("information"))

	// Optional fields validation
	if reference != "" && len(reference) > 25 {
		log.LogInfo("GeneratePaymentQRCode", "Invalid 'reference' parameter (max 25 characters): <"+reference+">")
		return "", "Invalid 'reference' parameter (max 25 characters)"
	}

	if remittanceText != "" && len(remittanceText) > 140 {
		log.LogInfo("GeneratePaymentQRCode", "Invalid 'remittance' parameter (max 140 characters): <"+remittanceText+">")
		return "", "Invalid 'remittance' parameter (max 140 characters)"
	}

	// Build the SEPA QR Code
	qrData := buildSEPAQRCodeData(iban, bic, name, amount, purpose, reference, remittanceText, information)

	// Log successful QR code generation
	log.LogInfo("GeneratePaymentQRCode", "SEPA QR Code Data generated successfully for name: <"+name+">")

	return qrData, "" // Return the QR code data and no error message
}

// buildSEPAQRCodeData assembles the SEPA QR code data following the format
func buildSEPAQRCodeData(iban, bic, name, amount, purpose, reference, remittanceText, information string) string {
	// Service Tag
	var sb strings.Builder
	sb.WriteString("BCD\n")

	// Version (EWR)
	sb.WriteString("002\n")

	// Character Set
	sb.WriteString("2\n")

	// Identification (SCT)
	sb.WriteString("SCT\n")

	// BIC (only if version 001)
	sb.WriteString(bic + "\n")

	// Name
	sb.WriteString(name + "\n")

	// IBAN
	sb.WriteString(iban + "\n")

	// Amount
	sb.WriteString(amount + "\n")

	// Purpose (max 4 characters)
	sb.WriteString(purpose + "\n")

	// Reference and Remittance Text
	if reference != "" {
		sb.WriteString(reference + "\n") // Reference
		sb.WriteString("\n")             // Empty remittance text
	} else {
		sb.WriteString("\n")                  // Empty reference
		sb.WriteString(remittanceText + "\n") // Remittance text
	}

	// Information (no newline at the end)
	sb.WriteString(information)

	return sb.String()
}
