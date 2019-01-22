package metrics

import (
	"time"

	statsd "github.com/morfien101/go-statsd"
)

var (
	on        = false
	stdClient *statsd.Client
)

// Setup will start the statsd client and enable the functions for it.
func Setup(stastdHost string, tagsInput map[string]string) {
	stdClient = statsd.NewClient(
		stastdHost,
		statsd.MaxPacketSize(1400),
		statsd.ReconnectInterval(time.Second*60),
		statsd.MetricPrefix("chefwaiter."),
		statsd.TagStyle(statsd.TagFormatDatadog),
		statsd.DefaultTags(
			convertTags(tagsInput)...,
		),
	)
	on = true
}

// Shutdown will stop the metric client
func Shutdown() {
	stdClient.Close()
}

func convertTags(tagsInput map[string]string) []statsd.Tag {
	if len(tagsInput) > 0 {
		tags := make([]statsd.Tag, 0)
		for key, value := range tagsInput {
			tags = append(tags, statsd.StringTag(key, value))
		}
		return tags
	}
	return nil
}

// Incr increments a counter metric
//
// Often used to note a particular event, for example incoming web request.
func Incr(stat string, count int64, tagsInput map[string]string) {
	if on {
		stdClient.Incr(stat, count, convertTags(tagsInput)...)
	}
}

// Decr decrements a counter metric
//
// Often used to note a particular event
func Decr(stat string, count int64, tagsInput map[string]string) {
	if on {
		stdClient.Decr(stat, count, convertTags(tagsInput)...)
	}
}

// Timing tracks a duration event, the time delta must be given in milliseconds
func Timing(stat string, delta int64, tagsInput map[string]string) {
	if on {
		stdClient.Timing(stat, delta, convertTags(tagsInput)...)
	}
}

// Gauge sets or updates constant value for the interval
//
// Gauges are a constant data type. They are not subject to averaging,
// and they don’t change unless you change them. That is, once you set a gauge value,
// it will be a flat line on the graph until you change it again. If you specify
// delta to be true, that specifies that the gauge should be updated, not set. Due to the
// underlying protocol, you can't explicitly set a gauge to a negative number without
// first setting it to zero.
func Gauge(stat string, value int64, tagsInput map[string]string) {
	if on {
		stdClient.Gauge(stat, value, convertTags(tagsInput)...)
	}
}

// GaugeDelta sends a change for a gauge
func GaugeDelta(stat string, value int64, tagsInput map[string]string) {
	if on {
		stdClient.GaugeDelta(stat, value, convertTags(tagsInput)...)
	}
}

//  Functions Unlikely to be used.
//
//  FIncr(stat string, count float64, tags ...Tag)
//  FDecr(stat string, count float64, tags ...Tag)
//  PrecisionTiming(stat string, delta time.Duration, tags ...Tag)
//  igauge(stat string, sign []byte, value int64, tags ...Tag)
//  GaugeDelta(stat string, value int64, tags ...Tag)
//  fgauge(stat string, sign []byte, value float64, tags ...Tag)
//  FGauge(stat string, value float64, tags ...Tag)
//  FGaugeDelta(stat string, value float64, tags ...Tag)
//  SetAdd(stat string, value string, tags ...Tag)
