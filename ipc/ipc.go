package ipc

import (
	"log"
	"bufio"
	"bytes"
	"io/ioutil"
	"time"
	"net"
	"os"
)

const (
	bspwmSocket = "BSPWM_SOCKET"
	defaultBspwmSocket = "/tmp/bspwm_0_0-socket"
)

type Subscriber struct {
	Scanner *bufio.Scanner
	conn    *net.UnixConn
}

func (s *Subscriber) Close() error {
	return s.conn.Close()
}

func NewSubscriber() (*Subscriber, error) {
	conn, err := newConn()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(conn)
	_, err = conn.Write(buildPayload("subscribe", "report"))
	if err != nil {
		return nil, err
	}

	return &Subscriber{
		Scanner: scanner,
		conn:    conn,
	}, nil
}

func newConn() (*net.UnixConn, error) {
	socketPath := os.Getenv(bspwmSocket)
	if len(socketPath) == 0 {
		socketPath = defaultBspwmSocket
	}
	return retry(10, 2*time.Second, func() (*net.UnixConn, error) {
		raddr, err := net.ResolveUnixAddr("unix", socketPath)
		if err != nil {
			return nil, err
		}
		return net.DialUnix(raddr.Network(), nil, raddr)
	})
}

func buildPayload(cmd ...string) []byte {
	var buffer bytes.Buffer

	for i := range cmd {
		buffer.WriteString(cmd[i])
		buffer.WriteByte('\x00')
	}

	return buffer.Bytes()
}

func Send(cmd ...string) (response []byte, err error) {
	conn, err := newConn()
	if err != nil {
		return response, err
	}
	defer conn.Close()

	_, err = conn.Write(buildPayload(cmd...))
	if err != nil {
		return response, err
	}

	return ioutil.ReadAll(conn)
}

func retry(attempts int, sleep time.Duration, f func() (*net.UnixConn, error)) (*net.UnixConn, error) {
    var raddr *net.UnixConn
    var err error
    for i := 0; ; i++ {
	raddr, err = f()
        if err == nil {
            return raddr, nil
        }

        if i >= (attempts - 1) {
            break
        }

        log.Println("Failed to open BSPWM_SOCKET:", err)

        time.Sleep(sleep)
    }

    log.Println("Unable to open BSPWM_SOCKET after", attempts, "attempts, giving up")

    return nil, err
}
