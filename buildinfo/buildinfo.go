package buildinfo

import (
	"github.com/lnashier/go-app/env"
	"github.com/lnashier/go-app/zson"
	"net/http"
	"time"
)

var (
	// Version set by build script
	Version = "undefined"
	// Hash set by build script
	// Usually it is git commit hash
	Hash = "undefined"
)

type Key string

const (
	KeyAppName    Key = "appName"
	KeyHost           = "host"
	KeyStartTime      = "startTime"
	KeyUptime         = "uptime"
	KeyVersion        = "version"
	KeyHash           = "hash"
	keyComponents     = "components"
)

// Report is buildinfo report
type Report map[Key]any

// Reporter creates buildinfo report
type Reporter func() Report

// Component interface to be implemented by components
type Component interface {
	// Monitor is to get component name, component buildinfo report
	Monitor() (string, Report)
}

// Componentable allows use of anonymous functions which satisfies the Component interface
type Componentable func() (string, Report)

// Monitor ...
func (c Componentable) Monitor() (string, Report) {
	return c()
}

// New returns a buildinfo client
func New(r Reporter, compos ...Component) *Client {
	return &Client{
		r:         r,
		host:      env.Hostname(),
		startTime: time.Now(),
		compos:    compos,
	}
}

type Client struct {
	r         Reporter
	host      string
	startTime time.Time
	compos    []Component
}

func (c *Client) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write(zson.Marshal(c.Report()))
}

func (c *Client) Report() Report {
	report := Report{
		KeyHost:      c.host,
		KeyStartTime: c.startTime.Format("2006-01-02T15:04:05Z"),
		KeyUptime:    int64(time.Since(c.startTime).Seconds()),
	}

	// If app custom Reporter is provided then get report from that too
	if c.r != nil {
		if exReport := c.r(); exReport != nil {
			for k, v := range exReport {
				report[k] = v
			}
		}
	}

	// if components are provided then get report from them too
	if len(c.compos) > 0 {
		cReports := make(map[string]Report)
		for _, comp := range c.compos {
			cname, cReport := comp.Monitor()
			cReports[cname] = cReport
		}
		report[keyComponents] = cReports
	}

	return report
}
