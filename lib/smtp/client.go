package smtp

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/hamcha/meiru/lib/errors"
)

type Client struct {
	conn      net.Conn
	reader    *bufio.Reader
	ServerExt []ClientServerExt
}

type clientServerReply struct {
	Code int
	Text string
}

type ClientServerExt struct {
	Name   string
	Params []string
}

var (
	ErrSrcClient errors.ErrorSource = "smtpc"

	ClientErrInvalidServerResponse = errors.NewType(ErrSrcClient, "invalid response from server")
	ClientErrReceivedServerError   = errors.NewType(ErrSrcClient, "received error from server")
	ClientErrNoServerResponse      = errors.NewType(ErrSrcClient, "no response from server")
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

	client := Client{
		conn:   sock,
		reader: reader,
	}

	return &client, err
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

	// Check if it's the greeting
	if resp[0].Code == 220 {
		// Ignore and get next response
		resp, err = c.getReplies()
		if err != nil {
			return err
		}
	}

	// Check if the greet was not successful
	if resp[0].Code != 250 {
		// Fall back to HELO
		c.cmd("HELO %s", host)
		resp, err = c.getReplies()
		if err != nil {
			return err
		}
		if err = getResponseError(resp[0]); err != nil {
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

func (c *Client) SetSender(addr string) error {
	c.cmd("MAIL FROM: <%s>", addr)
	resp, err := c.getReplies()
	if err != nil {
		return err
	}

	return getResponseError(resp[0])
}

func (c *Client) AddRecipient(addr string) error {
	c.cmd("RCPT TO: <%s>", addr)
	resp, err := c.getReplies()
	if err != nil {
		return err
	}

	return getResponseError(resp[0])
}

func (c *Client) SendData(data string) error {
	c.cmd("DATA")
	resp, err := c.getReplies()
	if err != nil {
		return err
	}

	if resp[0].Code != 354 {
		return errors.NewError(ClientErrReceivedServerError).WithInfo("Recv error: %d %s", resp[0].Code, resp[0].Text)
	}

	fmt.Fprintf(c.conn, "%s\r\n.\r\n", data)
	resp, err = c.getReplies()
	if err != nil {
		return err
	}

	return getResponseError(resp[0])
}

func getResponseError(reply clientServerReply) error {
	if reply.Code != 250 {
		return errors.NewError(ClientErrReceivedServerError).WithInfo("Recv error: %d %s", reply.Code, reply.Text)
	}
	return nil
}

func (c *Client) cmd(format string, a ...interface{}) {
	fmt.Fprintf(c.conn, format+"\r\n", a...)
}

func (c *Client) getReplies() ([]clientServerReply, error) {
	var replies []clientServerReply
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
			return replies, errors.NewError(ClientErrInvalidServerResponse).WithInfo("Recv: %s", str)
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
			return replies, errors.NewError(ClientErrInvalidServerResponse).WithError(err)
		}

		replies = append(replies, clientServerReply{
			Code: code,
			Text: str[separatorIndex+1:],
		})
	}

	// No response received?
	if len(replies) == 0 {
		return replies, errors.NewError(ClientErrNoServerResponse)
	}

	return replies, nil
}
