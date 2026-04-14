package congestion

import (
	"time"

	"github.com/quic-go/quic-go/internal/monotime"
	"github.com/quic-go/quic-go/internal/protocol"
)

// The pacer implements a token bucket pacing algorithm.
type ratePacer struct {
	limit           *Limiter
	maxDatagramSize protocol.ByteCount
	lastSentTime    monotime.Time
}

func newRatePacer() *ratePacer {
	p := &ratePacer{
		limit:           NewLimiter(Limit(750_000), maxBurstSizePackets*int(initialMaxDatagramSize)*8),
		maxDatagramSize: initialMaxDatagramSize,
	}
	return p
}

func (p *ratePacer) SentPacket(sendTime monotime.Time, size protocol.ByteCount) {
	p.limit.AllowN(sendTime, int(size*8))
	p.lastSentTime = sendTime
}

func (p *ratePacer) Budget(now monotime.Time) protocol.ByteCount {
	return protocol.ByteCount(p.limit.TokensAt(now) / 8)
}

// TimeUntilSend returns when the next packet should be sent.
// It returns zero if a packet can be sent immediately.
func (p *ratePacer) TimeUntilSend() monotime.Time {
	return p.lastSentTime.Add(5 * time.Millisecond)
	// r := p.limit.ReserveN(monotime.Now(), int(p.maxDatagramSize*8))
	// if !r.OK() {
	// 	// should not happen (maxDatagram smaller than burst size)
	// 	return monotime.Now().Add(protocol.MinPacingDelay)
	// }

	// delay := r.Delay()
	// r.Cancel() // don't consume the tokens yet; just checking

	// if delay <= 0 {
	// 	return 0 // send immediately
	// }
	// return monotime.Now().Add(delay)
}

func (p *ratePacer) SetMaxDatagramSize(s protocol.ByteCount) {
	p.maxDatagramSize = s
}

func (p *ratePacer) SetRate(rate uint) {
	p.limit.SetLimit(Limit(rate))
	// p.limit.SetBurst(maxBurstSizePackets * int(p.maxDatagramSize) * 8)
	p.limit.SetBurst(burst(int(rate), 5*time.Millisecond))
}

func burst(rate int, interval time.Duration) int {
	if interval == 0 {
		interval = time.Millisecond
	}
	f := float64(time.Second.Milliseconds() / interval.Milliseconds())

	return max(2*8*1500, int(2*float64(rate)/f)) // allow at least 2 packet per burst, but allow more if the rate is high enough
}
