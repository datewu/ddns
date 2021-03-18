package main

import (
	"crypto"

	"github.com/go-acme/lego/v4/registration"
)

type user struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func newUser(email string) *user {
	u := &user{
		Email: email,
	}
	return u
}

func (u *user) GetEmail() string {
	return u.Email
}
func (u user) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *user) GetPrivateKey() crypto.PrivateKey {
	return u.key
}
