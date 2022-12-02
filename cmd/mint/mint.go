package main

import (
	"github.com/cashubtc/cashu-feni/api"
	_ "github.com/cashubtc/cashu-feni/docs"
	"github.com/cashubtc/cashu-feni/log"
	log "github.com/sirupsen/logrus"
)

// @title Cashu (Feni) golang mint
// @version 0.0.1
// @description Ecash wallet and mint with Bitcoin Lightning support.
// @contact.url https://8333.space:3338
func main() {
	cashuLog.Configure(api.Config.LogLevel)
	log.Info("starting (feni) cashu mint server")
	m := api.New()
	m.StartServer()
}
