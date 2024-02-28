package http

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CacheControl is decomposed form of standard Cache response directives
// It answers questions like:
//   - May application cache the response?
//   - Should cache be revalidated with the origin server on every request?
//   - How long response can be cached?
//   - etc ...
//
// Each function answers a specific question.
type CacheControl struct {
	dtime   time.Time
	public  bool
	private bool
	maxAge  float64
	noCache bool
	noStore bool
}

// Decompose helps decompose Cache-Control header
// It looks at following standard Cache response directives:
//   - Cache-Control: public
//   - Cache-Control: private
//   - Cache-Control: max-age=<seconds>
//   - Cache-Control: no-cache
//     If no-cache is used, the ETag value is not checked
//     Effectively treating it as no-store
//   - Cache-Control: no-store
//
// It does not look at following standard Cache response directives:
//   - Cache-Control: must-revalidate
//   - Cache-Control: no-transform
//   - Cache-Control: proxy-revalidate
//   - Cache-Control: s-maxage=<seconds>
//
// It does not look at following Caching headers:
//   - Expires
//   - Age
func (cc *CacheControl) Decompose(h http.Header) {
	cc.dtime = time.Now()
	hparts := strings.Split(h.Get("Cache-Control"), ",")
	for _, hpart := range hparts {
		dparts := strings.Split(strings.TrimSpace(hpart), "=")
		if len(dparts) > 0 {
			switch dparts[0] {
			case "public":
				cc.public = true
			case "private":
				cc.private = true
			case "max-age":
				if len(dparts) == 2 {
					cc.maxAge, _ = strconv.ParseFloat(dparts[1], 64)
				}
			case "no-cache":
				cc.noCache = true
			case "no-store":
				cc.noStore = true
			}
		}
	}
	// sanity check
	if cc.noCache || cc.noStore {
		cc.maxAge = 0
	}
}

// Public provides value of Cache-Control: public directive
func (cc *CacheControl) Public() bool {
	return cc.public
}

// Private provides value of Cache-Control: private directive
func (cc *CacheControl) Private() bool {
	return cc.private
}

// MaxAge provides value of Cache-Control: max-age directive
func (cc *CacheControl) MaxAge() float64 {
	return cc.maxAge
}

// NoCache provides value of Cache-Control: no-cache directive
func (cc *CacheControl) NoCache() bool {
	return cc.noCache
}

// NoStore provides value of Cache-Control: no-store directive
func (cc *CacheControl) NoStore() bool {
	return cc.noStore
}

// MayCache tells if application may cache the response
func (cc *CacheControl) MayCache() bool {
	return !cc.noCache && !cc.noStore && cc.maxAge > 0
}

// TimeLeft tells time until expiration
//
// If MayCache() is false then cache is always considered expired
// otherwise remaining time is calculated since this Cache-Control
// was decomposed
//
// For example:
// return value 5 indicates cache will expire in 5 seconds
// return value 0 indicates cache has expired
func (cc *CacheControl) TimeLeft() float64 {
	if cc.Expired() {
		return 0
	}
	return math.Max(0, cc.maxAge-time.Since(cc.dtime).Seconds())
}

// Expired tells if cache is still valid
// If MayCache() is false then cache is always considered expired
// otherwise remaining time is calculated since this Cache-Control
// was decomposed
func (cc *CacheControl) Expired() bool {
	return !cc.MayCache() || time.Since(cc.dtime).Seconds() > cc.maxAge
}
