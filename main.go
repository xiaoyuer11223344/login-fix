package main

import (
	"context"
	"fmt"
	"github.com/go-rod/rod"
	"log"
	"login-fix/browser"
	"time"
)

func demo1() {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(30)*time.Second)
	defer cancel()

	b, err := browser.New(ctx, false, "", "http://120.26.57.12:8000/")
	if err != nil {
		log.Fatal(err)
	}
	defer b.Close()

	if err = b.Navigate(ctx, "http://zwff.aishi.com:85/"); err != nil {
		log.Fatal(err)
	}

	usernameInput, err := b.GetPage().ElementX("//*[@id='loginid']")
	usernameInput.Input("13012567125")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	// 查找密码输入框并填充文本 "123456"（如果适用）
	passwordInput, err := b.GetPage().ElementX("//*[@id='userpassword']")
	passwordInput.Input("Chint.wmq2024c")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	// 查找密码输入框并填充文本 "123456"（如果适用）
	imageInput, err := b.GetPage().ElementX("//input[@id='validatecode']")
	imageInput.Input("112233")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	// todo: Find LoginBtn elements
	var btnEL *rod.Element
	if btnEL, err = b.GetPage().ElementX("//*[@id='submit']"); err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	fmt.Println(btnEL)

}

func main() {
	demo1()
}
