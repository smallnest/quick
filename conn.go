package quick

import (
	"context"
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

var _ net.Conn = &Conn{}

// Conn is a generic quic connection implements net.Conn.
type Conn struct {
	conn    *net.UDPConn
	session quic.Session

	receiveStream quic.Stream
	sendStream    quic.Stream
}

func newConn(sess quic.Session, conn *net.UDPConn) (*Conn, error) {
	stream, err := sess.OpenStream()
	if err != nil {
		return nil, err
	}
	return &Conn{
		conn:       conn,
		session:    sess,
		sendStream: stream,
	}, nil
}

// Read implements the Conn Read method.
func (c *Conn) Read(b []byte) (int, error) {
	if c.receiveStream == nil {
		var err error
		c.receiveStream, err = c.session.AcceptStream(context.Background())
		// TODO: check stream id
		if err != nil {
			return 0, err
		}
		// quic.Stream.Close() closes the stream for writing
		err = c.receiveStream.Close()
		if err != nil {
			return 0, err
		}
	}

	return c.receiveStream.Read(b)
}

// Write implements the Conn Write method.
func (c *Conn) Write(b []byte) (int, error) {
	return c.sendStream.Write(b)
}

// LocalAddr returns the local network address.
func (c *Conn) LocalAddr() net.Addr {
	return c.session.LocalAddr()
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.session.RemoteAddr()
}

// Close closes the connection.
func (c *Conn) Close() error {
	if c.sendStream != nil {
		c.sendStream.Close()
	}
	if c.receiveStream.Close() != nil {
		c.receiveStream.Close()
	}
	return c.conn.Close()
}

// SetDeadline sets the deadline associated with the listener. A zero time value disables the deadline.
func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

// SetReadBuffer sets the size of the operating system's receive buffer associated with the connection.
func (c *Conn) SetReadBuffer(bytes int) error {
	return c.conn.SetReadBuffer(bytes)
}

// SetWriteBuffer sets the size of the operating system's transmit buffer associated with the connection.
func (c *Conn) SetWriteBuffer(bytes int) error {
	return c.conn.SetWriteBuffer(bytes)
}
