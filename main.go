package main

import "github.com/rs/zerolog/log"

func main() {
	parseFlag()
	g := newGodday()
	go g.sync()
	p, err := g.newProvider()
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail to goaddy dns provider")
		return
	}
	u := &user{
		Email: "hihahajun@gmail.com",
	}
	b, err := newBot(u, p)
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail init lego bot")
		return
	}
	b.run("blog.wutuofu.com")
}
