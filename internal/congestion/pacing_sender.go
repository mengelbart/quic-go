package congestion

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/quic-go/quic-go/internal/monotime"
	"github.com/quic-go/quic-go/internal/protocol"
)

type pacingSender struct {
	pacer           PacerInt
	pacingRate      atomic.Uint64
	maxDatagramSize protocol.ByteCount
}

func NewPacingSender(initialMaxDatagramSize protocol.ByteCount) *pacingSender {
	ps := &pacingSender{
		maxDatagramSize: initialMaxDatagramSize,
	}
	ps.pacingRate.Store(1_000_000)
	// ps.pacer = newPacer(ps.getPacingRate)
	ps.pacer = newRatePacer()
	return ps
}

func (s *pacingSender) SetPacingRate(rate uint64) {
	s.pacingRate.Store(uint64(rate))
	s.pacer.SetRate(uint(rate))
}

func (s *pacingSender) getPacingRate() Bandwidth {
	return Bandwidth(s.pacingRate.Load())
}

func (s *pacingSender) TimeUntilSend(bytesInFlight protocol.ByteCount) monotime.Time {
	return s.pacer.TimeUntilSend()
}

func (s *pacingSender) HasPacingBudget(now monotime.Time) bool {
	return s.pacer.Budget(now) >= s.maxDatagramSize
}

func (s *pacingSender) OnPacketSent(sentTime monotime.Time, bytesInFlight protocol.ByteCount, packetNumber protocol.PacketNumber, bytes protocol.ByteCount, isRetransmittable bool) {
	s.pacer.SentPacket(sentTime, bytes)
}

func (s *pacingSender) CanSend(bytesInFlight protocol.ByteCount) bool {
	return true
}

func (s *pacingSender) MaybeExitSlowStart() {
}

func (s *pacingSender) OnPacketAcked(number protocol.PacketNumber, ackedBytes protocol.ByteCount, priorInFlight protocol.ByteCount, eventTime monotime.Time) {
}

func (s *pacingSender) OnCongestionEvent(number protocol.PacketNumber, lostBytes protocol.ByteCount, priorInFlight protocol.ByteCount) {
}

func (s *pacingSender) OnRetransmissionTimeout(packetsRetransmitted bool) {
}

func (s *pacingSender) SetMaxDatagramSize(size protocol.ByteCount) {
	if size < s.maxDatagramSize {
		panic(fmt.Sprintf("congestion BUG: decreased max datagram size from %d to %d", s.maxDatagramSize, size))
	}
	s.maxDatagramSize = size
	s.pacer.SetMaxDatagramSize(size)
}

// GetCongestionWindow implements [SendAlgorithmWithDebugInfos].
func (s *pacingSender) GetCongestionWindow() protocol.ByteCount {
	return math.MaxInt64
}

// InSlowStart implements [SendAlgorithmWithDebugInfos].
func (s *pacingSender) InSlowStart() bool {
	return false
}

// InRecovery implements [SendAlgorithmWithDebugInfos].
func (s *pacingSender) InRecovery() bool {
	return false
}

var (
	_ SendAlgorithm               = &pacingSender{}
	_ SendAlgorithmWithDebugInfos = &pacingSender{}
)
