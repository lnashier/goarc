package service

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/lnashier/go-app/pkg/buildinfo"
	"github.com/lnashier/go-app/pkg/config"
	"github.com/lnashier/go-app/pkg/log"
	"github.com/urfave/negroni"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	name             string
	cfg              *config.Config
	httpServer       *http.Server
	router           *mux.Router
	preempt          *negroni.Negroni
	healthController *healthController
	components       []Component
	exitCh           chan struct{}
}

func NewServer(cfg *config.Config) *Server {
	log.Info("server@New enter")
	defer log.Info("server@New exit")

	preempt := negroni.New()

	return &Server{
		name: cfg.GetString("name"),
		cfg:  cfg,
		httpServer: &http.Server{
			Addr:    ":" + cfg.GetString("server.port"),
			Handler: preempt,
		},
		preempt:          preempt,
		router:           mux.NewRouter(),
		healthController: newHealthController(),
		exitCh:           make(chan struct{}),
	}
}

// Start starts the server instance
// Routes should be registered before calling start
func (s *Server) Start() {
	log.Info("Server#Start enter")
	defer log.Info("Server#Start exit")

	s.setupHealthEndpoints()
	s.setupBuildInfoEndpoints()
	if s.cfg.GetBool("log.network") {
		s.preempt.UseHandler(&LogHandler{logger: log.DefaultLogger, handler: s.router})
	} else {
		s.preempt.UseHandler(s.router)
	}

	log.Info("Server#Start %s serving %s", s.name, s.httpServer.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("Server#Start HTTP server closed: %v", err.Error())
			return
		}
		log.Error("Server#Start error encountered starting HTTP server: %v", err.Error())
	}
}

func (s *Server) Stop() {
	log.Info("Server#Stop enter")
	defer log.Info("Server#Stop exit")

	close(s.exitCh)

	log.Info("Server#Stop server status set to down")
	s.healthController.setStatusDown()

	// Stop any long-running services within the app
	if len(s.components) > 0 {
		for _, comp := range s.components {
			if err := comp.Stop(); err != nil {
				log.Info("Server#Stop failed to stop a running component: %v", err)
			}
		}
	}

	log.Info("Server#Stop going to sleep for gracetime")
	time.Sleep(time.Second * time.Duration(s.cfg.GetInt("server.shutdown.gracetime")))
	log.Info("Server#Stop woke up after gracetime")

	err := s.httpServer.Shutdown(context.Background())
	if err != nil {
		log.Error("Server#Stop error encountered shutting down HTTP server: %v", err.Error())
		return
	}
	log.Info("Server#Stop server shutdown completed")
}

func (s *Server) setupHealthEndpoints() {
	log.Info("Server#setupHealthEndpoints enter")
	defer log.Info("Server#setupHealthEndpoints exit")

	s.router.Methods(http.MethodGet).Path("/alive").HandlerFunc(s.healthController.LiveHandler)
	s.router.Methods(http.MethodGet).Path("/health").HandlerFunc(s.healthController.ReadyHandler)
}

func (s *Server) setupBuildInfoEndpoints() {
	log.Info("Server#setupBuildInfoEndpoints enter")
	defer log.Info("Server#setupBuildInfoEndpoints exit")

	s.router.Handle("/buildinfo", buildinfo.New(
		func() buildinfo.Report {
			return buildinfo.Report{
				buildinfo.KeyAppName: s.name,
				buildinfo.KeyVersion: buildinfo.Version,
				buildinfo.KeyHash:    buildinfo.Hash,
			}
		},
	))
}

// Register registers a RouteHandler for a given path and http.Method
// preHandlers can be optionally supplied, they will run before routeHandler in the order supplied
func (s *Server) Register(path, method string, routeHandler http.Handler, preHandlers ...negroni.Handler) {
	chain := negroni.New(preHandlers...)
	chain.UseHandler(&reject{
		h:    routeHandler,
		done: s.exitCh,
	})
	path = "/" + strings.TrimLeft(path, "/")
	if err := s.router.Handle(path, chain).Methods(method).GetError(); err != nil {
		log.Panic("Server#Register error %v building route for %v", err.Error(), path)
	}
	log.Info("Server#Register route [%v]%v registered", method, path)
}

// Component registers a Component that will run within the app that requires stopping when server shuts down
func (s *Server) Component(comp Component) {
	s.components = append(s.components, comp)
}
