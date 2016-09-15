package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"
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

	dbpath, err := conf.QuerySingle("dbfile 0")
	if err == config.QueryErrSingleTooFewResults {
		log.Fatalln("The configuration value 'dbpath <path/to/db>' is missing! Please add it to the configuration file!")
	}
	if err == config.QueryErrSingleTooFewValues {
		log.Fatalln("The configuration value 'dbpath <path/to/db>' is declared without a value! Please add a path to the database file (will be created if missing).")
	}

	db, err := bolt.Open(dbpath, 0600, &bolt.Options{Timeout: 10 * time.Second})
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *dump {
		conf.Dump(os.Stderr)
		return
	}
}
