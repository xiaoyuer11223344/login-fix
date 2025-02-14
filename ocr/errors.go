package ocr

import "errors"

var (
	ErrInvalidImage     = errors.New("invalid image data")
	ErrServiceUnavailable = errors.New("OCR service unavailable")
	ErrTimeout          = errors.New("OCR request timed out")
	ErrInvalidResponse  = errors.New("invalid response from OCR service")
)

// Error constants for OCR operations
const (
	MaxImageSize    = 1024 * 1024 // 1MB
	MinImageSize    = 100         // 100 bytes
)
