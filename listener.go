package quick

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/quic-go/quic-go"
)

type listener struct {
	conn       *net.UDPConn
	quicServer *quic.Listener
}

var _ net.Listener = &listener{}

// Listen creates a QUIC listener on the given network interface
func Listen(network, laddr string, tlsConfig *tls.Config, quicConfig *quic.Config) (net.Listener, error) {
	udpAddr, err := net.ResolveUDPAddr(network, laddr)
	if err != nil {
		return nil, &net.OpError{Op: "listen", Net: network, Source: nil, Addr: nil, Err: err}
	}
	conn, err := net.ListenUDP(network, udpAddr)
	if err != nil {
		return nil, err
	}

	ln, err := quic.Listen(conn, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}
	return &listener{
		conn:       conn,
		quicServer: ln,
	}, nil
}

// Accept waits for and returns the next connection to the listener.
func (s *listener) Accept() (net.Conn, error) {
	return s.AcceptContext(context.Background())
}

// AcceptContext waits for and returns the next connection to the listener.
func (s *listener) AcceptContext(ctx context.Context) (net.Conn, error) {
	conn, err := s.quicServer.Accept(ctx)
	if err != nil {
		return nil, err
	}
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}

	qconn := &Conn{
		conn:   s.conn,
		qconn:  conn,
		stream: stream,
	}

	return qconn, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (s *listener) Close() error {
	return s.quicServer.Close()
}

// Addr returns the listener's network address.
func (s *listener) Addr() net.Addr {
	return s.quicServer.Addr()
}
