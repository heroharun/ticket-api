package api

import (
	"context"
	"ticket-purchase/internal/store"
	"ticket-purchase/internal/ticket"
)

type API struct {
	store store.Methods
}

func New(store store.Methods) *API {
	return &API{
		store: store,
	}
}

func (api *API) CreateTicketOptions(ctx context.Context, ticketOptions *ticket.TicketOption) (*ticket.TicketOption, error) {
	return api.store.UpsertTicketOptions(ctx, ticketOptions)
}

func (api *API) PurchaseTicket(ctx context.Context, purchase *ticket.Purchase) error {
	return api.store.PurchaseTicket(ctx, purchase)
}

func (api *API) GetTicketOptions(ctx context.Context, id string) (*ticket.TicketOption, error) {
	return api.store.GetTicketOptions(ctx, id)
}
