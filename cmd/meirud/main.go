package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/hamcha/meiru/lib/config"
)

var conf config.Config

func assert(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %d\r\n", err.Error())
	}
}

func main() {
	cfgpath := flag.String("config", "conf/meiru.conf", "Path to configuration file")
	flag.Parse()

	var err error
	conf, err = config.LoadConfig(*cfgpath)
	assert(err)

	props := conf.Query("user password")
	fmt.Println(props)
}
