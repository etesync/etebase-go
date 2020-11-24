package etebase

import "log"

func ExampleNewAccount() {
	acc := NewAccount(
		NewClient("your-account-name", DefaultClientOptions),
	)

	user := User{
		Username: "john",
		Email:    "john@etebase.com",
	}

	if err := acc.Signup(user, "my-password"); err != nil {
		log.Fatalf("signup error: %v", err)
	}
}
