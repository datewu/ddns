package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

// https://developer.godaddy.com/doc/endpoint/domains

const (
	daddyEnvKey    = "GODADDY_API_KEY"
	daddyEnvSecret = "GODADDY_API_SECRET"
)

type goDaddy struct {
	uri, key, sercet string
}

func (g *goDaddy) doWork() error {
	ipOnGateWay, err := g.getMyIP()
	if err != nil {
		return err
	}
	re, err := g.getRecord()
	if err != nil {
		return err
	}
	if re == ipOnGateWay {
		log.Info().
			Str("record", re).
			Msg("ip record matched :)")
		return nil
	}
	err = g.setRecord(ipOnGateWay)
	if err != nil {
		log.Error().
			Err(err).
			Str("gateway", ipOnGateWay).
			Msg("fail to update dns record")
		return err
	}
	log.Info().
		Str("gateway", ipOnGateWay).
		Msg("record update to gateway addr")
	return nil
}

func (g *goDaddy) sync() {
	g.init()
	for {
		go func() {
			// retry for dns issue
			for retry := 0; retry < 5; retry++ {
				err := g.doWork()
				if err != nil {
					log.Error().
						Err(err).
						Int("try", retry).
						Msg("do work failed")
					delay := math.Pow(2, float64(retry+1))
					time.Sleep(time.Duration(delay) * time.Second)
					continue
				}
				return
			}
		}()
		time.Sleep(6 * time.Hour)
	}
}

// container should be point value
func (g *goDaddy) doAPI(method, path string, payload io.Reader, container interface{}) error {
	cli := g.getClient()
	req, err := http.NewRequest(method, g.uri+path, payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("fail to creat request")
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	g.setAuthHeader(req)
	res, err := cli.Do(req)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("fail to do request")
		return err
	}
	defer res.Body.Close()
	if container != nil {
		err = json.NewDecoder(res.Body).Decode(container)
		if err != nil {
			log.Error().
				Err(err).
				Str("path", path).
				Msg("fail to unmarsh res.body")
			return err
		}
		return nil
	}
	if res.StatusCode > 399 || res.StatusCode < 200 {
		return errors.New("http response code not ok: " + res.Status)
	}
	return nil
}

func (g *goDaddy) init() {
	g.uri = "https://api.godaddy.com"
	g.key = os.Getenv(daddyEnvKey)
	g.sercet = os.Getenv(daddyEnvSecret)
}

func (g *goDaddy) setAuthHeader(r *http.Request) {
	token := fmt.Sprintf("sso-key %s:%s", g.key, g.sercet)
	r.Header.Set("Authorization", token)
}

func (g *goDaddy) getClient() *http.Client {
	g.uri = "https://api.godaddy.com"
	ts := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := &http.Client{
		Transport: ts,
		Timeout:   5 * time.Minute,
	}
	return c
}

func (g *goDaddy) getMyIP() (string, error) {
	url := "https://www.ip.cn/api/index?ip=&type=0"
	cli := g.getClient()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail to creat request")
		return "", err
	}
	res, err := cli.Do(req)
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail to do request")
		return "", err
	}
	defer res.Body.Close()
	data := struct {
		Rs       int    `json:"rs"`
		Code     int    `json:"code"`
		Address  string `json:"address"`
		IP       string `json:"ip"`
		Isdomain int    `json:"isdomain"`
	}{}
	err = json.NewDecoder(res.Body).Decode(&data)
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail to unmarsh res.body")
		return "", err
	}
	log.Info().
		Str("myip", data.IP).
		Msg("success got ip")
	return data.IP, nil
}

func (g *goDaddy) getRecord() (string, error) {
	domain := "wutuofu.com"
	name := "blog"
	path := fmt.Sprintf("/v1/domains/%s/records/A/%s", domain, name)
	data := []struct {
		Data string `json:"data"`
		Name string `json:"name"`
		TTL  int    `json:"ttl"`
		Type string `json:"type"`
	}{}
	err := g.doAPI(http.MethodGet, path, nil, &data)
	if err != nil {
		return "", err
	}
	if len(data) < 1 {
		return "", errors.New("empty array response")
	}
	return data[0].Data, nil
}

func (g *goDaddy) setRecord(ip string) error {
	domain := "wutuofu.com"
	name := "blog"
	path := fmt.Sprintf("/v1/domains/%s/records/A/%s", domain, name)
	data := make([]struct {
		Data string `json:"data"`
		TTL  int    `json:"ttl"`
	}, 1)
	data[0].Data = ip
	data[0].TTL = 600
	bs, err := json.Marshal(data)
	if err != nil {
		return err
	}
	r := bytes.NewReader(bs)
	err = g.doAPI(http.MethodPut, path, r, nil)
	if err != nil {
		return err
	}
	if len(data) < 1 {
		return errors.New("empty array response")
	}
	return nil
}
