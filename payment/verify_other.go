package payment

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// isValidBIC validates the BIC format (a simple regex for general BIC checks).
func isValidBIC(bic string) bool {
	bicRegex := `^[A-Z]{4}[-]{0,1}[A-Z]{2}[-]{0,1}[A-Z0-9]{2}[-]{0,1}[0-9]{3}$`
	matched, _ := regexp.MatchString(bicRegex, bic)
	return matched
}

// formatAmount ensures the amount is properly formatted with a dot as the decimal separator and two decimal places.
func formatAmount(amount string, currency string) string {
	if amount == "" {
		return ""
	}

	// Replace commas with dots for decimal separation.
	amount = strings.ReplaceAll(amount, ",", ".")

	// Ensure the amount has up to two decimal places.
	re := regexp.MustCompile(`^\d+(\.\d{1,2})?`)
	match := re.FindString(amount)
	if match == "" {
		return ""
	}

	// Parse the amount as a float to ensure proper formatting.
	parsedAmount, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return ""
	}

	// Format the amount to two decimal places and include the currency symbol.
	return fmt.Sprintf("%s%.2f", currency, parsedAmount)
}
