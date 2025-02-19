package main

import (
	"context"
	"log"
	"login-fix/browser"
	"strings"
	"time"
)

func main() {

	loginURL := "https://dftc-wps.dfmc.com.cn/"

	s := &browser.Selector{
		UserInput:     "//input[@id='loginid']",
		PasswordInput: "//input[@id='userpassword']",
		LoginBtn:      "//*[@id='submit']",
		CaptchaInput:  "//input[@id='validatecode']",
		CaptchaImg:    "//div[@class='e9login-form-vc-img']/img[1]",
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
	defer cancel()

	b, err := browser.New(ctx, false, "", "http://120.26.57.12:8000")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	if err = b.Navigate(ctx, loginURL); err != nil {
		log.Fatal(err)
	}

	if err = b.PerformLogin(ctx, s, "admin", "123456"); err != nil {
		log.Fatal(err)
	}

	html, err := b.GetHtmlContent()
	if err != nil {
		log.Fatal(err)
	}

	if strings.Contains(html, "e9header-quick-search-input") || strings.Contains(html, "portal-bay-window-outer") {
		log.Printf("login success")
	} else {
		log.Printf("login failed")
	}
}
