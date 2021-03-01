package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"log"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	god "github.com/go-acme/lego/v4/providers/dns/godaddy"
	"github.com/go-acme/lego/v4/registration"
)

const (
	tlsKey  = "TLS_KEY_FN"
	tlsCert = "TLS_CERT_FN"
)

type godomain struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *godomain) GetEmail() string {
	return u.Email
}
func (u godomain) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *godomain) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// blog.wutuofu.com
func newCert(domain string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	myUser := godomain{
		Email: "hihahajun@gmail.com",
		key:   privateKey,
	}

	config := lego.NewConfig(&myUser)
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		return err
	}

	g := god.NewDefaultConfig()
	g.APIKey = os.Getenv(daddyEnvKey)
	g.APISecret = os.Getenv(daddyEnvSecret)

	p, err := god.NewDNSProviderConfig(g)
	if err != nil {
		return err
	}
	err = client.Challenge.SetDNS01Provider(p)
	if err != nil {
		return err
	}
	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	myUser.Registration = reg
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(os.Getenv(tlsKey), certificates.PrivateKey, 0o600)
	if err != nil {
		return err
	}
	err = os.WriteFile(os.Getenv(tlsCert), certificates.Certificate, 0o600)
	if err != nil {
		return err
	}
	all, err := os.Create("cert/acme.json")
	if err != nil {
		return err
	}
	defer all.Close()
	return json.NewEncoder(all).Encode(certificates)
}
