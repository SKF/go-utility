package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"

	"github.com/SKF/go-utility/log"
)

const defaultMaxRetries = 2

// TCP is an TCP client structure supporting http operations for the datadog api
type TCP struct {
	conn       net.Conn
	maxRetries int
	host       string
	port       string
	apiKey     string
	useSsl     bool
}

// NewTCPClient creates a new http client for the datadog api
func NewTCPClient(host string, port string, apiKey string, useSsl bool) *TCP {
	client := &TCP{
		host:       host,
		port:       port,
		apiKey:     apiKey,
		maxRetries: defaultMaxRetries,
		useSsl:     useSsl,
	}

	return client
}

// WithMaxRetries configures the max allowed retries to post to the api
func (c *TCP) WithMaxRetries(maxRetries int) *TCP {
	c.maxRetries = maxRetries
	return c
}

// URL returns the host:port for the client to connect to
func (c *TCP) URL() string {
	return fmt.Sprintf("%s:%s", c.host, c.port)
}

// Connect attempts to establish a tcp connection to the datadog api
func (c *TCP) Connect() (err error) {
	url := c.URL()
	if c.conn, err = net.Dial("tcp", url); err != nil {
		return err
		// errors.Wrap(err, "failed to connect to datadog api").
		// 	WithParams("url", url)
	}

	// configure ssl connection if needed
	if c.useSsl {
		sslConfig := &tls.Config{ServerName: c.host}

		// prep the ssl connection and perform the initial handshake
		sslConn := tls.Client(c.conn, sslConfig)
		if err = sslConn.Handshake(); err != nil {
			return err
			// errors.Wrap(err, "failed initial ssl handshake to datadog api").
			// 	WithParams("host", c.host)
		}

		// set the ssl connection as the active connection
		c.conn = sslConn
	}

	return nil
}

// Disconnect tries to disconnect from the datadog api
func (c *TCP) Disconnect() (err error) {
	if c.conn == nil {
		return
	}

	// try to disconnect
	if err = c.conn.Close(); err != nil {
		return err
		// errors.Wrap(err, "failed to disconnect from the datadog api").
		// 	WithParams("url", c.Url())
	}

	return
}

// PostLogEntry tries to post the log entry to the DataDog lambda ingestion api
func (c *TCP) PostLogEntry(logEntry interface{}) (err error) {
	// make sure we have a connection
	if c.conn == nil {
		if err = c.Connect(); err != nil {
			return
			// errors.Wrap(err, "failed to post to datadog api - no connection could be established").
			// 	WithParams("logEntry", logEntry)
		}

		// we should have a connection now!
		if c.conn == nil {
			return
			// errors.Wrap(err, "failed to post to datadog api - no connection available after connecting").
			// 	WithParams("logEntry", logEntry)
		}
	}

	// marshal the input to a json string
	var jsonBytes []byte
	if jsonBytes, err = json.Marshal(logEntry); err != nil {
		// don't retry on marshal error... can't really be fixed
		return
		// errors.Wrap(err, "failed to marshal logEntry to json").
		// 	WithParams("logEntry", logEntry)
	}

	// datadog api requires us to post a single line string with the API Key as the prefix, and ending with a \n
	logEntryJSONStr := fmt.Sprintf("%s %s\n", c.apiKey, string(jsonBytes))

	// try to send the json data to the api, with possible retries
	numRetries := 0
	maxRetries := c.maxRetries

	// don't allow negative values on the retries...
	if maxRetries < 0 {
		maxRetries = 0
	}

	// keep retrying to send and after each retry... disconnect and reconnect to the datadog api
	for numRetries <= maxRetries {
		if _, err = c.conn.Write([]byte(logEntryJSONStr)); err == nil {
			break
		}

		if dcErr := c.Disconnect(); dcErr != nil {
			log.WithError(dcErr).
				WithField("logEntry", logEntry).
				Debugf("failed to disconnect from datadog api during a retry attempt")
		}

		if connErr := c.Connect(); connErr != nil {
			log.WithError(connErr).
				WithField("logEntry", logEntry).
				Debugf("failed to connect to datadog api during a retry attempt")
		}

		numRetries++
	}

	return err
}
