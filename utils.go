package main

import (
	"math"
	"time"

	"github.com/rs/zerolog/log"
)

type retryFunc func() error

func (r retryFunc) retry(ceil int) {
	for i := 0; i < ceil; i++ {
		err := r()
		if err != nil {
			log.Error().
				Err(err).
				Int("round", i).
				Msg("failed, retrying")
			delay := math.Pow(2, float64(i+1))
			time.Sleep(time.Duration(delay) * time.Second)
			continue
		}
		return
	}
}
