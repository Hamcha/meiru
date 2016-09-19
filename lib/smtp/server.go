package smtp

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/mail"
	"strings"

	"github.com/hamcha/meiru/lib/email"
)

var (
	ServerErrExceededMaximumSize = errors.New("server err: Client exceeded data size limit")
)

type ReceivedMailHandler func(e ServerEnvelope)
type AuthRequestHandler func(ServerAuthRequest) bool

type Server struct {
	svsocket net.Listener

	// Server info
	Hostname     string
	MaxSize      uint64
	LocalDomains []string
	RequireAuth  bool

	OnAuthRequest  AuthRequestHandler
	OnReceivedMail ReceivedMailHandler
}

type ServerAuthRequest struct {
	Username string
	Password string
}

type ServerEnvelope struct {
	Client     *serverClient
	Sender     string
	Recipients []string
	Data       []byte
}

type serverClient struct {
	socket          net.Conn
	server          *Server
	reader          *bufio.Reader
	currentEnvelope ServerEnvelope
	greeted         bool
	authenticated   bool
	authName        string

	// Client info
	SourceAddr net.Addr
	Hostname   string
}

const MeiruMOTD = "meiru-SMTPd - Welcome!"
const DefaultMaxSize uint64 = 10485760 // 10 MiB

func NewServer(bindAddr string, hostname string) (*Server, error) {
	if strings.IndexRune(bindAddr, ':') < 0 {
		bindAddr += ":25"
	}

	serversock, err := net.Listen("tcp", bindAddr)

	return &Server{
		Hostname:    hostname,
		MaxSize:     DefaultMaxSize,
		RequireAuth: true,

		svsocket: serversock,
	}, err
}

func (s *Server) ListenAndServe() error {
	// Accept loop
	for {
		// Wait for connection
		conn, err := s.svsocket.Accept()
		if err != nil {
			return err
		}

		go s.handleClient(conn)
	}
}

func (s *Server) Close() {
	s.svsocket.Close()
}

func (s *Server) handleClient(conn net.Conn) {
	c := serverClient{
		socket:        conn,
		server:        s,
		greeted:       false,
		authenticated: false,
		SourceAddr:    conn.RemoteAddr(),
	}

	// Send greeting
	fmt.Fprintf(c.socket, "220 %s ESMTP %s\r\n", c.server.Hostname, MeiruMOTD)

	// Wait and listen for commands
	c.reader = bufio.NewReader(conn)
	isOpen := true
	for isOpen {
		line, err := c.readLine()
		if err != nil {
			if err != io.EOF {
				log.Printf("[SMTPd] Read error from client: %s\r\n", err.Error())
			}
			return
		}

		isOpen = c.DoCommand(line)
	}

	c.Close()
}

