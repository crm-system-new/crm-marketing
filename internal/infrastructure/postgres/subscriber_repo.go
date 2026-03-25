package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/crm-system-new/crm-marketing/internal/domain"
	"github.com/crm-system-new/crm-shared/pkg/ddd"
	sharedpg "github.com/crm-system-new/crm-shared/pkg/postgres"
)

type SubscriberRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriberRepository(pool *pgxpool.Pool) *SubscriberRepository {
	return &SubscriberRepository{pool: pool}
}

func (r *SubscriberRepository) GetByID(ctx context.Context, id string) (*domain.Subscriber, error) {
	query := `SELECT id, email, first_name, last_name, status, preferences, version, created_at, updated_at
		FROM subscribers WHERE id = $1`

	s := &domain.Subscriber{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Email, &s.FirstName, &s.LastName, &s.Status,
		&s.Preferences, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ddd.ErrNotFound{Entity: "Subscriber", ID: id}
		}
		return nil, fmt.Errorf("query subscriber: %w", err)
	}
	return s, nil
}

func (r *SubscriberRepository) GetByEmail(ctx context.Context, email string) (*domain.Subscriber, error) {
	query := `SELECT id, email, first_name, last_name, status, preferences, version, created_at, updated_at
		FROM subscribers WHERE email = $1`

	s := &domain.Subscriber{}
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&s.ID, &s.Email, &s.FirstName, &s.LastName, &s.Status,
		&s.Preferences, &s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ddd.ErrNotFound{Entity: "Subscriber", ID: email}
		}
		return nil, fmt.Errorf("query subscriber by email: %w", err)
	}
	return s, nil
}

func (r *SubscriberRepository) Save(ctx context.Context, subscriber *domain.Subscriber) error {
	query := `INSERT INTO subscribers (id, email, first_name, last_name, status, preferences, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.pool.Exec(ctx, query,
		subscriber.ID, subscriber.Email, subscriber.FirstName, subscriber.LastName,
		subscriber.Status, subscriber.Preferences,
		subscriber.Version, subscriber.CreatedAt, subscriber.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert subscriber: %w", err)
	}
	return nil
}

func (r *SubscriberRepository) Update(ctx context.Context, subscriber *domain.Subscriber) error {
	oldVersion := subscriber.Version
	subscriber.IncrementVersion()

	query := `UPDATE subscribers SET email=$1, first_name=$2, last_name=$3, status=$4,
		preferences=$5, updated_at=$6, version=$7
		WHERE id=$8 AND version=$9`

	return sharedpg.ExecWithOptimisticLockPool(ctx, r.pool, query,
		subscriber.Email, subscriber.FirstName, subscriber.LastName, subscriber.Status,
		subscriber.Preferences, subscriber.UpdatedAt, subscriber.Version,
		subscriber.ID, oldVersion,
	)
}

func (r *SubscriberRepository) List(ctx context.Context, filter domain.SubscriberFilter) ([]*domain.Subscriber, int, error) {
	countQuery := `SELECT COUNT(*) FROM subscribers WHERE 1=1`
	listQuery := `SELECT id, email, first_name, last_name, status, preferences, version, created_at, updated_at
		FROM subscribers WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Status != nil {
		clause := fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += clause
		listQuery += clause
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Email != nil {
		clause := fmt.Sprintf(" AND email = $%d", argIdx)
		countQuery += clause
		listQuery += clause
		args = append(args, *filter.Email)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count subscribers: %w", err)
	}

	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list subscribers: %w", err)
	}
	defer rows.Close()

	var subscribers []*domain.Subscriber
	for rows.Next() {
		s := &domain.Subscriber{}
		if err := rows.Scan(
			&s.ID, &s.Email, &s.FirstName, &s.LastName, &s.Status,
			&s.Preferences, &s.Version, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan subscriber: %w", err)
		}
		subscribers = append(subscribers, s)
	}

	return subscribers, total, nil
}
