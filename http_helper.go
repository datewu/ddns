package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type reqAgent struct {
	cli         *http.Client
	cb          func(*http.Request)
	method, url string
}

// container must be point value
func (r *reqAgent) doHTTP(payload io.Reader, container interface{}) error {
	req, err := http.NewRequest(r.method, r.url, payload)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", r.url).
			Msg("fail to creat request")
		return err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	}
	if r.cb != nil {
		r.cb(req)
	}
	res, err := r.cli.Do(req)
	if err != nil {
		log.Error().
			Err(err).
			Str("url", r.url).
			Msg("fail to do request")
		return err
	}
	defer res.Body.Close()
	if container != nil {
		err = json.NewDecoder(res.Body).Decode(container)
		if err != nil {
			log.Error().
				Err(err).
				Str("url", r.url).
				Msg("fail to unmarsh res.body")
			return err
		}
	}
	if res.StatusCode > 399 || res.StatusCode < 200 {
		return errors.New("http response code not ok: " + res.Status)
	}
	return nil
}
