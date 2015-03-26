package replication

import (
	"errors"
	"time"
)

var (
	ErrGetEventTimeout = errors.New("Get event timeout, try get later")
	ErrNeedSyncAgain   = errors.New("Last sync error or closed, try sync and get event again")
	ErrSyncClosed      = errors.New("Sync was closed")
)

type BinlogStreamer struct {
	ch  chan *BinlogEvent
	ech chan error
	err error
}

func (s *BinlogStreamer) GetEvent() (*BinlogEvent, error) {
	// we use a very very long timeout here
	return s.GetEventTimeout(time.Second * 3600 * 24 * 30)
}

// if timeout, ErrGetEventTimeout will returns
func (s *BinlogStreamer) GetEventTimeout(d time.Duration) (*BinlogEvent, error) {
	if s.err != nil {
		return nil, ErrNeedSyncAgain
	}

	select {
	case c := <-s.ch:
		return c, nil
	case s.err = <-s.ech:
		return nil, s.err
	case <-time.After(d):
		return nil, ErrGetEventTimeout
	}
}

func (s *BinlogStreamer) Close() {
	s.CloseWithError(ErrSyncClosed)
}

func (s *BinlogStreamer) CloseWithError(err error) {
	if err == nil {
		err = ErrSyncClosed
	}
	select {
	case s.ech <- err:
	default:
	}
}

func newBinlogStreamer() *BinlogStreamer {
	s := new(BinlogStreamer)

	s.ch = make(chan *BinlogEvent, 1024)
	s.ech = make(chan error, 4)

	return s
}
