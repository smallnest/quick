package quick

import (
	"context"
	"crypto/tls"
	"net"

	quic "github.com/quic-go/quic-go"
)

// Dial creates a new QUIC connection
// it returns once the connection is established and secured with forward-secure keys
func Dial(addr string, tlsConfig *tls.Config, quicConfig *quic.Config) (net.Conn, error) {
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		return nil, err
	}

	// DialAddr returns once a forward-secure connection is established
	quicSession, err := quic.DialAddr(context.Background(), addr, tlsConfig, quicConfig)
	if err != nil {
		return nil, err
	}

	stream, err := quicSession.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{
		conn:   udpConn,
		qconn:  quicSession,
		stream: stream,
	}, nil
}
