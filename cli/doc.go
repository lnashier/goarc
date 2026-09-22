// Package cli provides a CLI goarc.Service backed by cobra: NewService and
// Register wire up subcommands, and the service is composable into a
// larger process via goarc.Up/Run like any other goarc.Service — with the
// caveat that a registered command must itself honor ctx being done to
// return promptly, since Start blocks on user code it does not own.
package cli