func (c *serverClient) DoCommand(line string) bool {
	cmd := strings.ToUpper(line)
	switch {
	// HELO: SMTP HELO (required greeting)
	case strings.HasPrefix(cmd, "HELO"):
		// Check for hostname
		hostname := ""
		if len(line) > 5 {
			hostname = strings.TrimSpace(line[5:])
		}
		if len(hostname) < 1 {
			// No hostname provided, scold the ruffian
			c.reply(501, "No HELO hostname provided")
			break
		}
		c.Hostname = hostname
		c.greeted = true

		// Reply with my hostname
		hello := fmt.Sprintf("%s Hello! 😊", c.server.Hostname)
		c.reply(250, hello)

	// ELHO: ESMTP HELO (w/ extension list)
	case strings.HasPrefix(cmd, "EHLO"):
		// Check for hostname
		hostname := ""
		if len(line) > 5 {
			hostname = strings.TrimSpace(line[5:])
		}
		if len(hostname) < 1 {
			// No hostname provided, scold the ruffian
			c.reply(501, "No EHLO hostname provided")
			break
		}
		c.Hostname = hostname
		c.greeted = true

		// Reply with my hostname
		clientHost, _, _ := net.SplitHostPort(c.SourceAddr.String())
		hello := fmt.Sprintf("%s Hello %s [%s]! 😊", c.server.Hostname, c.Hostname, clientHost)

		// Prepare extension list
		maxsize := fmt.Sprintf("SIZE %d", c.server.MaxSize)
		c.replyMulti(250, []string{hello, "PIPELINING", "SMTPUTF8", "AUTH PLAIN", maxsize})

	// NOOP
	case strings.HasPrefix(cmd, "NOOP"):
		c.reply(250, "OK 👍")

	// QUIT: Close current connection with client
	case strings.HasPrefix(cmd, "QUIT"):
		c.reply(221, "Have a nice day! 🎉")
		return false

	// RSET: Reset current envelope (start from scratch)
	case strings.HasPrefix(cmd, "RSET"):
		c.currentEnvelope = ServerEnvelope{
			Client: c,
		}
		c.reply(250, "All is forgotten")

	// MAIL FROM: Start a envelope and set sender
	case strings.HasPrefix(cmd, "MAIL FROM:"):
		// Reject if we haven't been greeted already
		if !c.greeted {
			c.reply(503, "Rude! 😠 Say HELO/EHLO first!")
			break
		}
		// Reject if there is a envelope already active
		if len(c.currentEnvelope.Sender) > 0 {
			c.reply(503, "An envelope is already open, call RSET if you want to start over")
			break
		}
		// Reject empty addresses
		if len(line) < 11 {
			c.reply(550, "No address specified")
			break
		}
		// Trim whitespace around line and reject garbage
		trimmed := strings.TrimSpace(line[10:])
		if strings.IndexByte(trimmed, '>') < 0 {
			c.reply(555, "Garbage not permitted")
			break
		}
		// Try to parse address
		addr, err := mail.ParseAddress(trimmed)
		if err != nil || !email.IsValidAddress(addr.Address) {
			c.reply(501, "The address you specified is malformed")
			break
		}

		// Check if local address (require auth)
		if c.isAddressInternal(addr.Address) && c.server.RequireAuth {
			// Check if client is authenticated
			if !c.authenticated {
				c.reply(530, "Emails from this domain require authentication. Please authenticate first!")
				break
			} else {
				// Check if authenticated for a different address
				if strings.ToLower(c.authName) != strings.ToLower(addr.Address) {
					errstr := fmt.Sprintf("Authenticated for a different address (%s), use that or authenticate as \"%s\" instead!", c.authName, addr)
					c.reply(530, errstr)
					break
				}
			}
		}

		// Set envelope client if not set
		c.currentEnvelope.Client = c

		// Set address as sender
		c.currentEnvelope.Sender = addr.Address
		c.reply(250, "OK 👍")

	// RCPT TO: Add recipient to envelope
	case strings.HasPrefix(cmd, "RCPT TO:"):
		// Reject if there isn't an active envelope
		if len(c.currentEnvelope.Sender) < 1 {
			c.reply(503, "No envelopes to add recipients to, please start one with MAIL FROM")
			break
		}
		// Reject empty addresses
		if len(line) < 11 {
			c.reply(550, "No address specified")
			break
		}
		// Trim whitespace around line and reject garbage
		trimmed := strings.TrimSpace(line[10:])
		if strings.IndexByte(trimmed, '>') > 0 && !strings.HasSuffix(trimmed, ">") {
			c.reply(555, "Garbage not permitted")
			break
		}
		// Try to parse address
		addr, err := mail.ParseAddress(trimmed)
		if err != nil || !email.IsValidAddress(addr.Address) {
			c.reply(501, "The address you specified is malformed")
			break
		}

		// Ask for AUTH if outbound email
		if !c.authenticated && c.server.RequireAuth {
			c.reply(530, "Outbound emails require authentication. Please authenticate first!")
			break
		}

		// Check for proper auth if necessary
		if c.authenticated && strings.ToLower(c.authName) != strings.ToLower(c.currentEnvelope.Sender) {
			errstr := fmt.Sprintf("Authenticated for a different address (%s) than sender (%s), use that or authenticate as \"%s\" instead!", c.authName, c.currentEnvelope.Sender, c.currentEnvelope.Sender)
			c.reply(530, errstr)
			break
		}

		// Add address to recipients
		c.currentEnvelope.Recipients = append(c.currentEnvelope.Recipients, addr.Address)
		c.reply(250, "OK 👍")

	// DATA: Receive mail data from client
	case strings.HasPrefix(cmd, "DATA"):
		// Reject if there isn't an active envelope
		if len(c.currentEnvelope.Sender) < 1 || len(c.currentEnvelope.Recipients) < 1 {
			c.reply(503, "Please specify both a sender and at least one recipient first")
			break
		}

		// Check for proper auth if necessary
		if c.authenticated && strings.ToLower(c.authName) != strings.ToLower(c.currentEnvelope.Sender) {
			errstr := fmt.Sprintf("Authenticated for a different address (%s) than sender (%s), use that or authenticate as \"%s\" instead!", c.authName, c.currentEnvelope.Sender, c.currentEnvelope.Sender)
			c.reply(530, errstr)
			break
		}

		c.reply(354, "Fire away! End with <CRLF>.<CRLF>")
		_, err := c.readDATA()
		if err != nil {
			log.Printf("[SMTPd] Client read error: %s\r\n", err.Error())
			return false
		}
		c.reply(250, "Your message is on its way! ✈")

	// AUTH: Authenticate client
	case strings.HasPrefix(cmd, "AUTH"):
		parts := strings.Split(strings.TrimSpace(line), " ")
		if len(parts) < 2 {
			c.reply(504, "Please specify the authentication method")
			break
		}
		method := strings.ToUpper(parts[1])
		switch method {
		case "PLAIN":
			b64str := ""
			if len(parts) < 3 {
				c.reply(334, "")
				var err error
				b64str, err = c.readLine()
				if err != nil {
					log.Printf("[SMTPd] Client read error: %s\r\n", err.Error())
					return false
				}
			} else {
				b64str = parts[2]
			}
			data, err := base64.StdEncoding.DecodeString(b64str)
			if err != nil {
				c.reply(501, "That doesn't look like Base64… 🤔")
				break
			}
			user, pass, err := decodePlainResponse(data)
			if err != nil {
				c.reply(501, "The PLAIN auth string is malformed")
				break
			}
			c.authenticated = c.server.OnAuthRequest(ServerAuthRequest{
				Username: user,
				Password: pass,
			})
			if c.authenticated {
				c.authName = user
				c.reply(235, "You're authenticated!")
			} else {
				c.reply(535, "Sorry, I cannot accept those credentials!")
			}
		default:
			c.reply(504, "I don't support that authentication method, sorry! 😟")
		}

	// Command not recognized
	default:
		c.reply(502, "Command not recognized 😕")
	}

	return true
}

