// Package health provides liveness/readiness endpoints for an http.Service:
// Controller (a goarc.Component) flips readiness to not-found once stopped,
// while liveness stays up until the process exits; App wires both up as
// /alive and /ready.
package health
