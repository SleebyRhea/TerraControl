package main

import (
	"errors"
	"net"
	"strconv"
)

const (
	defaultPort        = 8080
	defaultMaxCommands = 500
)

// Configuration -
type Configuration struct {
	ip   net.IP
	port int

	hostname   string
	rootprefix string
}

// Port - Return the port in string form (ex :8080)
func (c *Configuration) Port() string {
	p := defaultPort
	if c.port != 0 {
		p = c.port
	}
	return sprintf(":%d", p)
}

// SetPort - Sets the configured port
func (c *Configuration) SetPort(p int) error {
	if p > 65535 || p < 1 {
		return errors.New("Invalid port given")
	}
	if !isPortAvailable(p) {
		return errors.New("Port is already in use")
	}
	c.port = p
	return nil
}

func isPortAvailable(p int) bool {
	// https://coolaj86.com/articles/how-to-test-if-a-port-is-available-in-go/
	ps := strconv.Itoa(p)
	l, err := net.Listen("tcp", ":"+ps)
	if err != nil {
		return false
	}
	_ = l.Close()
	return true
}
