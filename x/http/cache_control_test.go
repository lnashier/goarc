package http

import (
	"net/http"
	"testing"
	"time"
)

func TestCacheControl_Decompose(t *testing.T) {
	cases := []struct {
		name    string
		header  string
		public  bool
		private bool
		maxAge  float64
		noCache bool
		noStore bool
	}{
		{"public with max-age", "public, max-age=60", true, false, 60, false, false},
		{"private", "private", false, true, 0, false, false},
		{"no-cache clears max-age", "max-age=60, no-cache", false, false, 0, true, false},
		{"no-store clears max-age", "max-age=60, no-store", false, false, 0, false, true},
		{"empty header", "", false, false, 0, false, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var cc CacheControl
			h := http.Header{}
			h.Set("Cache-Control", c.header)
			cc.Decompose(h)

			if cc.Public() != c.public {
				t.Errorf("Public() = %v, want %v", cc.Public(), c.public)
			}
			if cc.Private() != c.private {
				t.Errorf("Private() = %v, want %v", cc.Private(), c.private)
			}
			if cc.MaxAge() != c.maxAge {
				t.Errorf("MaxAge() = %v, want %v", cc.MaxAge(), c.maxAge)
			}
			if cc.NoCache() != c.noCache {
				t.Errorf("NoCache() = %v, want %v", cc.NoCache(), c.noCache)
			}
			if cc.NoStore() != c.noStore {
				t.Errorf("NoStore() = %v, want %v", cc.NoStore(), c.noStore)
			}
		})
	}
}

func TestCacheControl_MayCache(t *testing.T) {
	var cc CacheControl
	h := http.Header{}
	h.Set("Cache-Control", "public, max-age=60")
	cc.Decompose(h)

	if !cc.MayCache() {
		t.Fatal("MayCache() = false, want true")
	}

	var noCache CacheControl
	h2 := http.Header{}
	h2.Set("Cache-Control", "public, max-age=60, no-cache")
	noCache.Decompose(h2)
	if noCache.MayCache() {
		t.Fatal("MayCache() = true, want false when no-cache is set")
	}
}

func TestCacheControl_TimeLeftAndExpired(t *testing.T) {
	var cc CacheControl
	h := http.Header{}
	h.Set("Cache-Control", "public, max-age=1")
	cc.Decompose(h)

	if cc.Expired() {
		t.Fatal("Expired() = true immediately after Decompose, want false")
	}
	if left := cc.TimeLeft(); left <= 0 || left > 1 {
		t.Fatalf("TimeLeft() = %v, want (0, 1]", left)
	}

	time.Sleep(1100 * time.Millisecond)
	if !cc.Expired() {
		t.Fatal("Expired() = false after max-age elapsed, want true")
	}
	if left := cc.TimeLeft(); left != 0 {
		t.Fatalf("TimeLeft() = %v after expiry, want 0", left)
	}
}

func TestCacheControl_NeverCacheableIsAlwaysExpired(t *testing.T) {
	var cc CacheControl
	cc.Decompose(http.Header{}) // no Cache-Control header at all: max-age 0

	if !cc.Expired() {
		t.Fatal("Expired() = false with no max-age, want true")
	}
	if cc.TimeLeft() != 0 {
		t.Fatalf("TimeLeft() = %v, want 0", cc.TimeLeft())
	}
}
