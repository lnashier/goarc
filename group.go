package goarc

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Group composes named Services into a single Service, so a process can run
// several components — e.g. an HTTP service, a gRPC service, and a
// background worker — together under one call to Run or Up.
//
// Group itself implements Service, so a Group can contain another Group.
type Group struct {
	members []groupMember
}

type groupMember struct {
	name string
	svc  Service
}

// NewGroup creates an empty Group.
func NewGroup() *Group {
	return &Group{}
}

// Add registers a named Service with the group. The name identifies the
// service in errors returned by Start and Stop; it need not be unique.
//
// Add is not safe to call concurrently with Start or Stop, or after Start
// has been called.
func (g *Group) Add(name string, svc Service) *Group {
	g.members = append(g.members, groupMember{name: name, svc: svc})
	return g
}

// Start starts every member concurrently, each sharing a context derived
// from ctx, and blocks until all of them have returned.
//
// The moment any member's Start returns — successfully or not — Start
// cancels the shared derived context, so every other member observes that
// as its own signal to self-terminate (per the Service contract) and the
// whole group winds down together rather than leaving siblings running
// orphaned.
//
// Start returns a joined error naming every member whose Start returned a
// non-nil error.
func (g *Group) Start(ctx context.Context) error {
	if len(g.members) == 0 {
		<-ctx.Done()
		return nil
	}

	memberCtx, cancelMembers := context.WithCancel(ctx)
	defer cancelMembers()

	type result struct {
		name string
		err  error
	}
	done := make(chan result, len(g.members))
	for _, m := range g.members {
		m := m
		go func() {
			done <- result{m.name, m.svc.Start(memberCtx)}
		}()
	}

	var errs []error
	for range g.members {
		r := <-done
		if r.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.name, r.err))
		}
		// As soon as any member is done, ask the rest to wind down too.
		cancelMembers()
	}
	return errors.Join(errs...)
}

// Stop stops every member concurrently, each given the same ctx, and waits
// for all of them to finish.
//
// Stop is safe to call at any time: before Start, concurrently with a
// running Start, after Start has already returned, or more than once — it
// relies on each member's own Stop honoring that same contract.
//
// Stop returns a joined error naming every member whose Stop returned a
// non-nil error.
func (g *Group) Stop(ctx context.Context) error {
	if len(g.members) == 0 {
		return nil
	}

	errs := make([]error, len(g.members))
	var wg sync.WaitGroup
	for i, m := range g.members {
		wg.Add(1)
		go func(i int, m groupMember) {
			defer wg.Done()
			if err := m.svc.Stop(ctx); err != nil {
				errs[i] = fmt.Errorf("%s: %w", m.name, err)
			}
		}(i, m)
	}
	wg.Wait()
	return errors.Join(errs...)
}
