package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/boltdb/bolt"

	"github.com/hamcha/meiru/lib/config"
	"github.com/hamcha/meiru/lib/errors"
	"github.com/hamcha/meiru/lib/imap"
	"github.com/hamcha/meiru/lib/mailstore"
	"github.com/hamcha/meiru/lib/smtp"
)

var conf config.Config

func assert(err interface{}) {
	switch e := err.(type) {
	case *errors.Error:
		if e != nil {
			log.Fatalf("[%s] FATAL\n\t%s\r\n", e.Type.Source, e.Error())
		}
	case error:
		if e != nil {
			log.Fatalf("[meirud] FATAL\n\t%s\r\n", e.Error())
		}
	}
}

func assertCfg(err *errors.Error, cfg string) {
	if err != nil {
		if err.Type == config.QueryErrSingleTooFewResults {
			log.Fatalf("The configuration value '%s' is missing! Please add it to the configuration file!\r\n", cfg)
		}
		if err.Type == config.QueryErrSingleTooFewValues {
			log.Fatalf("The configuration value 'dbpath <path/to/db>' is declared without a value! Please add a path to the database file (will be created if missing).\r\n", cfg)
		}
	}
}

func main() {
	cfgpath := flag.String("config", "conf/meiru.conf", "Path to configuration file")
	dump := flag.Bool("dump-cfg", false, "Dump parsed configuration and exit")
	flag.Parse()

	var err error
	conf, err = config.LoadConfig(*cfgpath)
	assert(err)

	dbpath, cfgerr := conf.QuerySingle("dbfile 0")
	assertCfg(cfgerr, "dbfile </path/to/db>")

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

	hostname, cfgerr := conf.QuerySingle("hostname 0")
	assertCfg(cfgerr, "hostname <my.host.name>")

	bind, cfgerr := conf.QuerySingle("bind 0")
	assertCfg(cfgerr, "bind <host/ip>[:port]")

	// Create mailstore for SMTP and IMAP servers

	store := mailstore.NewStore(db)
	store.LoadConfig(&conf)

	queue, queuechan := startSendQueue(db, hostname)
	_, smtpchan := startSMTPServer(bind, hostname, queue)
	_, imapchan := startIMAPServer(bind, store)

	select {
	case err = <-smtpchan:
		assert(err)
	case err = <-imapchan:
		assert(err)
	case err = <-queuechan:
		assert(err)
	}
}

func startSMTPServer(bind, hostname string, queue *SendQueue) (*smtp.Server, <-chan error) {
	// Create SMTP server and start listening
	smtpd, err := smtp.NewServer(bind, hostname)
	assert(err)

	// Set configuration options to server options
	loadSMTPOptions(smtpd)

	// Setup auth handler
	smtpd.OnAuthRequest = HandleLocalAuthRequest

	// Setup sendmail handler
	smtpd.OnReceivedMail = queue.QueueMail

	// Check for custom max size
	maxsize, err := conf.QuerySingle("max_size 0")
	if err == nil {
		maxsizeInt, err := parseByteSize(maxsize)
		if err != nil {
			log.Fatalf("The value of 'max_size' (%s) was not recognized as a valid size\r\n", maxsize)
		} else {
			smtpd.MaxSize = maxsizeInt
		}
	}

	log.Printf("[SMTPd] Listening on %s\r\n", bind)

	// Start serving SMTP connections
	return smtpd, runServer(smtpd.ListenAndServe)
}

func startIMAPServer(bind string, store *mailstore.MailStore) (*imap.Server, <-chan error) {
	// Create IMAP server and start listening
	imapd, err := imap.NewServer(bind, store)
	assert(err)

	// Setup auth handler
	imapd.OnAuthRequest = HandleLocalAuthRequest

	log.Printf("[IMAPd] Listening on %s\r\n", bind)

	// Start serving IMAP connections
	return imapd, runServer(imapd.ListenAndServe)
}

func loadSMTPOptions(smtpd *smtp.Server) {
	domains, err := conf.Query("domain")
	assert(err)
	domainCount := len(domains)

	// Warn if there are no domains configured
	if domainCount < 1 {
		//TODO Check for open relay
		log.Println("[meirud] No domain configured! Ignore this warning if this is the wanted behavior (open relay)")
		return
	}

	smtpd.LocalDomains = make([]string, domainCount)
	for i, domainProperty := range domains {
		if len(domainProperty.Values) < 1 {
			log.Fatalln("Defined domain block without domain name in configuration!")
		}
		smtpd.LocalDomains[i] = domainProperty.Values[0]
	}

	log.Printf("[SMTPd] Loaded %d domain(s)\r\n", domainCount)
}

func startSendQueue(db *bolt.DB, hostname string) (*SendQueue, <-chan error) {
	queue := NewSendQueue(db, hostname)
	return queue, runServer(queue.Serve)
}

func runServer(fn func() error) <-chan error {
	errch := make(chan error)
	go func() {
		errch <- fn()
	}()
	return errch
}
