<p align="center">
  <img width="120" src="https://github.com/etesync/etesync-web/blob/master/src/images/logo.svg" />
  <h1 align="center">Etebase - Encrypt Everything</h1>
</p>

A (work in progress) Go library for Etebase

![GitHub tag](https://img.shields.io/github/tag/etesync/etebase-go.svg)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/etesync/etebase-go)](https://pkg.go.dev/github.com/etesync/etebase-go)
[![Build Status](https://github.com/etesync/etebase-go/workflows/Build/badge.svg)](https://github.com/etesync/etebase-go/actions/)
[![codecov](https://codecov.io/gh/etesync/etebase-go/branch/master/graph/badge.svg?token=G7A71HXMIR)](https://codecov.io/gh/etesync/etebase-go)
[![Chat with us](https://img.shields.io/badge/chat-IRC%20|%20Matrix%20|%20Web-blue.svg)](https://www.etebase.com/community-chat/)

**Star** and **watch** to show your interest and get notified once it's ready!

## TODO:
- [x] Authentication
  - [x] Signup
  - [x] Login
  - [x] Logout
  - [x] Password Change
- [ ] Collections
- [ ] Items
- [ ] Invitations

# Testing

To test, run the `etesync/test-server` image using the latest version, e.g.,

```
docker run -p 3735:3735 -d etesync/test-server:latest
```

and then set `ETEBASE_TEST_HOST` to the host:port on which that is running; for the docker invocation above, that's
```
export ETEBASE_TEST_HOST=localhost:3735
```

and then run the tests

```
go test -v -v ./...
```

# Documentation

In addition to the API documentation, there are docs available at https://docs.etebase.com
