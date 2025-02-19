# login-fix

封装自动化登录SDK

```text
	s := &browser.Selector{
		UserInput:     "//input[@id='loginid']",
		PasswordInput: "//input[@id='userpassword']",
		LoginBtn:      "//*[@id='submit']",
		CaptchaInput:  "//input[@id='validatecode']",
		CaptchaImg:    "//div[@class='e9login-form-vc-img']/img[1]",
	}
```