package buildinfo

import (
	shttp "github.com/lnashier/goarc/http"
	"net/http"
)

// App configures the build information related endpoint /buildinfo for the given service.
// To customize the endpoints, get a New Handler and register the endpoints with custom names.
// To customize the report info, pass in a custom Reporter.
// See Reporter, Report and Key
func App(s *shttp.Service) error {
	s.Register("/buildinfo", http.MethodGet, New())
	return nil
}
