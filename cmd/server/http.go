// Package http
package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"ticket-purchase/internal/api"
	"ticket-purchase/internal/ticket"
	"time"

	"github.com/bnkamalesh/errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
}

type HTTP struct {
	lock                      *sync.Mutex
	server                    *http.Server
	shutdownInitiated         bool
	serverStartTime           time.Time
	liveHealthResponse        map[string]string
	shutdownInitiatedResponse []byte
	api                       *api.API
}

func (ht *HTTP) Start() error {
	ht.serverStartTime = time.Now()
	ht.AppendHealthResponse(
		"http",
		fmt.Sprintf("OK: %s", ht.serverStartTime.Format(time.RFC3339Nano)),
	)
	err := ht.server.ListenAndServe()
	if err != nil {
		return errors.Wrap(err, "http listen and serve ended")
	}
	return nil
}

func (ht *HTTP) ResetHealthResponse() {
	ht.lock.Lock()
	ht.liveHealthResponse = map[string]string{}
	ht.lock.Unlock()
}

func (ht *HTTP) AppendHealthResponse(key, value string) {
	ht.lock.Lock()
	ht.liveHealthResponse[key] = value
	ht.lock.Unlock()
}

func (ht *HTTP) healhtResponse() (message []byte, status int) {
	ht.lock.Lock()
	if ht.shutdownInitiated {
		message = ht.shutdownInitiatedResponse
		status = http.StatusServiceUnavailable
	} else {
		message, _ = json.Marshal(ht.liveHealthResponse)
		status = http.StatusOK
	}
	ht.lock.Unlock()

	return message, status
}

func (ht *HTTP) Shutdown(ctx context.Context) error {
	ht.lock.Lock()
	ht.shutdownInitiated = true
	ht.shutdownInitiatedResponse = []byte(fmt.Sprintf("server is shutting down | %s", time.Now().Format(time.RFC3339Nano)))
	ht.lock.Unlock()

	err := ht.server.Shutdown(ctx)
	if err != nil {
		return errors.Wrap(err, "failed shutting down http server")
	}
	return nil
}

func (ht *HTTP) Health(w http.ResponseWriter, req *http.Request) {
	msg, status := ht.healhtResponse()
	if status == http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(status)
	_, _ = w.Write(msg)
}

func (ht *HTTP) ticketOptions(w http.ResponseWriter, r *http.Request) {
	//request.Body
	var to ticket.TicketOption
	err := json.NewDecoder(r.Body).Decode(&to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	res, err := ht.api.CreateTicketOptions(r.Context(), &to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON response body
	w.Write(jsonData)
}

func (ht *HTTP) purchase(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var purchase ticket.Purchase
	err := json.NewDecoder(r.Body).Decode(&purchase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	purchase.TicketOptionID = id
	err = ht.api.PurchaseTicket(r.Context(), &purchase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (ht *HTTP) ticket(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	to, err := ht.api.GetTicketOptions(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonData, err := json.Marshal(to)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

func New(cfg *Config, api *api.API, debug bool) *HTTP {
	ht := &HTTP{
		lock: &sync.Mutex{},
		api:  api,
	}
	ht.ResetHealthResponse()
	// create a new router
	router := chi.NewRouter()
	if debug {
		router.Use(middleware.Logger)
	}

	router.Get("/-/health", ht.Health)
	// to create a ticket option
	router.Post("/ticket_options", ht.ticketOptions)
	// to purchase a ticket
	router.Post("/ticket_options/{id}/purchases", ht.purchase)
	// return remaining allocation for a ticket option
	router.Get("/ticket/{id}", ht.ticket)

	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "swagger-ui/index.html")
	})

	ht.server = &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
	}
	ht.liveHealthResponse = map[string]string{}

	return ht
}
