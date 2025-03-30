package log

import (
	"fmt"
	"regexp"
)

// Define log levels as constants
const (
	LOGLEVEL_NONE = iota
	LOGLEVEL_ERROR
	LOGLEVEL_WARNING
	LOGLEVEL_INFO
	LOGLEVEL_DEBUG
)

// Define a global variable to control whether to anonymize sensitive information sensitive information should be put in <> brackets
var anonymise = true
var currentLogLevel = LOGLEVEL_DEBUG

// SetLogLevel allows you to change the current log level at runtime
func SetLogLevel(level int) {
	currentLogLevel = level
}

// SetAnonymize allows you to toggle the anonymization of sensitive information
func SetAnonymize(anonymize bool) {
	anonymise = anonymize
}

// anonymizeMessage replaces sensitive data with {}
func anonymizeMessage(message string) string {
	// Simple regex pattern to match email addresses, credit card numbers, and other sensitive data
	// You can customize the patterns based on what you want to anonymize
	sensitivePatterns := []string{
		`\<[^>]*\>`, // other sensitive data in <>
		`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`, // Emails
		`(?:\d[ -]*?){13,19}`,                            // Credit card numbers
		`(?i)password\s*[:=]\s*[\w!@#$%^&*()]+`,          // Passwords
		`(?i)ssn\s*[:=]\s*\d{3}-\d{2}-\d{4}`,             // SSN format
		`(?i)iban\s*[:=]\s*[A-Z]{2}\d{2}[A-Z0-9]{1,30}`,  // IBAN format
	}
	// Replace sensitive patterns with {}
	for _, pattern := range sensitivePatterns {
		re := regexp.MustCompile(pattern)
		message = re.ReplaceAllString(message, "<Cencored>")
	}

	return message
}

// Log is the generic log function that handles logging based on the level and tag
func Log(level int, tag string, message string) {
	if level <= currentLogLevel {
		if anonymise {
			message = anonymizeMessage(message)
		}
		// Prefix the message with the log level and tag
		var logLevel string
		switch level {
		case LOGLEVEL_ERROR:
			logLevel = "ERROR"
		case LOGLEVEL_WARNING:
			logLevel = "WARNING"
		case LOGLEVEL_INFO:
			logLevel = "INFO"
		case LOGLEVEL_DEBUG:
			logLevel = "DEBUG"
		}

		// Print the log message with the tag and level
		fmt.Printf("[%s] %s: %s\n", logLevel, tag, message)
	}
}

// LogError logs error messages if the current log level allows it
func LogError(tag string, err error) {
	if err != nil {
		Log(LOGLEVEL_ERROR, tag, err.Error())
	}
}

// LogErrorString logs error messages from a string input
func LogErrorString(tag string, err string) {
	Log(LOGLEVEL_ERROR, tag, err)
}

// LogInfo logs informational messages if the current log level allows it
func LogInfo(tag string, message string) {
	Log(LOGLEVEL_INFO, tag, message)
}

// LogDebug logs debug messages if the current log level allows it
func LogDebug(tag string, message string) {
	Log(LOGLEVEL_DEBUG, tag, message)
}

// LogWarning logs warning messages if the current log level allows it
func LogWarning(tag string, message string) {
	Log(LOGLEVEL_WARNING, tag, message)
}
