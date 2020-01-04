package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/SKF/go-utility/log"
)

const (
	defaultMaxRetries = 2
	backoffInMS       = 100
	maxBackoffInMS    = 1000
)

// TCP is an TCP client structure supporting http operations for the datadog api
type TCP struct {
	mutex      sync.RWMutex
	conn       net.Conn
	host       string
	port       string
	useSsl     bool
	apiKey     string
	maxRetries int
}

// NewTCPClient creates a new http client for the datadog api
func NewTCPClient(host string, port string, apiKey string, useSsl bool) *TCP {
	client := &TCP{
		host:       host,
		port:       port,
		useSsl:     useSsl,
		apiKey:     apiKey,
		maxRetries: defaultMaxRetries,
	}

	return client
}

// WithMaxRetries configures the max allowed retries to post to the api.
// Number of retries can't be negative
func (c *TCP) WithMaxRetries(maxRetries int) *TCP {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if maxRetries >= 0 {
		c.maxRetries = maxRetries
	}
	return c
}

// Connect attempts to establish a tcp connection to the datadog api
func (c *TCP) Connect() (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn != nil {
		return
	}

	url := fmt.Sprintf("%s:%s", c.host, c.port)
	if c.conn, err = net.Dial("tcp", url); err != nil {
		return errors.Wrapf(err, "failed to connect to datadog api on [%s]", url)
	}

	// configure ssl connection if needed
	if c.useSsl {
		sslConfig := &tls.Config{ServerName: c.host}

		// prep the ssl connection and perform the initial handshake
		sslConn := tls.Client(c.conn, sslConfig)
		if err = sslConn.Handshake(); err != nil {
			return errors.Wrap(err, "failed initial ssl handshake to datadog api")
		}

		// set the ssl connection as the active connection
		c.conn = sslConn
	}

	return
}

// Disconnect tries to disconnect from the datadog api
func (c *TCP) Disconnect() (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.conn == nil {
		return
	}

	if err = c.conn.Close(); err != nil {
		return errors.Wrap(err, "failed to disconnect from the datadog api")
	}

	return
}

// Reconnect tries to reconnect to the Datadog API
func (c *TCP) Reconnect() {
	if err := c.Disconnect(); err != nil {
		log.WithError(err).
			Debugf("failed to disconnect from datadog api during a retry attempt")
	}

	if err := c.Connect(); err != nil {
		log.WithError(err).
			Debugf("failed to connect to datadog api during a retry attempt")
	}
}

// PostLogEntry tries to post the log entry to the Datadog API
func (c *TCP) PostLogEntry(logEntry interface{}) (err error) {
	if err = c.Connect(); err != nil {
		return errors.Wrap(err, "failed to connect")
	}

	jsonBytes, err := json.Marshal(logEntry)
	if err != nil {
		return errors.Wrap(err, "failed to marshal logEntry to json")
	}

	// datadog api requires us to post a single line string with the API Key as the prefix, and ending with a \n
	logEntryJSONStr := fmt.Sprintf("%s %s\n", c.apiKey, string(jsonBytes))

	// keep retrying to send and after each retry... disconnect and reconnect to the datadog api
	numRetries := 0
	for numRetries <= c.maxRetries {
		c.mutex.RLock()
		if _, err = c.conn.Write([]byte(logEntryJSONStr)); err == nil {
			c.mutex.RUnlock()
			break
		}
		c.mutex.RUnlock()

		numRetries++

		backoff := backoffInMS * numRetries
		if backoff > maxBackoffInMS {
			backoff = maxBackoffInMS
		}
		time.Sleep(time.Millisecond * time.Duration(backoff))

		c.Reconnect()
	}

	return err
}
