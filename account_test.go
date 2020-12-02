package etebase_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/etesync/etebase-go"
)

type AccountSuite struct {
	suite.Suite

	account  *etebase.Account
	user     etebase.User
	password string
}

func (s *AccountSuite) newClient() *etebase.Client {
	return etebase.NewClient(etebase.ClientOptions{
		Host: os.Getenv("ETEBASE_TEST_HOST"),
	})
}

func (s *AccountSuite) SetupSuite() {
	acc, err := etebase.Signup(s.newClient(), s.user, s.password)
	s.Require().NoError(err)

	acc, err = etebase.Login(s.newClient(), s.user.Username, s.password)
	s.Require().NoError(err)
	s.account = acc

	ok, err := acc.IsEtebaseServer()
	s.Require().NoError(err)
	s.Require().True(ok)
}

func (s *AccountSuite) TestPasswordChange() {

}

func TestAcountSuite(t *testing.T) {
	var (
		host     = os.Getenv("ETEBASE_TEST_HOST")
		id       = fmt.Sprintf("%d", time.Now().Unix())
		username = "test-user-" + id

		user = etebase.User{
			Username: username,
			Email:    username + "@test.com",
		}
		password = "secret"
	)

	if host == "" {
		t.Skip("Define ETEBASE_TEST_HOST to run this test")
	}

	suite.Run(t, &AccountSuite{
		user:     user,
		password: password,
	})
}
