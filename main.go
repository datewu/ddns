package main

import "github.com/rs/zerolog/log"

func main() {
	parseFlag()
	g := newGodday()
	go g.sync()
	err := runCertBot(g)
	if err != nil {
		log.Error().
			Err(err).
			Msg("fail to run cert bot")
	}
}

func runCertBot(g *goDaddy) error {
	p, err := g.newProvider()
	if err != nil {
		return err
	}
	u := &user{
		Email: "hihahajun@gmail.com",
	}
	b, err := newBot(u, p)
	if err != nil {
		return err
	}
	return b.run("blog.wutuofu.com")
}
