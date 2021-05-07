package mitm

import (
	"crypto/tls"
	"io"
	"net"
)

type mitmListener struct {
	conn *tls.Conn
}

func (l *mitmListener) Accept() (net.Conn, error) {
	if l.conn != nil {
		conn := l.conn
		l.conn = nil
		return conn, nil
	} else {
		return nil, io.EOF
	}
}

func (l *mitmListener) Close() error {
	return nil
}

func (l *mitmListener) Addr() net.Addr {
	return nil
}
