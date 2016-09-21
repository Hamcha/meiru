package main

import (
	"errors"
	"net"

	"github.com/boltdb/bolt"
	"github.com/hamcha/meiru/lib/email"
	"github.com/hamcha/meiru/lib/smtp"
)

var (
	ErrSQCannotResolveDomain      = errors.New("sendqueue err: Cannot resolve remote mail server")
	ErrSQCannotConnectToRemote    = errors.New("sendqueue err: Cannot connect to remote mail server")
	ErrSQCommunicationErrorRemote = errors.New("sendqueue err: Communication error while talking to remote mail server")
)

type SendQueue struct {
	db       *bolt.DB
	inbound  chan sqInboundMailData
	outbound chan sqOutboundMailData

	Hostname string
}

type sqInboundMailData struct {
	Sender    string
	Recipient string
	Data      *string
}

type sqOutboundMailData struct {
	Sender       string
	Recipient    string
	RemoteDomain string
	Data         *string
}

func NewSendQueue(db *bolt.DB, hostname string) *SendQueue {
	return &SendQueue{
		db:       db,
		inbound:  make(chan sqInboundMailData),
		outbound: make(chan sqOutboundMailData),

		Hostname: hostname,
	}
}

func (s *SendQueue) QueueMail(envelope smtp.ServerEnvelope) {
	var inboundMails []sqInboundMailData
	var outboundMails []sqOutboundMailData

	for _, recipient := range envelope.Recipients {
		if envelope.Client.IsAddressInternal(recipient) {
			inboundMails = append(inboundMails, sqInboundMailData{
				Sender:    envelope.Sender,
				Recipient: recipient,
				Data:      &envelope.Data,
			})
		} else {
			_, host := email.SplitAddress(recipient)
			remoteServer, err := getRemoteServerAddr(host)
			if err != nil {
				s.HandleDeliveryError(envelope.Sender, err)
			}
			outboundMails = append(outboundMails, sqOutboundMailData{
				Sender:       envelope.Sender,
				Recipient:    recipient,
				RemoteDomain: remoteServer,
				Data:         &envelope.Data,
			})
		}
	}

	go func() {
		for _, mail := range inboundMails {
			s.inbound <- mail
		}
	}()
	go func() {
		for _, mail := range outboundMails {
			s.outbound <- mail
		}
	}()
}

func (s *SendQueue) SaveIntenalMail(data sqInboundMailData) error {
	//TODO
	return nil
}

func (s *SendQueue) SendExternalMail(data sqOutboundMailData) error {
	client, err := smtp.NewClient(data.RemoteDomain)
	if err != nil {
		return ErrSQCannotConnectToRemote
	}
	if err = client.Greet(s.Hostname); err != nil {
		return ErrSQCommunicationErrorRemote
	}
	if err = client.SetSender(data.Sender); err != nil {
		return ErrSQCommunicationErrorRemote
	}
	if err = client.AddRecipient(data.Recipient); err != nil {
		return ErrSQCommunicationErrorRemote
	}
	if err = client.SendData(*data.Data); err != nil {
		return ErrSQCommunicationErrorRemote
	}

	client.Close()
	return nil
}

func (s *SendQueue) HandleDeliveryError(sender string, err error) error {
	//TODO
	return nil
}

func (s *SendQueue) Serve() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err, _ = r.(error)
		}
	}()

	for {
		select {
		case inboundMail := <-s.inbound:
			err := s.SaveIntenalMail(inboundMail)
			if err != nil {
				s.HandleDeliveryError(inboundMail.Sender, err)
			}
		case outboundMail := <-s.outbound:
			err := s.SendExternalMail(outboundMail)
			if err != nil {
				s.HandleDeliveryError(outboundMail.Sender, err)
			}
		}
	}

	//return nil
}

func getRemoteServerAddr(host string) (string, error) {
	mx, err := net.LookupMX(host)
	if err != nil {
		return "", err
	}
	//TODO Return array and order by preference
	if len(mx) < 1 {
		return "", ErrSQCannotResolveDomain
	}
	return mx[0].Host, nil
}
