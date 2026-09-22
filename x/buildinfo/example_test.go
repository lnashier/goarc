package buildinfo_test

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"github.com/lnashier/goarc/v2/x/buildinfo"
)

func Example() {
	h := buildinfo.New(func() buildinfo.Report {
		return buildinfo.Report{buildinfo.KeyVersion: "v1.2.3"}
	})

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/buildinfo", nil))

	var report map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &report)
	fmt.Println(report["version"])
	// Output:
	// v1.2.3
}
