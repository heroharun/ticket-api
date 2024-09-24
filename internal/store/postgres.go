package store

import (
	"context"
	"database/sql"
	"fmt"
	"ticket-purchase/internal/ticket"
)

type Methods interface {
	UpsertTicketOptions(ctx context.Context, ticketOptions *ticket.TicketOption) (*ticket.TicketOption, error)
	PurchaseTicket(ctx context.Context, purchase *ticket.Purchase) error
	GetTicketOptions(ctx context.Context, id string) (*ticket.TicketOption, error)
}

type Store struct {
	cli *sql.DB
}

func (s Store) UpsertTicketOptions(ctx context.Context, ticketOptions *ticket.TicketOption) (*ticket.TicketOption, error) {
	//first check if the ticket option already exist
	var name, desc string
	var id int
	err := s.cli.QueryRowContext(ctx, "SELECT id, name, description FROM ticket_options WHERE name = $1, description = $2", ticketOptions.Name, ticketOptions.Desc).Scan(&id, &name, &desc)

	if id != 0 {
		//update allocation
		_, err = s.cli.ExecContext(ctx, "UPDATE ticket_options SET allocation = allocation + $1 WHERE id = $2", ticketOptions.Allocation, id)
		return &ticket.TicketOption{ID: id, Name: name, Desc: desc, Allocation: ticketOptions.Allocation}, nil
	}

	_, err = s.cli.ExecContext(ctx, "INSERT INTO ticket_options (name, description, allocation) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $1, description = $2, allocation = $3 RETURNING id",
		ticketOptions.Name, ticketOptions.Desc, ticketOptions.Allocation)
	if err != nil {
		return nil, err
	}

	var to ticket.TicketOption
	err = s.cli.QueryRowContext(ctx, "SELECT id, name, description, allocation FROM ticket_options WHERE name = $1", ticketOptions.Name).Scan(&to.ID, &to.Name, &to.Desc, &to.Allocation)
	if err != nil {
		return nil, err
	}
	return &to, nil
}

func (s Store) PurchaseTicket(ctx context.Context, purchase *ticket.Purchase) error {
	// İşlemi bir işlem (transaction) içinde gerçekleştir
	tx, err := s.cli.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	//first check if the ticket option is available and purchase quantity is less than or equal to the allocation
	var allocation int
	err = tx.QueryRowContext(ctx, "SELECT allocation FROM ticket_options WHERE id = $1", purchase.TicketOptionID).Scan(&allocation)
	if err != nil {
		return err
	}
	if allocation < purchase.Quantity {
		return fmt.Errorf("not enough tickets available")
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO purchases (quantity, user_id, ticket_option_id) VALUES ($1, $2, $3)", purchase.Quantity, purchase.UserID, purchase.TicketOptionID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "UPDATE ticket_options SET allocation = allocation - $1 WHERE id = $2", purchase.Quantity, purchase.TicketOptionID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s Store) GetTicketOptions(ctx context.Context, id string) (*ticket.TicketOption, error) {
	var to ticket.TicketOption
	err := s.cli.QueryRow("SELECT id, name, description, allocation FROM ticket_options WHERE id = $1", id).Scan(&to.ID, &to.Name, &to.Desc, &to.Allocation)
	if err != nil {
		return nil, err
	}
	return &to, nil
}

func New(cli *sql.DB) *Store {
	return &Store{
		cli: cli,
	}
}
