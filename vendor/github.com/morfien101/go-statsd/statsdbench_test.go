package statsd

import (
	"net"
	"testing"
	"time"
	//	"strconv"
	//	unix4ever "github.com/Unix4ever/statsd"
	//	cactus "github.com/cactus/go-statsd-client/statsd"
	//	"github.com/peterbourgon/g2s"
	//	quipo "github.com/quipo/statsd"
	//	ac "gopkg.in/alexcesaro/statsd.v2"
)

const (
	addr        = ":0"
	prefix      = "metricPrefix."
	prefixNoDot = "metricPrefix"
	counterKey  = "foo.bar.counter"
	gaugeKey    = "foo.bar.gauge"
	gaugeValue  = 42
	timingKey   = "foo.bar.timing"
	timingValue = 153 * time.Millisecond
	flushPeriod = 100 * time.Millisecond
)

type logger struct{}

func (logger) Println(v ...interface{}) {}

func BenchmarkComplexDelivery(b *testing.B) {
	inSocket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})
	if err != nil {
		b.Error(err)
	}

	go func() {
		buf := make([]byte, 1500)
		for {
			_, err := inSocket.Read(buf)
			if err != nil {
				return
			}
		}

	}()

	client := NewClient(inSocket.LocalAddr().String(), MetricPrefix("foo."))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Incr("number.requests", 33)
		client.Timing("another.value", 157)
		client.PrecisionTiming("response.time.for.some.api", 150*time.Millisecond)
		client.PrecisionTiming("response.time.for.some.api.case1", 150*time.Millisecond)
	}

	_ = client.Close()
	_ = inSocket.Close()
}

func BenchmarkTagged(b *testing.B) {
	inSocket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP: net.IPv4(127, 0, 0, 1),
	})
	if err != nil {
		b.Error(err)
	}

	go func() {
		buf := make([]byte, 1500)
		for {
			_, err := inSocket.Read(buf)
			if err != nil {
				return
			}
		}

	}()

	client := NewClient(inSocket.LocalAddr().String(), MetricPrefix(prefixNoDot), MaxPacketSize(1432),
		FlushInterval(flushPeriod), SendLoopCount(2), DefaultTags(StringTag("host", "foo")),
		SendQueueCapacity(10), BufPoolCapacity(40))

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Incr(counterKey, 1, StringTag("route", "api.one"), IntTag("status", 200))
		client.Timing("another.value", 157, StringTag("service", "db"))
		client.PrecisionTiming("response.time.for.some.api", 150*time.Millisecond, IntTag("status", 404))
		client.PrecisionTiming("response.time.for.some.api.case1", 150*time.Millisecond, StringTag("service", "db"), IntTag("status", 200))
	}
	_ = client.Close()
	_ = inSocket.Close()
}

//
//func BenchmarkAlexcesaro(b *testing.B) {
//	s := newServer()
//	c, err := ac.New(
//		ac.Address(s.Addr()),
//		ac.Prefix(prefixNoDot),
//		ac.FlushPeriod(flushPeriod),
//	)
//	if err != nil {
//		b.Fatal(err)
//	}
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		c.Increment(counterKey)
//		c.Gauge(gaugeKey, gaugeValue)
//		c.Timing(timingKey, timingValue)
//	}
//	c.Close()
//	s.Close()
//}
//
//func BenchmarkGoStatsd(b *testing.B) {
//	s := newServer()
//	c := NewClient(s.Addr(), MetricPrefix(prefixNoDot), MaxPacketSize(1432),
//		FlushInterval(flushPeriod), SendLoopCount(2))
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		c.Incr(counterKey, 1)
//		c.Gauge(gaugeKey, gaugeValue)
//		c.Timing(timingKey, int64(timingValue))
//	}
//	_ = c.Close()
//	s.Close()
//}
//
//func BenchmarkCactus(b *testing.B) {
//	s := newServer()
//	c, err := cactus.NewBufferedClient(s.Addr(), prefix, flushPeriod, 1432)
//	if err != nil {
//		b.Fatal(err)
//	}
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		_ = c.Inc(counterKey, 1, 1)
//		_ = c.Gauge(gaugeKey, gaugeValue, 1)
//		_ = c.Timing(timingKey, int64(timingValue), 1)
//	}
//	_ = c.Close()
//	s.Close()
//}
//
//func BenchmarkG2s(b *testing.B) {
//	s := newServer()
//	c, err := g2s.Dial("udp", s.Addr())
//	if err != nil {
//		b.Fatal(err)
//	}
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		c.Counter(1, counterKey, 1)
//		c.Gauge(1, gaugeKey, strconv.Itoa(gaugeValue))
//		c.Timing(1, timingKey, timingValue)
//	}
//	s.Close()
//}
//
//func BenchmarkQuipo(b *testing.B) {
//	s := newServer()
//	c := quipo.NewStatsdBuffer(flushPeriod, quipo.NewStatsdClient(s.Addr(), prefix))
//	c.Logger = logger{}
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		_ = c.Incr(counterKey, 1)
//		_ = c.Gauge(gaugeKey, gaugeValue)
//		_ = c.Timing(timingKey, int64(timingValue))
//	}
//	_ = c.Close()
//	s.Close()
//}
//
//func BenchmarkUnix4ever(b *testing.B) {
//	s := newServer()
//	c := unix4ever.NewStatsdClient(s.Addr(), prefix, 1400, flushPeriod, 10*time.Second)
//
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		_ = c.Incr(counterKey, 1)
//		_ = c.Gauge(gaugeKey, gaugeValue)
//		_ = c.Timing(timingKey, int64(timingValue))
//	}
//	_ = c.Close()
//	s.Close()
//}
//
//type server struct {
//	conn   *net.UDPConn
//	closed chan struct{}
//}
//
//func newServer() *server {
//	addr, err := net.ResolveUDPAddr("udp", addr)
//	if err != nil {
//		panic(err)
//	}
//	conn, err := net.ListenUDP("udp", addr)
//	if err != nil {
//		panic(err)
//	}
//	s := &server{conn: conn, closed: make(chan struct{})}
//	go func() {
//		buf := make([]byte, 512)
//		for {
//			_, err := conn.Read(buf)
//			if err != nil {
//				s.closed <- struct{}{}
//				return
//			}
//		}
//	}()
//	return s
//}
//
//func (s *server) Addr() string {
//	return s.conn.LocalAddr().String()
//}
//
//func (s *server) Close() {
//	_ = s.conn.Close()
//	<-s.closed
//}
//
