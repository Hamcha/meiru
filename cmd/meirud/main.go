package main

import (
	"flag"
	"log"
	"os"

	"github.com/hamcha/meiru/lib/config"
)

var conf config.Config

func assert(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s\r\n", err.Error())
	}
}

func main() {
	cfgpath := flag.String("config", "conf/meiru.conf", "Path to configuration file")
	dump := flag.Bool("dump-cfg", false, "Dump parsed configuration and exit")
	flag.Parse()

	var err error
	conf, err = config.LoadConfig(*cfgpath)
	assert(err)

	if *dump {
		conf.Dump(os.Stderr)
		return
	}
}
