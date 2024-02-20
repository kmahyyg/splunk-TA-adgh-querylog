package comm

import (
	"errors"
	"net"
	"sync"
	"time"
)

type TCPClient struct {
	o    sync.Once
	conn *net.TCPConn
	mu   *sync.Mutex
	dest string
}

var (
	ErrDestNotSet       = errors.New("destination not set")
	ErrTooManyConnError = errors.New("too many connection errors")
)

func (tc *TCPClient) SetDest(dest string) {
	tc.o.Do(func() {
		tc.mu = &sync.Mutex{}
		tc.mu.Lock()
		tc.dest = dest
		tc.mu.Unlock()
	})
}

func (tc *TCPClient) Connect() error {
	if tc.dest == "" {
		return ErrDestNotSet
	}
	if tc.conn != nil {
		_, err := tc.conn.Write([]byte{0x00})
		if err != nil {
			dstAddr, err2 := net.ResolveTCPAddr("tcp", tc.dest)
			if err2 != nil {
				return err2
			}
			retryCounter := 0
			retrySleeper, err2 := time.ParseDuration("3s")
			if err != nil {
				return err2
			}
			tc.mu.Lock()
			tc.conn, err2 = net.DialTCP("tcp", nil, dstAddr)
			tc.mu.Unlock()
			for err2 != nil || retryCounter >= 3 {
				retryCounter += 1
				time.Sleep(retrySleeper)
				retrySleeper *= 2
				tc.mu.Lock()
				tc.conn, err2 = net.DialTCP("tcp", nil, dstAddr)
				tc.mu.Unlock()
				if err2 == nil {
					return nil
				}
			}
			return ErrTooManyConnError
		}
	}
	return nil
}

func (tc *TCPClient) Write(data []byte) (int, error) {
	err := tc.Connect()
	if err != nil {
		return -1, err
	}
	return tc.conn.Write(data)
}
