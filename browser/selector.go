package browser

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
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
	selector := &Selector{}

	// Find username input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range userInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.UserInput = el.MustGetXPath(false)
					log.Printf("Found username input")
					goto foundPass
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			log.Printf("Username input not found, %d retrying...", i+1)
		}
	}

foundPass:
	// Find password input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range passInputSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.PasswordInput = el.MustGetXPath(false)
					log.Printf("Found password input")
					goto foundButton
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			log.Printf("Password input not found, %d retrying...", i+1)
		}
	}

foundButton:
	// Find login button with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range loginBtnSelectors {
			if el, err := b.page.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					log.Printf("Found login button")
					goto foundCaptchaInput
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			log.Printf("Login button not found, %d retrying...", i+1)
		}
	}

foundCaptchaInput:
	if b.captchaHandler != nil {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaInputSelectors {
				if el, err := b.page.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaInput = el.MustGetXPath(false)
						log.Printf("Found Captcha Input")
						goto foundCaptchaImage
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				log.Printf("Captcha input not found, %d retrying...", i+1)
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
						log.Printf("Found Captcha Image")
						goto over
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				log.Printf("Captcha image not found, %d retrying...", i+1)
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
