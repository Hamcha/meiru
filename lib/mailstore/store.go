package mailstore

import (
	"strings"

	"github.com/hamcha/meiru/lib/email"
	"github.com/hamcha/meiru/lib/errors"
)

type InboundMailData struct {
	Recipient  string
	RealSender string
	MailData   string
}

var (
	ErrMSNoValidRecipient = errors.NewType(ErrSrcMailstore, "could not deliver mail to a valid recipient")
)

func (m *MailStore) Save(mail InboundMailData) *errors.Error {
	_, err := m.getUser(mail.Recipient)
	if err != nil {
		return err
	}

	return nil
}

func (m *MailStore) getUser(address string) (User, *errors.Error) {
	// Parse recipient
	name, domain := email.SplitAddress(address)

	// Try get domain
	dom, ok := m.Domains[strings.ToLower(domain)]
	if !ok {
		return User{}, errors.NewError(ErrMSNoValidRecipient).WithInfo("Delivery failure reason: domain '%s' is not internal", domain)
	}

	// Try get user
	user, ok := dom.Users[strings.ToLower(name)]
	if !ok {
		// If there is no such user, try catch-all
		if dom.CatchAll != "" {
			user, ok = dom.Users[dom.CatchAll]
			if !ok {
				return User{}, errors.NewError(ErrMSNoValidRecipient).WithInfo("Delivery failure reason: Catch-all '%s@%s' does not map to a valid user", dom.CatchAll, domain)
			}
		} else {
			return User{}, errors.NewError(ErrMSNoValidRecipient).WithInfo("Delivery failure reason: Could not find valid user or catch-all")
		}
	}

	return user, nil
}
