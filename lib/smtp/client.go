package smtp

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	conn      net.Conn
	reader    *bufio.Reader
	ServerExt []ClientServerExt
}

type ClientServerReply struct {
	Code int
	Text string
}

type ClientServerExt struct {
	Name   string
	Params []string
}

var (
	ClientErrInvalidServerResponse = errors.New("smtp client err: invalid response from server")
	ClientErrNoServerResponse      = errors.New("smtp client err: no response from server")
)

func NewClient(host string) (*Client, error) {
	if strings.IndexRune(host, ':') < 0 {
		host += ":25"
	}
	sock, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(sock)

	return &Client{
		conn:   sock,
		reader: reader,
	}, nil
}

func (c *Client) Close() {
	c.cmd("QUIT")
	c.getReplies()
	c.conn.Close()
}

func (c *Client) Greet(host string) error {
	c.cmd("EHLO %s", host)
	resp, err := c.getReplies()
	if err != nil {
		return err
	}

	// No response received?
	if len(resp) == 0 {
		return ClientErrNoServerResponse
	}

	// Check if the greet was not successful
	if resp[0].Code != 250 {
		// Fall back to HELO
		c.cmd("HELO %s", host)
		resp, err = c.getReplies()
		if err != nil {
			return err
		}
	}

	// Check for extensions
	if len(resp) > 1 {
		for _, reply := range resp[1:] {
			// Parse server extension keyword and parameters and add it to the list
			parts := strings.Split(reply.Text, " ")
			ext := ClientServerExt{Name: parts[0]}
			if len(parts) > 1 {
				ext.Params = parts[1:]
			}
			c.ServerExt = append(c.ServerExt, ext)
		}
	}

	return nil
}

func (c *Client) cmd(format string, a ...interface{}) {
	fmt.Fprintf(c.conn, format+"\r\n", a...)
}

func (c *Client) getReplies() ([]ClientServerReply, error) {
	var replies []ClientServerReply
	var hasMore = true

	for hasMore {
		str, err := c.reader.ReadString('\n')
		if err != nil {
			return replies, err
		}

		// Trim \r if present (should be)
		str = strings.TrimRight(str, "\r\n")

		// Get first separator
		spaceSep := strings.IndexByte(str, ' ')
		tickSep := strings.IndexByte(str, '-')

		if spaceSep < 0 && tickSep < 0 {
			return replies, ClientErrInvalidServerResponse
		}

		// If separator is '-' other replies are available
		hasMore = tickSep > 0 && (spaceSep < 0 || tickSep < spaceSep)

		// Create reply struct and add it to the list
		separatorIndex := spaceSep
		if hasMore {
			separatorIndex = tickSep
		}

		code, err := strconv.Atoi(str[0:separatorIndex])
		if err != nil {
			return replies, ClientErrInvalidServerResponse
		}

		replies = append(replies, ClientServerReply{
			Code: code,
			Text: str[separatorIndex+1:],
		})
	}

	return replies, nil
}
