package managesieve

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
)

type Client interface {
	Close() error

	RawCmd(format string, args ...any) ([]string, string, error)

	Login(user, pass string) (string, error)
	StartTLS() (string, error)
	Logout() error
	Capability() (string, error)
	HaveSpace(name string, size int64) (string, error)
	ListScripts() ([]Script, string, error)
}

type baseClient struct {
	capabilities *Capabilities
	conn         net.Conn
	debugLog     Logger
	text         *textproto.Conn

	serverName string
}

func NewClient(addr string, options ...Option) (Client, error) {
	params, err := newParameters(options...)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", addr, params.dialTimeout)
	if err != nil {
		return nil, err
	}

	client := &baseClient{
		conn:       conn,
		debugLog:   params.logger,
		serverName: strings.Split(addr, ":")[0],
		text:       textproto.NewConn(conn),
	}
	defer func() {
		if err != nil {
			client.text.Close()
			client.conn.Close()
		}
	}()

	content, _, err := client.readResponse()
	if err != nil {
		return nil, err
	}

	client.capabilities, err = ParseCapabilities(content)

	return client, err
}

// readResponse consumes all server messages until the next response is found
func (client *baseClient) readResponse() ([]string, string, error) {
	result := []string{}
	var line string
	var response string
	var err error
	for {
		line, err = client.text.ReadLine()
		if err != nil {
			break
		}
		client.debugLog.Printf("<<< %s\n", line)
		// TODO ensure we parse response messages that contain multiple lines (e.g., PUTSCRIPT warnings)
		// TODO parse tags (e.g. from NOOP)
		if strings.HasPrefix(line, "OK") {
			response, err = strconv.Unquote(strings.TrimSpace(line[2:]))
			break
		}
		if strings.HasPrefix(line, "NO") {
			response, _ := strconv.Unquote(strings.TrimSpace(line[2:]))
			err = errors.New(response)
			break
		}
		if strings.HasPrefix(line, "BYE") {
			response, _ := strconv.Unquote(strings.TrimSpace(line[3:]))
			err = errors.New(response)
			break
		}

		result = append(result, line)
	}

	return result, response, err
}

func (client *baseClient) RawCmd(format string, args ...any) ([]string, string, error) {
	client.debugLog.Printf(fmt.Sprintf(">>> %s\n", format), args...)
	id, err := client.text.Cmd(format, args...)
	if err != nil {
		return nil, "", err
	}

	client.text.StartResponse(id)
	defer client.text.EndResponse(id)

	return client.readResponse()
}

func (client *baseClient) Close() error {
	return client.text.Close()
}

// rfc5804 commands

// Login authenticates the user on the server (AUTHENTICATE command)
//
// TODO SASL support
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.1
func (client *baseClient) Login(user, pass string) (string, error) {
	auth := base64.StdEncoding.EncodeToString([]byte(user + "\x00" + user + "\x00" + pass))
	_, response, err := client.RawCmd(`%s "PLAIN" %s`, Authenticate, strconv.Quote(auth))
	return response, err
}

// StartTLS executes a TLS negotiation with the server, and upgrades the client
// connection to use the encrypted connection
//
// TODO store the TLS state on the client (similar to the upstream SMTP client)
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.2
func (client *baseClient) StartTLS() (string, error) {
	_, _, err := client.RawCmd(StartTLS)
	if err != nil {
		return "", err
	}

	client.conn = tls.Client(client.conn, &tls.Config{
		ServerName: client.serverName,
	})
	client.text = textproto.NewConn(client.conn)

	content, message, err := client.readResponse()
	if err != nil {
		return message, err
	}

	client.capabilities, err = ParseCapabilities(content)

	return message, err
}

// Logout terminates the current session with the server
//
// The connection is not closed here; it must be done using `Client.Close`
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.3
func (client *baseClient) Logout() error {
	_, _, err := client.RawCmd(Logout)
	return err
}

