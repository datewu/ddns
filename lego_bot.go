package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"os"
	"time"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
)

type bot struct {
	u                 *user
	cli               *lego.Client
	provider          challenge.Provider
	privateKey, certs []byte
}

func newBot(u *user, p challenge.Provider) (*bot, error) {
	config := lego.NewConfig(u)
	config.Certificate.KeyType = certcrypto.RSA2048
	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}
	b := &bot{
		u:        u,
		cli:      client,
		provider: p,
	}
	return b, nil
}

func (b *bot) run(domain string) error {
	loaded, err := b.loadCertKey()
	if err != nil {
		return err
	}
	if !loaded {
		r := retryFunc(func() error {
			return b.newCert(domain)
		})
		r.retry(5)
	}
	for {
		r := retryFunc(func() error {
			return b.renew(domain, false)
		})
		r.retry(5)
		time.Sleep(7 * 24 * time.Hour)
	}
}

func (b *bot) newCert(domain string) error {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	b.u.key = privateKey
	err = b.cli.Challenge.SetDNS01Provider(b.provider)
	if err != nil {
		return err
	}
	// New users will need to register
	reg, err := b.cli.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return err
	}
	b.u.Registration = reg
	request := certificate.ObtainRequest{
		Domains: []string{domain},
		Bundle:  true,
	}
	certificates, err := b.cli.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	b.privateKey = certificates.PrivateKey
	b.certs = certificates.Certificate
	return b.saveAll(certificates)
}

func (b *bot) needRenew(days int) (bool, error) {
	cs, err := certcrypto.ParsePEMBundle(b.certs)
	if err != nil {
		return false, err
	}
	deadline := time.Now().Add(time.Duration(days) * 24 * time.Hour)
	for _, v := range cs {
		if v.NotAfter.Before(deadline) {
			return true, nil
		}
	}
	return false, nil
}

func (b *bot) renew(domain string, reuse bool) error {
	need, err := b.needRenew(10)
	if err != nil {
		return err
	}
	if !need {
		return nil
	}
	var privateKey crypto.PrivateKey
	if reuse {
		privateKey, err = certcrypto.ParsePEMPrivateKey(b.privateKey)
		if err != nil {
			return err
		}
	}
	request := certificate.ObtainRequest{
		Domains:    []string{domain},
		Bundle:     true,
		PrivateKey: privateKey,
	}
	res, err := b.cli.Certificate.Obtain(request)
	if err != nil {
		return err
	}
	return b.saveAll(res)
}
func (b *bot) saveAll(res *certificate.Resource) error {
	const (
		tlsKey   = "TLS_KEY_FN"
		tlsCert  = "TLS_CERT_FN"
		acmeJSON = "cert/acme.json"
	)
	err := os.WriteFile(os.Getenv(tlsKey), res.PrivateKey, 0o600)
	if err != nil {
		return err
	}
	err = os.WriteFile(os.Getenv(tlsCert), res.Certificate, 0o600)
	if err != nil {
		return err
	}
	all, err := os.Create(acmeJSON)
	if err != nil {
		return err
	}
	defer all.Close()
	return json.NewEncoder(all).Encode(res)
}

func (b *bot) loadCertKey() (bool, error) {
	const (
		tlsKey  = "TLS_KEY_FN"
		tlsCert = "TLS_CERT_FN"
	)
	kb, err := os.ReadFile(os.Getenv(tlsKey))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	cb, err := os.ReadFile(os.Getenv(tlsCert))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	b.privateKey = kb
	b.certs = cb
	return true, nil
}
