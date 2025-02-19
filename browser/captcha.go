package browser

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"strings"

	"log"
	"login-fix/ocr"
)

type CaptchaHandler struct {
	browser *Browser
	client  *ocr.Client
	config  *ocr.Config
}

func NewCaptchaHandler(browser *Browser, ocrBaseURL string) (*CaptchaHandler, error) {
	if ocrBaseURL == "" {
		return nil, errors.New("OCR base URL must be configured")
	}

	client, err := ocr.NewClient(ocrBaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCR client: %w", err)
	}

	config := ocr.NewConfigWithURL(ocrBaseURL)
	if err = config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid OCR config: %w", err)
	}

	return &CaptchaHandler{
		browser: browser,
		client:  client,
		config:  config,
	}, nil
}

func (h *CaptchaHandler) HandleCaptcha(imgEL *rod.Element) (string, error) {

	data, err := imgEL.Screenshot(proto.PageCaptureScreenshotFormat("png"), 1)
	if err != nil {
		return "", fmt.Errorf("captcha image data failed: %w", err)
	}

	base64Data := base64.StdEncoding.EncodeToString(data)
	if len(base64Data) < ocr.MinImageSize || len(base64Data) > ocr.MaxImageSize {
		return "", ocr.ErrInvalidImage
	}

	result, err := h.client.RecognizeCaptcha(base64Data)
	if err != nil {
		return "", fmt.Errorf("OCR failed: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" {
		return "", ocr.ErrInvalidResponse
	}

	log.Printf("Captcha recognized")
	return result, nil
}
