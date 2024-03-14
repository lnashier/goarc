package buildinfo

import (
	"github.com/lnashier/goarc/x/env"
	"github.com/lnashier/goarc/x/zson"
	"net/http"
	"time"
)

var (
	// Version set by build script
	// It can be set via ldflags
	Version = "undefined"
	// Hash set by build script; usually it is git commit hash
	// It can be set via ldflags
	Hash = "undefined"
)

// Key defines Report keys
type Key string

const (
	KeyHost      Key = "host"
	KeyStartTime     = "startTime"
	KeyUptime        = "uptime"
	KeyVersion       = "version"
	KeyHash          = "hash"
)

// Report is build-info report
type Report map[Key]any

// Reporter provides build-info report
type Reporter func() Report

// Client provides Key build information about current service:
//
//	KeyHost
//	KeyVersion
//	KeyHash
//	KeyStartTime
//	KeyUptime
//
// Custom Reporter can override all the Report Key since it runs after inbuilt reporter,
// Custom Reporter can override the default reporting keys, it runs after the built-in reporter.
type Client struct {
	r         Reporter
	host      string
	startTime time.Time
}

func New(r ...Reporter) *Client {
	var r1 Reporter
	if len(r) > 0 {
		r1 = r[0]
	}
	return &Client{
		r:         r1,
		host:      env.Hostname(),
		startTime: time.Now(),
	}
}

// ServeHTTP makes Client http.Handler
func (c *Client) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(zson.Marshal(c.report()))
}

func (c *Client) report() Report {
	report := Report{
		KeyHost:      c.host,
		KeyVersion:   Version,
		KeyHash:      Hash,
		KeyStartTime: c.startTime.Format("2006-01-02T15:04:05Z"),
		KeyUptime:    int64(time.Since(c.startTime).Seconds()),
	}

	// If custom Reporter is provided then get info from that too
	if c.r != nil {
		if exReport := c.r(); exReport != nil {
			for k, v := range exReport {
				report[k] = v
			}
		}
	}

	return report
}
