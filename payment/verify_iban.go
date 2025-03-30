package payment

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Mr-Comand/qrCodeGeneratorAPI/log"
)

var countryLengths = map[string]int{
	"DE": 22, // Germany
	"FR": 27, // France
	"GB": 22, // United Kingdom
	"IT": 27, // Italy
	"ES": 24, // Spain
	"NL": 22, // Netherlands
}

// isValidIBAN validates the IBAN format with a stricter check
func isValidIBAN(iban string) bool {
	// Strip all whitespaces and convert to uppercase for uniformity
	iban = strings.ToUpper(strings.ReplaceAll(iban, " ", ""))

	// Regex for IBAN structure validation (basic)
	ibanRegex := `^[A-Z]{2}\d{2}[A-Z0-9]{1,30}$`
	if matched, _ := regexp.MatchString(ibanRegex, iban); !matched {
		return false
	}

	// Country-specific length validation (IBAN lengths differ by country)
	// Get the country code (first two letters)
	countryCode := iban[:2]

	// Check if the length of the IBAN matches the expected length for the country
	if length, exists := countryLengths[countryCode]; exists {
		if len(iban) != length {
			return false
		}
	} else {
		log.LogWarning("Payment Verify", "Unknown country code <"+countryCode+"> in IBAN")
	}

	// IBAN checksum validation using the modulus 97 method
	// 1. Move the first four characters (country code and check digits) to the end
	rearrangedIBAN := iban[4:] + iban[:4]
	// 2. Convert letters to numbers (A=10, B=11, ..., Z=35)
	rearrangedIBAN = convertLettersToNumbers(rearrangedIBAN)
	// 3. Check if the result modulo 97 is 1 (valid checksum)
	if mod97(rearrangedIBAN) != 1 {
		return false
	}
	return true
}

// convertLettersToNumbers converts letters to numbers (A=10, B=11, ..., Z=35)
func convertLettersToNumbers(iban string) string {
	var result strings.Builder
	for _, ch := range iban {
		if ch >= 'A' && ch <= 'Z' {
			// Convert letters A-Z to numbers 10-35
			result.WriteString(strconv.Itoa(int(ch-'A') + 10))
		} else {
			// Keep numbers as they are
			result.WriteByte(byte(ch))
		}
	}
	return result.String()
}

// mod97 calculates the modulo 97 of a string of digits (IBAN validation)
func mod97(iban string) int {
	// We need to calculate modulo 97, so we process the string in chunks of 9 digits at a time
	n := len(iban)
	remainder := 0

	for i := 0; i < n; i++ {
		// Bring in the next digit
		remainder = remainder*10 + int(iban[i]-'0')

		// Only apply modulo 97 every time we have 9 digits
		if remainder >= 1000000000 {
			remainder %= 97
		}
	}

	return remainder % 97
}
