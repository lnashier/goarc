package buildinfo

import (
	shttp "github.com/lnashier/goarc/http"
	"net/http"
)

// App configures the build information related endpoint /buildinfo for the given service.
// To customize the endpoints, get a New Client and register the endpoints with custom names.
// See Reporter, Report and Key
func App(s *shttp.Service) error {
	s.Register("/buildinfo", http.MethodGet, New())
	return nil
}
