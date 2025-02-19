package browser

import (
	"context"
	"fmt"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

type Browser struct {
	browser *rod.Browser
	page    *rod.Page
	mu      sync.Mutex

	// extra
	captchaHandler *CaptchaHandler      // Handler for processing captcha challenges
	authTokens     map[string]string    // Store detected auth tokens
	lastStatus     int                  // Store last HTTP status code
	lastResponse   string               // Store last response body for error detection
	selectorCache  map[string]*Selector // Cache successful selectors by URL for better performance
}

var MyDevice = devices.Device{
	Title:          "Chrome computer",
	Capabilities:   []string{"touch", "mobile"},
	UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	AcceptLanguage: "en",
	Screen: devices.Screen{
		DevicePixelRatio: 2,
		Horizontal: devices.ScreenSize{
			Width:  1500,
			Height: 900,
		},
		Vertical: devices.ScreenSize{
			Width:  1500,
			Height: 900,
		},
	},
}

// New
// @Description: 初始化go-rod browser
// @param headless
// @param proxy
// @return *Browser
// @return error
func New(ctx context.Context, headless bool, proxy string, ocrBaseURL string) (*Browser, error) {
	l := launcher.New().Context(ctx).Headless(headless).NoSandbox(true)
	l.Set("ignore-certificate-errors").
		Delete("disable-component-extensions-with-background-pages").
		Set("disable-extensions").
		Append("disable-features", "BlinkGenPropertyTrees").
		Set("hide-scrollbars").
		Set("mute-audio").
		Set("no-default-browser-check").
		Delete("no-startup-window").
		Set("password-store", "basic").
		Set("safebrowsing-disable-auto-update").
		Set("disable-gpu").
		Set("no-default-browser-check").
		Set("disable-images", "true").
		Set("enable-automation", "false").                    // 防止监测 webdriver
		Set("disable-blink-features", "AutomationControlled") // 禁用 blink 特征，绕过了加速乐检测

	if proxy != "" {
		l = l.Proxy(proxy)
	}

	browser := rod.New().ControlURL(l.MustLaunch()).MustConnect()
	//browser.DefaultDevice(MyDevice)

	b := &Browser{
		browser:       browser,
		authTokens:    make(map[string]string),
		selectorCache: make(map[string]*Selector),
	}

	// 网络流量监听器
	// router := browser.HijackRequests()
	// router.MustAdd("*", func(ctx *rod.Hijack) {
	//	b.handleNetworkResponse(ctx)
	//	ctx.ContinueRequest(&proto.FetchContinueRequest{})
	// })
	//go router.Run()

	// 初始化OCR识别处理器
	if ocrBaseURL != "" {
		captchaHandler, err := NewCaptchaHandler(b, ocrBaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize captcha handler: %w", err)
		}
		b.captchaHandler = captchaHandler
	}

	return b, nil
}

// Close
// @Description: 资源释放
// @receiver b
// @return error
func (b *Browser) Close() error {
	if b.page != nil {
		if err := b.page.Close(); err != nil {
			return fmt.Errorf("failed to close page: %w", err)
		}
	}
	if err := b.browser.Close(); err != nil {
		return fmt.Errorf("failed to close browser: %w", err)
	}
	return nil
}

func (b *Browser) findElement(selector, name string) (*rod.Element, error) {

	var el *rod.Element
	var err error

	// Wait for element with retry and fallback
	for i := 0; i < 3; i++ {
		// Try XPath first
		el, err = b.page.ElementX(selector)
		if err == nil && el != nil {
			if visible, _ := el.Visible(); visible {
				log.Printf("%s Element found and ready", name)
				return el, nil
			}
		}

		if i < 2 {
			//time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			time.Sleep(time.Duration(1 << uint(i)))
			log.Printf("%s Element not found, %d retrying...", name, i+1)
		}
	}

	return nil, fmt.Errorf("%s element not found or not visible after retries", name)
}
func (b *Browser) PerformLogin(ctx context.Context, selector *Selector, username, password string) error {
	var err error

	// todo: Find UserInput elements
	var userEL *rod.Element
	if userEL, err = b.findElementWithContext(ctx, selector.UserInput, "username input", 10*time.Second); err != nil {
		return err
	}
	if err = userEL.Input(username); err != nil {
		return fmt.Errorf("failed to input username: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// todo: Find PasswordInput elements
	var passEl *rod.Element
	if passEl, err = b.findElementWithContext(ctx, selector.PasswordInput, "password input", 10*time.Second); err != nil {
		return err
	}
	if err = passEl.Input(password); err != nil {
		return fmt.Errorf("failed to input password: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// todo: Find captcha elements
	if b.captchaHandler != nil {
		var captchaText string
		if selector.CaptchaImg != "" && selector.CaptchaInput != "" {
			log.Printf("Handling captcha challenge")
			if imgEL, _err := b.findElementWithContext(ctx, selector.CaptchaImg, "captcha image", 10*time.Second); _err == nil {
				if captchaText, _err = b.captchaHandler.HandleCaptcha(imgEL); _err == nil {
					var captchaEl *rod.Element
					if captchaEl, _err = b.findElementWithContext(ctx, selector.CaptchaInput, "captcha input", 10*time.Second); err == nil {
						if _err = captchaEl.Input(captchaText); _err != nil {

						}
					}
				}
			}
			log.Printf("Captcha input completed")
		} else {
			log.Printf("No captcha elements found, proceeding without captcha")
		}
	}
	// todo: Find LoginBtn elements
	var btnEL *rod.Element
	if btnEL, err = b.findElementWithContext(ctx, selector.LoginBtn, "login button", 10*time.Second); err != nil {
		return err
	}
	// todo: exec btn
	if _, err = btnEL.Eval(`(xpath) => {
			const xpathExpression = xpath;
			const result = document.evaluate(
				xpathExpression,
				document,
				null,
				XPathResult.FIRST_ORDERED_NODE_TYPE,
				null
			);
	
			const element = result.singleNodeValue; // 获取匹配的节点
			if (element) {
				element.click()
				return true;
			}
			return false;
		}`, selector.LoginBtn); err != nil {
		return fmt.Errorf("failed to click login button: %v", err)
	}

	time.Sleep(3 * time.Second)

	return nil
}

func (b *Browser) findElementWithContext(parentCtx context.Context, selector, description string, timeout time.Duration) (*rod.Element, error) {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	ch := make(chan *rod.Element, 1)
	errCh := make(chan error, 1)

	go func() {
		el, err := b.findElement(selector, description)
		if err != nil {
			errCh <- err
			return
		}
		ch <- el
	}()

	select {
	case el := <-ch:
		return el, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout while finding %s: %v", description, ctx.Err())
	}
}

func (b *Browser) Login(ctx context.Context, selector *Selector, username, password string) error {
	start := time.Now()
	if err := b.PerformLogin(ctx, selector, username, password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	log.Printf("Login form submitted %s", time.Since(start))
	return nil
}

func (b *Browser) Navigate(ctx context.Context, url string) error {
	var err error

	log.Printf("Starting navigation %s", url)

	// Clean up previous session
	if b.page != nil {
		if err = b.page.Close(); err != nil {
			log.Printf("%s, Error during cleanup", url)
		}
	}

	// Create new browser page
	var page *rod.Page
	loginURL := strings.TrimRight(url, "/")
	page, err = b.browser.Page(proto.TargetCreateTarget{URL: loginURL})
	if err != nil {
		return fmt.Errorf("page creation failed: %w", err)
	}

	b.page = page.Context(ctx)

	if err = b.page.WaitLoad(); err != nil {
		return fmt.Errorf("page load failed: %w", err)
	}

	//// Create error channel for timeout handling
	//errChan := make(chan error, 1)
	//go func() {
	//	defer close(errChan)
	//
	//	// Configure viewport
	//	//if err = b.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
	//	//	Width:  1920,
	//	//	Height: 1080,
	//	//}); err != nil {
	//	//	errChan <- fmt.Errorf("viewport setup failed: %w", err)
	//	//	return
	//	//}
	//
	//	// Wait for initial page load
	//	if err = b.page.WaitLoad(); err != nil {
	//		errChan <- fmt.Errorf("page load failed: %w", err)
	//		return
	//	}
	//
	//	errChan <- nil
	//}()
	//
	//// Wait for navigation to complete or timeout
	//select {
	//case err = <-errChan:
	//	if err != nil {
	//		return fmt.Errorf("navigation failed: %w", err)
	//	}
	//case <-ctx.Done():
	//	return fmt.Errorf("login timed out after %v", ctx.Err())
	//}

	log.Println("Navigation completed successfully")
	return nil
}

// GetHtmlContent
// @Description: 获取当前rod.Page对象的页面信息
// @receiver b
// @return string
// @return error
func (b *Browser) GetHtmlContent() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.page != nil {
		return b.page.HTML()
	} else {
		return "", fmt.Errorf("browser page is nil")
	}
}

// GetPage
// @Description: 获取当前rod.Page对象
// @receiver b
// @return *rod.Page
func (b *Browser) GetPage() *rod.Page {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.page
}
