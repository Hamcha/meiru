package imap

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	svsocket net.Listener

	Hostname string
}

type serverClient struct {
	socket        net.Conn
	server        *Server
	reader        *bufio.Reader
	authenticated bool
	authName      string
}

func NewServer(bindAddr string) (*Server, error) {
	if strings.IndexRune(bindAddr, ':') < 0 {
		bindAddr += ":143"
	}

	serversock, err := net.Listen("tcp", bindAddr)

	return &Server{
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
		authenticated: false,
	}

	clientHost, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Send greeting
	fmt.Fprintf(c.socket, "* OK meiru-IMAPd Ready for operation, %s! \r\n", clientHost)

	// Wait and listen for commands
	c.reader = bufio.NewReader(conn)
	isOpen := true
	for isOpen {
		line, err := c.readLine()
		if err != nil {
			if err != io.EOF {
				log.Printf("[IMAPd] Read error from client: %s\r\n", err.Error())
			}
			return
		}

		isOpen = c.DoCommand(line)
	}

	c.Close()
}

func (c *serverClient) DoCommand(line string) bool {
	// Cleanup line
	line = strings.TrimSpace(line)

	// Get tag
	tagSep := strings.IndexByte(line, ' ')
	if tagSep < 0 {
		c.reply("*", "BAD invalid tag")
		return true
	}
	tag := line[:tagSep]
	cmd := strings.ToUpper(line[tagSep+1:])

	switch {

	// NOOP
	case strings.HasPrefix(cmd, "NOOP"):
		c.reply(tag, "OK ..well this was a waste of bandwidth.")

	// CAPABILITY: List supported capabilities/extensions
	case strings.HasPrefix(cmd, "CAPABILITY"):
		c.replyMulti(tag, []string{
			"CAPABILITY IMAP4rev1",
			"OK It's not you, it's the mail server!",
		})

	// LOGIN: Authenticate client
	case strings.HasPrefix(cmd, "LOGIN"):
		//TODO

	// LOGOUT: Close current connection with client
	case strings.HasPrefix(cmd, "LOGOUT"):
		c.reply("*", "BYE Have a nice day! 🎉")
		c.reply(tag, "OK Logged out")
		return false

	// Command not recognized
	default:
		c.reply(tag, "BAD Command not recognized 😕")
	}

	return true
}

func (c *serverClient) Close() {
	c.socket.Close()
}

func (c *serverClient) replyMulti(tag string, lines []string) {
	linecount := len(lines)
	if linecount > 1 {
		for _, line := range lines[0 : linecount-1] {
			fmt.Fprintf(c.socket, "* %s\r\n", line)
		}
	}
	c.reply(tag, lines[linecount-1])
}

func (c *serverClient) reply(tag string, line string) {
	fmt.Fprintf(c.socket, "%s %s\r\n", tag, line)
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