// Capability requests the server information about its capabilities
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.4
func (client *baseClient) Capability() (string, error) {
	content, message, err := client.RawCmd(Capability)
	if err != nil {
		return message, err
	}

	client.capabilities, err = ParseCapabilities(content)

	return message, err
}

// HaveSpace queries the server if there's enough space for a script
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.5
func (client *baseClient) HaveSpace(name string, size int64) (string, error) {
	_, message, err := client.RawCmd(`%s "%s" %d`, HaveSpace, name, size)
	return message, err
}

// PutScript submits the script to the server
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.6
func (client *baseClient) PutScript(name string, data []byte) (string, error) {
	data = EnsureCRLF(data)

	_, message, err := client.RawCmd(`%s "%s" {%d+}\r\n%s`, PutScript, name, len(data), string(data))

	return message, err
}

// ListScripts lists all scripts the user have saved on the server
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.7
func (client *baseClient) ListScripts() ([]Script, string, error) {
	content, message, err := client.RawCmd(ListScripts)
	if err != nil {
		return nil, message, err
	}

	scripts := make([]Script, len(content))

	for index, entry := range content {
		if strings.HasSuffix(entry, "ACTIVE") {
			scripts[index].Active = true
		}

		scripts[index].Name, err = strconv.Unquote(strings.TrimSuffix(entry, " ACTIVE"))
		if err != nil {
			break
		}
	}

	return scripts, message, err
}

// SetActive sets a script as active in the server, i.e. the script that'll be
// executed on incoming events. An empty string disables the current active one,
// effectively disabling sieve filtering.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.8
func (client *baseClient) SetActive(name string) (string, error) {
	_, message, err := client.RawCmd(`%s "%s"`, SetActive, name)

	return message, err
}

// GetScript retrieves the contents of a script
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.9
func (client *baseClient) GetScript(name string) ([]string, string, error) {
	content, message, err := client.RawCmd(`%s "%s"`, GetScript, name)
	if err != nil {
		return nil, message, err
	}

	// TODO capture and validate the length of the script response
	matches, err := regexp.MatchString(`\{([0-9]+)\}`, content[0])
	if err != nil {
		return nil, message, err
	}

	if !matches {
		return nil, message, fmt.Errorf("unexpected response content format")
	}

	return content[1:], message, err
}

// DeleteScript removes a script. Active scripts cannot be removed, and must be
// disabled using `SetActive` prior to deletion.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.10
func (client *baseClient) DeleteScript(name string) (string, error) {
	_, message, err := client.RawCmd(`%s "%s"`, DeleteScript, name)

	return message, err
}

// RenameScript renames a script. Errors either if the old script doesn't exist
// or the new name is already in use. An active script will remain active after
// the rename.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.11
func (client *baseClient) RenameScript(old, new string) (string, error) {
	_, message, err := client.RawCmd(`%s "%s" "%s"`, RenameScript, old, new)

	return message, err
}

// CheckScript validates the script content without storing it. It'll check both
// syntax and extension support.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.12
func (client *baseClient) CheckScript(data []byte) (string, error) {
	data = EnsureCRLF(data)

	_, message, err := client.RawCmd(`%s {%d+}\r\n%s`, CheckScript, len(data), string(data))

	return message, err
}

// NoOp does nothing beyond requesting an OK response. This is useful to keep
// the connection alive or to re-sync the communication. If a tag is provided
// then the server will include it on the answer.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.13
func (client *baseClient) NoOp(tag string) (string, error) {
	if len(tag) > 0 {
		tag = fmt.Sprintf(` "%s"`, tag)
	}

	_, message, err := client.RawCmd(`%s%s`, NoOp, tag)

	return message, err
}

// Unauthenticate returns the server to its non-authenticated state. Not all
// servers implement this extension.
//
// see https://www.rfc-editor.org/rfc/rfc5804#section-2.14.1
func (client *baseClient) Unauthenticate() (string, error) {
	_, message, err := client.RawCmd(Unauthenticate)

	return message, err
}
