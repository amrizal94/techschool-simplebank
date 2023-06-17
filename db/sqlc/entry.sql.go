// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.18.0
// source: entry.sql

package db

import (
	"context"
)

const createEntry = `-- name: CreateEntry :one
INSERT INTO entries (
  accounts_id,
  amount 
) VALUES (
  $1, $2
)
RETURNING id, accounts_id, amount, created_at
`

type CreateEntryParams struct {
	AccountsID int64 `json:"accounts_id"`
	Amount     int64 `json:"amount"`
}

func (q *Queries) CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error) {
	row := q.db.QueryRowContext(ctx, createEntry, arg.AccountsID, arg.Amount)
	var i Entry
	err := row.Scan(
		&i.ID,
		&i.AccountsID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getEntry = `-- name: GetEntry :one
SELECT id, accounts_id, amount, created_at FROM entries
WHERE id = $1 LIMIT 1
`

func (q *Queries) GetEntry(ctx context.Context, id int64) (Entry, error) {
	row := q.db.QueryRowContext(ctx, getEntry, id)
	var i Entry
	err := row.Scan(
		&i.ID,
		&i.AccountsID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const listEntries = `-- name: ListEntries :many
SELECT id, accounts_id, amount, created_at FROM entries
WHERE accounts_id = $1
ORDER BY id
LIMIT $2
OFFSET $3
`

type ListEntriesParams struct {
	AccountsID int64 `json:"accounts_id"`
	Limit      int32 `json:"limit"`
	Offset     int32 `json:"offset"`
}

func (q *Queries) ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error) {
	rows, err := q.db.QueryContext(ctx, listEntries, arg.AccountsID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Entry{}
	for rows.Next() {
		var i Entry
		if err := rows.Scan(
			&i.ID,
			&i.AccountsID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
