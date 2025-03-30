package main

import (
	"embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/Mr-Comand/qrCodeGeneratorAPI/log"
	"github.com/Mr-Comand/qrCodeGeneratorAPI/payment"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

//go:embed html/*
var content embed.FS

func main() {
	// Read environment variables
	ReadEnvironmentVariables()

	// Print startup message
	log.Log(log.LOGLEVEL_NONE, "System", "Starting server on http://localhost:8080")

	// Set up the routes
	http.Handle("/", AddPrefix("html", http.FileServer(http.FS(content))))
	http.HandleFunc("/api/qrcode", generateQRCode)
	http.HandleFunc("/api/payment", generatePaymentQRCode)

	// Start the server
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.LogError("System", err)
	}
}

// ReadEnvironmentVariables reads the required environment variables for currencies
func ReadEnvironmentVariables() {
	// Read the log level from the environment variable (default to DEBUG if not set)
	logLevelEnv := os.Getenv("LOG_LEVEL")
	switch strings.ToUpper(logLevelEnv) {
	case "ERROR":
		log.SetLogLevel(log.LOGLEVEL_ERROR)
	case "WARNING":
		log.SetLogLevel(log.LOGLEVEL_WARNING)
	case "INFO":
		log.SetLogLevel(log.LOGLEVEL_INFO)
	case "DEBUG":
		log.SetLogLevel(log.LOGLEVEL_DEBUG)
	default:
		// Default to DEBUG if the environment variable is not set or invalid
		log.SetLogLevel(log.LOGLEVEL_DEBUG)
	}

	// Read the anonymize flag from the environment variable (default to true if not set)
	anonymizeEnv := os.Getenv("ANONYMIZE")
	if anonymizeEnv == "false" {
		log.SetAnonymize(false)
	} else {
		log.SetAnonymize(true)
	}

	allowedCurrenciesEnv := os.Getenv("ALLOWED_CURRENCIES")
	if allowedCurrenciesEnv != "" {
		// Split the string into a slice and set allowed currencies
		payment.SetAllowedCurrencies(strings.Split(allowedCurrenciesEnv, ","))
	}

	// Read the default currency from the environment variable or use EUR
	defaultCurrencyEnv := os.Getenv("DEFAULT_CURRENCY")
	if defaultCurrencyEnv != "" {
		payment.SetDefaultCurrency(defaultCurrencyEnv)
	}

}
func generateQRCode(w http.ResponseWriter, r *http.Request) {
	data := r.URL.Query().Get("data")
	if data == "" {
		http.Error(w, "Missing 'data' query parameter", http.StatusBadRequest)
		log.LogInfo("QR Code", "Request whith missing 'data' query parameter")
		return
	}

	w.Header().Set("Content-Type", "image/png")

	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		log.LogError("QR Code", err)
		return
	}

	pngData, err := qr.PNG(256)
	if err != nil {
		http.Error(w, "Failed to encode QR code", http.StatusInternalServerError)
		log.LogError("QR Code", err)
		return
	}

	w.Write(pngData)
}
func generatePaymentQRCode(w http.ResponseWriter, r *http.Request) {
	qrData, errorStr := payment.GeneratePaymentQRCode(r)
	if errorStr != "" {
		errorMessageImage(w, errorStr)
		return
	}
	if qrData == "" {
		errorMessageImage(w, "Failed to generate QR code data")
		log.LogErrorString("Payment QR Code", "Failed to generate QR code data")
		return
	} else {
		// Generate QR code
		w.Header().Set("Content-Type", "image/png")
		qr, err := qrcode.New(qrData, qrcode.Medium)
		if err != nil {
			errorMessageImage(w, "Failed to generate QR code")
			log.LogError("Payment QR Code", err)
			return
		}

		pngData, err := qr.PNG(256)
		if err != nil {
			errorMessageImage(w, "Failed to generate QR code")
			log.LogError("Payment QR Code", err)
			return
		}

		w.Write(pngData)
		log.LogInfo("GeneratePaymentQRCode", "QR Code generated successfully")
	}
}

// errorMessage generates an image with the provided error message as text.
func errorMessageImage(w http.ResponseWriter, message string) {
	// Set dimensions to match the QR code aspect ratio (e.g., 200x200)
	width, height := 200, 200
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)

	// Create a drawer to add the text
	d := font.Drawer{
		Dst:  img,
		Src:  image.Black,
		Face: basicfont.Face7x13,
	}

	// Split the message into lines (if it is too long, break it at spaces or manually insert '\n')
	maxLineWidth := width - 20 // Keep some padding on the sides
	words := strings.Fields(message)
	lines := []string{}
	currentLine := ""

	// Split the message into lines based on width
	for _, word := range words {
		// Measure the width of the current word
		lineWidth := d.MeasureString(currentLine + " " + word)
		if lineWidth.Ceil() < maxLineWidth {
			// Add the word to the current line if it fits
			if currentLine == "" {
				currentLine = word
			} else {
				currentLine = currentLine + " " + word
			}
		} else {
			// Otherwise, push the current line and start a new line
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	// Append the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	// Calculate the starting vertical position
	lineHeight := basicfont.Face7x13.Metrics().Height.Ceil()
	totalHeight := len(lines) * lineHeight
	startY := (height - totalHeight) / 2

	// Draw each line of text
	for _, line := range lines {
		// Calculate the horizontal position to center the line
		textWidth := d.MeasureString(line)
		x := (width - int(textWidth>>6)) / 2
		y := startY + lineHeight

		// Set the starting position for the text
		d.Dot = fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)}
		d.DrawString(line)
		startY += lineHeight // Move down for the next line
	}

	// Set response header and encode the image as PNG
	w.Header().Set("Content-Type", "image/png")
	if err := png.Encode(w, img); err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
	}
}

// AddPrefix adds a prefix to the URL path before passing it to the handler
func AddPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}

	// Ensure the prefix ends with a slash to avoid URL concatenation issues
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add the prefix to the URL path
		r.URL.Path = prefix + strings.TrimPrefix(r.URL.Path, "/")
		r.URL.RawPath = prefix + strings.TrimPrefix(r.URL.RawPath, "/")

		// Serve the modified request
		h.ServeHTTP(w, r)
	})
}
