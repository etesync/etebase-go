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

// SetupSuite signs-up a new user and makes sure that the account instance is
// pointing to a valid etebase server.
// It run once, before the tests in the suite are run.
func (s *AccountSuite) SetupSuite() {
	acc, err := etebase.Signup(s.newClient(), s.user, s.password)
	s.Require().NoError(err)

	ok, err := acc.IsEtebaseServer()
	s.Require().NoError(err)
	s.Require().True(ok)
}

// SetupTest logs-in and keep a reference of the account.
// It run before each test.
func (s *AccountSuite) SetupTest() {
	acc, err := etebase.Login(s.newClient(), s.user.Username, s.password)
	s.Require().NoError(err)
	s.account = acc
}

// TestPasswordChange changes the password and tries to login a user using the new password.
// It changes the password to the previous one so that SetupTest can still login.
func (s *AccountSuite) TestPasswordChange() {
	newPassword := "a-new-password"

	s.Require().NoError(
		s.account.PasswordChange(newPassword),
	)
	_, err := etebase.Login(s.newClient(), s.user.Username, newPassword)
	s.Require().NoError(err)

	s.Require().NoError(
		s.account.PasswordChange(s.password),
	)
}

// TestLogout logs-out an account twice. The second time it shouldn't be
// authorized.
func (s *AccountSuite) TestLogout() {
	s.Require().NoError(
		s.account.Logout(),
	)
	err := s.account.Logout()

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "Unauthorized")
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
