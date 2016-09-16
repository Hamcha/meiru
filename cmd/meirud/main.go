package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"

	"github.com/hamcha/meiru/lib/config"
	"github.com/hamcha/meiru/lib/smtp"
)

var conf config.Config

func assert(err error) {
	if err != nil {
		log.Fatalf("FATAL ERROR: %s\r\n", err.Error())
	}
}

func assertCfg(err error, cfg string) {
	if err == config.QueryErrSingleTooFewResults {
		log.Fatalf("The configuration value '%s' is missing! Please add it to the configuration file!\r\n", cfg)
	}
	if err == config.QueryErrSingleTooFewValues {
		log.Fatalf("The configuration value 'dbpath <path/to/db>' is declared without a value! Please add a path to the database file (will be created if missing).\r\n", cfg)
	}
}

func main() {
	cfgpath := flag.String("config", "conf/meiru.conf", "Path to configuration file")
	dump := flag.Bool("dump-cfg", false, "Dump parsed configuration and exit")
	flag.Parse()

	var err error
	conf, err = config.LoadConfig(*cfgpath)
	assert(err)

	dbpath, err := conf.QuerySingle("dbfile 0")
	assertCfg(err, "dbfile </path/to/db>")

	db, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *dump {
		conf.Dump(os.Stderr)
		return
	}

	hostname, err := conf.QuerySingle("hostname 0")
	assertCfg(err, "hostname <my.host.name>")

	bind, err := conf.QuerySingle("bind 0")
	assertCfg(err, "bind <host/ip>[:port]")

	server, err := smtp.NewServer(bind, hostname)
	assert(err)

	log.Printf("[SMTPd] Listening on %s\r\n", bind)

	assert(server.ListenAndServe())
}
