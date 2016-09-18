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

	// Get required configuration values for the SMTP server

	hostname, err := conf.QuerySingle("hostname 0")
	assertCfg(err, "hostname <my.host.name>")

	bind, err := conf.QuerySingle("bind 0")
	assertCfg(err, "bind <host/ip>[:port]")

	// Create SMTP server and start listening
	server, err := smtp.NewServer(bind, hostname)
	assert(err)

	// Set configuration options to server options
	setServerConfigurationOptions(server)

	// Setup received mail handler
	server.OnReceivedMail = HandleReceivedMail

	// Setup auth handler
	server.OnAuthRequest = HandleLocalAuthRequest

	// Check for custom max size
	maxsize, err := conf.QuerySingle("max_size 0")
	if err == nil {
		maxsizeInt, err := parseByteSize(maxsize)
		if err != nil {
			log.Fatalf("The value of 'max_size' (%s) was not recognized as a valid size\r\n", maxsize)
		} else {
			server.MaxSize = maxsizeInt
		}
	}

	log.Printf("[SMTPd] Listening on %s\r\n", bind)

	// Start serving SMTP connections
	assert(server.ListenAndServe())
}

func setServerConfigurationOptions(server *smtp.Server) {
	domains, err := conf.Query("domain")
	assert(err)
	domainCount := len(domains)

	// Warn if there are no domains configured
	if domainCount < 1 {
		//TODO Check for open relay
		log.Println("[meirud] No domain configured! Ignore this warning if this is the wanted behavior (open relay)")
		return
	}

	log.Printf("[meirud] Loaded %d domain(s)\r\n", domainCount)
	server.LocalDomains = make([]string, domainCount)
	for i, domainProperty := range domains {
		if len(domainProperty.Values) < 1 {
			log.Fatalln("Defined domain block without domain name in configuration!")
		}
		server.LocalDomains[i] = domainProperty.Values[0]
	}
}
