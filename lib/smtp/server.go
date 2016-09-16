package smtp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var (
	ServerErrExceededMaximumSize = errors.New("server err: Client exceeded data size limit")
)

type Server struct {
	svsocket net.Listener

	// Server info
	Hostname string
	MaxSize  uint64
}

type serverTransaction struct {
	Sender     string
	Recipients []string
	Data       []byte
}

type serverClient struct {
	socket             net.Conn
	server             *Server
	currentTransaction serverTransaction

	// Client info
	Hostname string
}

const MeiruMOTD = "meiru-SMTPd - Welcome!"
const DefaultMaxSize uint64 = 10485760 // 10 MiB

func NewServer(bindAddr string, hostname string) (*Server, error) {
	if strings.IndexRune(bindAddr, ':') < 0 {
		bindAddr += ":25"
	}

	serversock, err := net.Listen("tcp", bindAddr)

	return &Server{
		Hostname: hostname,
		MaxSize:  DefaultMaxSize,

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
		socket: conn,
		server: s,
	}

	// Send greeting
	c.Greet()

	// Wait and listen for commands
	reader := bufio.NewReader(conn)
	isOpen := true
	for isOpen {
		line, err := c.readLine(reader)
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

func (c *serverClient) Greet() {
	fmt.Fprintf(c.socket, "220 %s ESMTP %s\r\n", c.server.Hostname, MeiruMOTD)
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

		// Reply with my hostname
		clientHost, _, _ := net.SplitHostPort(c.socket.RemoteAddr().String())
		hello := fmt.Sprintf("%s Hello %s [%s]! 😊", c.server.Hostname, c.Hostname, clientHost)
		maxsize := fmt.Sprintf("SIZE %d", c.server.MaxSize)
		c.replyMulti(250, []string{hello, "PIPELINING", maxsize})

	// NOOP
	case strings.HasPrefix(cmd, "NOOP"):
		c.reply(250, "OK 👍")

	// QUIT: Close current connection with client
	case strings.HasPrefix(cmd, "QUIT"):
		c.reply(221, "Have a nice day! 🎉")
		return false

	// RSET: Reset current transaction (start from scratch)
	case strings.HasPrefix(cmd, "RSET"):
		c.currentTransaction = serverTransaction{}
		c.reply(250, "All is forgotten")

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

func (c *serverClient) readLine(reader *bufio.Reader) (string, error) {
	var err error
	line := ""

	for {
		var curline string
		curline, err = reader.ReadString('\n')
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
			chr, err = reader.ReadByte()
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
