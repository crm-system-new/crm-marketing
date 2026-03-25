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

type SegmentRepository struct {
	pool *pgxpool.Pool
}

func NewSegmentRepository(pool *pgxpool.Pool) *SegmentRepository {
	return &SegmentRepository{pool: pool}
}

func (r *SegmentRepository) GetByID(ctx context.Context, id string) (*domain.Segment, error) {
	query := `SELECT id, name, criteria, subscriber_count, version, created_at, updated_at
		FROM segments WHERE id = $1`

	s := &domain.Segment{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Criteria, &s.SubscriberCount,
		&s.Version, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ddd.ErrNotFound{Entity: "Segment", ID: id}
		}
		return nil, fmt.Errorf("query segment: %w", err)
	}
	return s, nil
}

func (r *SegmentRepository) Save(ctx context.Context, segment *domain.Segment) error {
	query := `INSERT INTO segments (id, name, criteria, subscriber_count, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		segment.ID, segment.Name, segment.Criteria, segment.SubscriberCount,
		segment.Version, segment.CreatedAt, segment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert segment: %w", err)
	}
	return nil
}

func (r *SegmentRepository) SaveInTx(ctx context.Context, tx pgx.Tx, segment *domain.Segment) error {
	query := `INSERT INTO segments (id, name, criteria, subscriber_count, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := tx.Exec(ctx, query,
		segment.ID, segment.Name, segment.Criteria, segment.SubscriberCount,
		segment.Version, segment.CreatedAt, segment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert segment in tx: %w", err)
	}
	return nil
}

func (r *SegmentRepository) Update(ctx context.Context, segment *domain.Segment) error {
	oldVersion := segment.Version
	segment.IncrementVersion()

	query := `UPDATE segments SET name=$1, criteria=$2, subscriber_count=$3,
		updated_at=$4, version=$5
		WHERE id=$6 AND version=$7`

	return sharedpg.ExecWithOptimisticLockPool(ctx, r.pool, query,
		segment.Name, segment.Criteria, segment.SubscriberCount,
		segment.UpdatedAt, segment.Version,
		segment.ID, oldVersion,
	)
}

func (r *SegmentRepository) UpdateInTx(ctx context.Context, tx pgx.Tx, segment *domain.Segment) error {
	oldVersion := segment.Version
	segment.IncrementVersion()

	query := `UPDATE segments SET name=$1, criteria=$2, subscriber_count=$3,
		updated_at=$4, version=$5
		WHERE id=$6 AND version=$7`

	return sharedpg.ExecWithOptimisticLock(ctx, tx, query,
		segment.Name, segment.Criteria, segment.SubscriberCount,
		segment.UpdatedAt, segment.Version,
		segment.ID, oldVersion,
	)
}

func (r *SegmentRepository) List(ctx context.Context, filter domain.SegmentFilter) ([]*domain.Segment, int, error) {
	countQuery := `SELECT COUNT(*) FROM segments WHERE 1=1`
	listQuery := `SELECT id, name, criteria, subscriber_count, version, created_at, updated_at
		FROM segments WHERE 1=1`
	args := []any{}
	argIdx := 1

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count segments: %w", err)
	}

	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list segments: %w", err)
	}
	defer rows.Close()

	var segments []*domain.Segment
	for rows.Next() {
		s := &domain.Segment{}
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Criteria, &s.SubscriberCount,
			&s.Version, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan segment: %w", err)
		}
		segments = append(segments, s)
	}

	return segments, total, nil
}
