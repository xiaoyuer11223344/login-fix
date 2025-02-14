package browser

import (
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

type FormDesc struct {
	Form      *rod.Element
	Score     int
	HasLogin  bool
	HasPass   bool
	HasSubmit bool
	Position  proto.Point
	selector  *Selector
}

type Selector struct {
	UserInput     string `yaml:"userInput" json:"userInput"`
	PasswordInput string `yaml:"passwordInput" json:"passwordInput"`
	LoginBtn      string `yaml:"loginBtn" json:"loginBtn"`
	RememberMe    string `yaml:"rememberMe" json:"rememberMe"`
	CaptchaInput  string `yaml:"captchaInput" json:"captchaInput"`
	CaptchaImg    string `yaml:"captchaImg" json:"captchaImg"`
}

// findFormElements
// @Description: 匹配表单内元素
// @receiver b
// @param form
// @return *Selector
// @return error
func (b *Browser) findElements(
	userInputSelectors,
	passInputSelectors,
	loginBtnSelectors,
	captchaInputSelectors,
	captchaImageSelectors []string,
) (*Selector, error) {

	logger := log.WithField("action", "find_form_elements")

	selector := &Selector{}

	// Find username input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range userInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.UserInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.UserInput).Debug("Found username input")
					goto foundPass
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Username input not found, retrying...")
		}
	}

foundPass:
	// Find password input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range passInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.PasswordInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.PasswordInput).Debug("Found password input")
					goto foundButton
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Password input not found, retrying...")
		}
	}

foundButton:
	// Find login button with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range loginBtnSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					logger.WithField("xpath", selector.LoginBtn).Debug("Found login button")
					goto foundCaptchaInput
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Login button not found, retrying...")
		}
	}

foundCaptchaInput:
	if b.captchaHandler != nil {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaInputSelectors {
				if el, err := b.page.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaInput = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaInput).Debug("Found Captcha Input")
						goto foundCaptchaImage
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha input not found, retrying...")
			}
		}
	}

foundCaptchaImage:
	if b.captchaHandler != nil && selector.CaptchaInput != "" {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaImageSelectors {
				if el, err := b.page.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaImg = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaImg).Debug("Found Captcha Image")
						goto over
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha image not found, retrying...")
			}
		}
	}

over:
	//return selector, nil
	if selector.UserInput != "" && selector.PasswordInput != "" && selector.LoginBtn != "" {
		return selector, nil
	}

	return nil, fmt.Errorf("not form all elements found")
}
