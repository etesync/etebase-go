package etebase

func ExampleLogin() {
	var (
		client = NewClient(PartnerClientOptions("your-name"))
		user   = User{
			Username: "john",
			Email:    "john@etebase.com",
		}
		password = "john's-secret"
	)

	if _, err := Signup(client, user, password); err != nil {
		panic(err)
	}

	acc, err := Login(client, user.Username, password)
	if err != nil {
		panic(err)
	}

	_ = acc
}
