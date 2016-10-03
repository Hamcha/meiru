package mailstore

import (
	"log"

	"github.com/boltdb/bolt"
	"github.com/hamcha/meiru/lib/config"
	"github.com/hamcha/meiru/lib/errors"
)

var (
	ErrSrcMailstore errors.ErrorSource = "mailstore"
)

type MailStore struct {
	db *bolt.DB

	Domains map[string]Domain
}

type Domain struct {
	Users    map[string]User
	CatchAll string
}

type User struct {
	MailboxDir string
}

func NewStore(db *bolt.DB) *MailStore {
	return &MailStore{
		db:      db,
		Domains: make(map[string]Domain),
	}
}

func (m *MailStore) LoadConfig(cfg *config.Config) error {
	domainProps, err := cfg.Query("domain")
	if err != nil {
		return err
	}

	for _, domain := range domainProps {
		if len(domain.Values) < 0 {
			log.Fatalln("Defined domain block without domain name in configuration!")
		}
		domainName := domain.Values[0]
		catchAll, _ := cfg.QuerySingleSub("catch-all 0", domain.Block)

		m.Domains[domainName] = Domain{
			Users:    make(map[string]User),
			CatchAll: catchAll,
		}

		// Get all users
		users, err := cfg.QuerySub("user", domain.Block)
		if err == nil {
			for _, user := range users {
				if len(user.Values) < 0 {
					log.Fatalln("Defined user block without username in configuration!")
				}
				username := user.Values[0]
				boxDir, _ := cfg.QuerySingleSub("box 0", user.Block)
				//TODO Fallback to default box if missing on user
				m.Domains[domainName].Users[username] = User{
					MailboxDir: boxDir,
				}
			}
		}
	}

	return nil
}
