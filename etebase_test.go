// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package etebase_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/etesync/etebase-go"
)

type EtebaseSuite struct {
	suite.Suite

	account  *etebase.Account
	user     etebase.User
	password string
}

const (
	hostEnv = "ETEBASE_TEST_HOST"
)

func (s *EtebaseSuite) newClient() *etebase.Client {
	host := os.Getenv(hostEnv)
	if host == "" {
		s.T().Skip("Define " + hostEnv + " to run this test")
	}

	return etebase.NewClient(etebase.ClientOptions{
		Host:   host,
		Logger: s.T(), // testing.T implements Logf
	})
}

// SetupSuite signs-up a new user and makes sure that the account instance is
// pointing to a valid etebase server.
// It run once, before the tests in the suite are run.
func (s *EtebaseSuite) SetupSuite() {
	acc, err := etebase.Signup(s.newClient(), s.user, s.password)
	s.Require().NoError(err)

	ok, err := acc.IsEtebaseServer()
	s.Require().NoError(err)
	s.Require().True(ok)
}

// SetupTest logs-in and keep a reference of the account.
// It run before each test.
func (s *EtebaseSuite) SetupTest() {
	acc, err := etebase.Login(s.newClient(), s.user.Username, s.password)
	s.Require().NoError(err)
	s.account = acc
}

// TestPasswordChange changes the password and tries to login a user using the new password.
// It changes the password to the previous one so that SetupTest can still login.
func (s *EtebaseSuite) TestPasswordChange() {
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

func (s *EtebaseSuite) TestLogin() {
	s.Run("UserNotFound", func() {
		_, err := etebase.Login(s.newClient(), "some-not-existing-user", s.password)
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "not found")
	})

	s.Run("WrongPassword", func() {
		_, err := etebase.Login(s.newClient(), s.user.Username, "wrong-password")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "Wrong password")
	})
}

func (s *EtebaseSuite) TestCollections() {
	col, err := etebase.NewEncryptedCollection("testType", nil, nil)
	col.AccessLevel = etebase.Admin
	s.Require().NoError(err)

	s.Require().NoError(
		s.account.CreateCollection(col),
	)

	s.Run("ListMulti", func() {
		found, err := s.account.ListCollections([][]byte{
			col.Type,
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(found)
	})

	s.Run("GetCollection", func() {
		found, err := s.account.GetCollection(col.Item.UID)
		s.Require().NoError(err)
		s.Require().NotEmpty(found.Stoken)

		expected := col
		expected.Stoken = found.Stoken

		s.Require().Equal(expected, found)
	})

	s.Run("NotFound", func() {
		_, err := s.account.GetCollection("not-an-existing-uid")
		s.Require().Error(err)
		s.Require().Contains(err.Error(), "not found")
	})
}

// TestLogout logs-out an account twice. The second time it shouldn't be
// authorized.
func (s *EtebaseSuite) TestLogout() {
	s.Require().NoError(
		s.account.Logout(),
	)
	err := s.account.Logout()

	s.Require().Error(err)
	s.Require().Contains(err.Error(), "Invalid token.")
}

func TestEtebaseSuite(t *testing.T) {
	var (
		id       = fmt.Sprintf("%d", time.Now().Unix())
		username = "test-user-" + id

		user = etebase.User{
			Username: username,
			Email:    username + "@test.com",
		}
		password = "secret"
	)

	suite.Run(t, &EtebaseSuite{
		user:     user,
		password: password,
	})
}
