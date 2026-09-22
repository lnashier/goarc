package grpc

import "github.com/lnashier/goarc/v2"

// Component is an alias for goarc.Component: long-running work the service
// runs alongside itself and must stop when the service stops.
type Component = goarc.Component
