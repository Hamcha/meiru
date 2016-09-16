package smtp

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

type Server struct {
	Hostname string
	svsocket net.Listener
}

type serverClient struct {
	socket net.Conn
	server *Server

	// Client info
	Hostname string
}

const MeiruMOTD = "meiru-SMTPd - Welcome!"

func NewServer(bindAddr string, hostname string) (*Server, error) {
	if strings.IndexRune(bindAddr, ':') < 0 {
		bindAddr += ":25"
	}

	serversock, err := net.Listen("tcp", bindAddr)

	return &Server{
		svsocket: serversock,
		Hostname: hostname,
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
	isOpen = true
	for isOpen {
		line, err := readLine(reader)
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
	fmt.Fprintf(c.socket, "220 %s SMTP %s\r\n", c.server.Hostname, MeiruMOTD)
}

func (c *serverClient) DoCommand(line string) {
	cmd := strings.ToUpper(line)
	switch {
	case strings.HasPrefix(cmd, "HELO"):
		// Check for hostname
		hostname := ""
		if len(cmd) > 5 {
			hostname = strings.TrimSpace(cmd[5:])
		}

		if len(hostname) > 0 {
			c.Hostname = hostname

			// Reply with my hostname
			hello := fmt.Sprintf("%s Hello! 😊", c.server.Hostname)
			c.reply(250, hello)
		} else {
			// No hostname provided, scold the ruffian
			c.reply(501, "No HELO hostname provided")
		}
	case strings.HasPrefix(cmd, "EHLO"):
		// Not currently supported
		c.reply(502, "EHLO is not supported yet 😟")
	case strings.HasPrefix(cmd, "QUIT"):
		c.reply(221, "Have a nice day!")
		return false
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

func readLine(reader *bufio.Reader) (string, error) {
	var err error
	line := ""

	for {
		var curline string
		curline, err = reader.ReadString('\n')
		if err != nil {
			break
		}
		line += curline
		if len(curline) > 1 && curline[len(curline)-2] == '\r' {
			break
		} else {
			chr, err := reader.ReadByte()
			if err != nil || chr == '\r' {
				break
			}
			line += string(chr)
		}
	}
	if len(line) > 1 {
		line = line[0 : len(line)-2]
	}
	return line, err
}