func (c *serverClient) Close() {
	c.socket.Close()
}

func (c *serverClient) replyMulti(code int, lines []string) {
	linecount := len(lines)
	if linecount > 1 {
		for _, line := range lines[0 : linecount-1] {
			fmt.Fprintf(c.socket, "%d-%s\r\n", code, line)
		}
	}
	c.reply(code, lines[linecount-1])
}

func (c *serverClient) reply(code int, line string) {
	fmt.Fprintf(c.socket, "%d %s\r\n", code, line)
}

func (c *serverClient) readLine() (string, error) {
	var err error
	line := ""

	for {
		var curline string
		curline, err = c.reader.ReadString('\n')
		if err != nil {
			break
		}
		line += curline
		if uint64(len(line)) > c.server.MaxSize {
			err = ServerErrExceededMaximumSize
			break
		}
		if strings.HasSuffix(curline, "\r\n") {
			break
		} else {
			var chr byte
			chr, err = c.reader.ReadByte()
			if err != nil {
				break
			}
			line += string(chr)
			if strings.HasSuffix(line, "\n\r") {
				break
			}
		}
	}

	return strings.TrimRight(line, "\r\n"), err
}

func (c *serverClient) readDATA() (string, error) {
	data := ""
	checkNext := false
	for {
		line, err := c.readLine()
		if err != nil {
			return data, err
		}

		if len(line) < 1 && !checkNext {
			checkNext = true
		}
		if checkNext && line == "." {
			break
		}
		data += line + "\r\n"
	}

	return strings.TrimRight(data, "\r\n"), nil
}

func (c *serverClient) isAddressInternal(addr string) bool {
	atIndex := strings.LastIndexByte(addr, '@')
	remoteDomain := strings.ToLower(addr[atIndex+1:])
	for _, localDomain := range c.server.LocalDomains {
		if strings.ToLower(localDomain) == remoteDomain {
			return true
		}
	}

	return false
}

func (e *ServerEnvelope) isInternal() bool {
	for _, recp := range e.Recipients {
		if !e.Client.isAddressInternal(recp) {
			return false
		}
	}

	return true
}
