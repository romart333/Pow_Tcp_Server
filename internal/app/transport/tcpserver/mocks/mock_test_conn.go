package mocks

import (
	"errors"
	"net"
	"time"
)

// TestConn is a mock implementation of net.Conn for testing
type TestConn struct {
	ReadBuf       []byte    // Buffer for Read operations
	WriteBuf      []byte    // Buffer for Write operations
	ReadPos       int       // Current position in ReadBuf
	Closed        bool      // Connection closed flag
	Deadline      time.Time // Deadline set for connection
	LocalAddress  net.Addr  // Mock local address
	RemoteAddress net.Addr  // Mock remote address
}

func (c *TestConn) Read(b []byte) (n int, err error) {
	if c.ReadPos >= len(c.ReadBuf) {
		return 0, errors.New("EOF")
	}
	n = copy(b, c.ReadBuf[c.ReadPos:])
	c.ReadPos += n
	return n, nil
}

func (c *TestConn) Write(b []byte) (n int, err error) {
	c.WriteBuf = append(c.WriteBuf, b...)
	return len(b), nil
}

func (c *TestConn) Close() error {
	c.Closed = true
	return nil
}

func (c *TestConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *TestConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *TestConn) SetDeadline(t time.Time) error {
	c.Deadline = t
	return nil
}

func (c *TestConn) SetReadDeadline(t time.Time) error {
	c.Deadline = t
	return nil
}

func (c *TestConn) SetWriteDeadline(t time.Time) error {
	c.Deadline = t
	return nil
}
