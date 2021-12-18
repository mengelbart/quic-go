package quic

import (
	"fmt"
	"sync"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

type datagramQueue struct {
	mx            sync.Mutex
	nextFrame     *wire.DatagramFrame
	nextFrameSize protocol.ByteCount

	sendQueue chan *wire.DatagramFrame
	rcvQueue  chan []byte

	closeErr error
	closed   chan struct{}

	hasData func()

	logger  utils.Logger
	version protocol.VersionNumber
}

func newDatagramQueue(hasData func(), logger utils.Logger, v protocol.VersionNumber) *datagramQueue {
	return &datagramQueue{
		hasData:       hasData,
		sendQueue:     make(chan *wire.DatagramFrame, protocol.DatagramSendQueueLen),
		nextFrame:     nil,
		nextFrameSize: protocol.InvalidByteCount,
		rcvQueue:      make(chan []byte, protocol.DatagramRcvQueueLen),
		closed:        make(chan struct{}),
		logger:        logger,
		version:       v,
	}
}

// AddAndWait queues a new DATAGRAM frame for sending.
// It blocks until the frame has been dequeued.
func (h *datagramQueue) AddAndWait(f *wire.DatagramFrame) error {
	h.mx.Lock()
	defer h.mx.Unlock()

	select {
	case h.sendQueue <- f:
		if h.nextFrame == nil {
			h.nextFrame = <-h.sendQueue
		}
		h.hasData()
	case <-h.closed:
		return h.closeErr
	default:
		return fmt.Errorf("datagram queue full, dropping packet")
	}

	select {
	case <-h.closed:
		return h.closeErr
	default:
		return nil
	}
}

// Get dequeues a DATAGRAM frame for sending.
func (h *datagramQueue) Get() *wire.DatagramFrame {
	h.mx.Lock()
	defer h.mx.Unlock()

	next := h.nextFrame
	select {
	case h.nextFrame = <-h.sendQueue:
	default:
		h.nextFrame = nil
	}
	return next
}

func (h *datagramQueue) NextFrameSize() protocol.ByteCount {
	h.mx.Lock()
	defer h.mx.Unlock()
	if h.nextFrame == nil {
		return 0
	}
	return h.nextFrame.Length(h.version)
}

// HandleDatagramFrame handles a received DATAGRAM frame.
func (h *datagramQueue) HandleDatagramFrame(f *wire.DatagramFrame) {
	data := make([]byte, len(f.Data))
	copy(data, f.Data)
	select {
	case h.rcvQueue <- data:
	default:
		h.logger.Debugf("Discarding DATAGRAM frame (%d bytes payload)", len(f.Data))
	}
}

// Receive gets a received DATAGRAM frame.
func (h *datagramQueue) Receive() ([]byte, error) {
	select {
	case data := <-h.rcvQueue:
		return data, nil
	case <-h.closed:
		return nil, h.closeErr
	}
}

func (h *datagramQueue) CloseWithError(e error) {
	h.closeErr = e
	close(h.closed)
}
