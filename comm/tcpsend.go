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
	if tc.conn == nil {
		err := tc.establishTcpConn()
		if err != nil {
			retryCounter := 0
			retrySleeper, err2 := time.ParseDuration("3s")
			if err != nil {
				return err2
			}
			for err2 != nil || retryCounter >= 3 {
				retryCounter += 1
				time.Sleep(retrySleeper)
				retrySleeper *= 2
				err2 = tc.establishTcpConn()
				if err2 == nil {
					return nil
				}
			}
			return ErrTooManyConnError
		}
	} else {
		return nil
	}
	return nil
}

func (tc *TCPClient) establishTcpConn() error {
	dstAddr, err2 := net.ResolveTCPAddr("tcp", tc.dest)
	if err2 != nil {
		return err2
	}
	tc.mu.Lock()
	tc.conn, err2 = net.DialTCP("tcp", nil, dstAddr)
	tc.mu.Unlock()
	return err2
}

func (tc *TCPClient) Write(data []byte) (int, error) {
	fData := append(data, '\n')
	wr, err2 := tc.conn.Write(fData)
	if err2 != nil {
		retryCounter := 0
		retrySleeper, err3 := time.ParseDuration("3s")
		if err3 != nil {
			return -1, err3
		}
		for err2 != nil {
			if retryCounter <= 9 {
				retryCounter += 1
				time.Sleep(retrySleeper)
				retrySleeper *= 2
				err2 = tc.establishTcpConn()
				if err2 == nil {
					wr, err2 = tc.conn.Write(fData)
				}
			} else {
				return -1, ErrTooManyConnError
			}
		}
	}
	return wr, err2
}
