package env

import (
	"os"
	"testing"
)

func TestEnvironment_Predicates(t *testing.T) {
	cases := []struct {
		env              Environment
		local, dev, prod bool
	}{
		{Local, true, false, false},
		{Dev, false, true, false},
		{Prod, false, false, true},
		{Environment("staging"), false, false, false},
		{Environment(""), false, false, false},
	}
	for _, c := range cases {
		t.Run(string(c.env), func(t *testing.T) {
			if got := c.env.IsLocal(); got != c.local {
				t.Errorf("IsLocal() = %v, want %v", got, c.local)
			}
			if got := c.env.IsDev(); got != c.dev {
				t.Errorf("IsDev() = %v, want %v", got, c.dev)
			}
			if got := c.env.IsProd(); got != c.prod {
				t.Errorf("IsProd() = %v, want %v", got, c.prod)
			}
			if got := c.env.String(); got != string(c.env) {
				t.Errorf("String() = %q, want %q", got, string(c.env))
			}
		})
	}
}

func TestParseEnv_ReadsENVVariable(t *testing.T) {
	t.Setenv("ENV", "prod")
	if got := parseEnv(); got != Prod {
		t.Fatalf("parseEnv() = %q, want %q", got, Prod)
	}
}

func TestParseEnv_DefaultsToLocalWhenUnset(t *testing.T) {
	orig, had := os.LookupEnv("ENV")
	os.Unsetenv("ENV")
	t.Cleanup(func() {
		if had {
			os.Setenv("ENV", orig)
		}
	})

	if got := parseEnv(); got != Local {
		t.Fatalf("parseEnv() = %q, want %q", got, Local)
	}
}

func TestGet_IsMemoized(t *testing.T) {
	first := Get()
	if first == "" {
		t.Fatal("Get() = \"\", want a non-empty Environment")
	}
	if second := Get(); second != first {
		t.Fatalf("Get() returned %q then %q, want the same cached value both times", first, second)
	}
}

func TestHostname_NonEmpty(t *testing.T) {
	if Hostname() == "" {
		t.Fatal("Hostname() = \"\", want a non-empty string")
	}
}
