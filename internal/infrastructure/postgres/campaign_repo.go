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

type CampaignRepository struct {
	pool *pgxpool.Pool
}

func NewCampaignRepository(pool *pgxpool.Pool) *CampaignRepository {
	return &CampaignRepository{pool: pool}
}

func (r *CampaignRepository) GetByID(ctx context.Context, id string) (*domain.Campaign, error) {
	query := `SELECT id, name, description, status, channel, segment_id, scheduled_at, sent_count, open_rate, click_rate, version, created_at, updated_at
		FROM campaigns WHERE id = $1`

	c := &domain.Campaign{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Description, &c.Status, &c.Channel,
		&c.SegmentID, &c.ScheduledAt, &c.SentCount, &c.OpenRate, &c.ClickRate,
		&c.Version, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ddd.ErrNotFound{Entity: "Campaign", ID: id}
		}
		return nil, fmt.Errorf("query campaign: %w", err)
	}
	return c, nil
}

func (r *CampaignRepository) Save(ctx context.Context, campaign *domain.Campaign) error {
	query := `INSERT INTO campaigns (id, name, description, status, channel, segment_id, scheduled_at, sent_count, open_rate, click_rate, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.pool.Exec(ctx, query,
		campaign.ID, campaign.Name, campaign.Description, campaign.Status, campaign.Channel,
		campaign.SegmentID, campaign.ScheduledAt, campaign.SentCount, campaign.OpenRate, campaign.ClickRate,
		campaign.Version, campaign.CreatedAt, campaign.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert campaign: %w", err)
	}
	return nil
}

func (r *CampaignRepository) SaveInTx(ctx context.Context, tx pgx.Tx, campaign *domain.Campaign) error {
	query := `INSERT INTO campaigns (id, name, description, status, channel, segment_id, scheduled_at, sent_count, open_rate, click_rate, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := tx.Exec(ctx, query,
		campaign.ID, campaign.Name, campaign.Description, campaign.Status, campaign.Channel,
		campaign.SegmentID, campaign.ScheduledAt, campaign.SentCount, campaign.OpenRate, campaign.ClickRate,
		campaign.Version, campaign.CreatedAt, campaign.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert campaign in tx: %w", err)
	}
	return nil
}

func (r *CampaignRepository) Update(ctx context.Context, campaign *domain.Campaign) error {
	oldVersion := campaign.Version
	campaign.IncrementVersion()

	query := `UPDATE campaigns SET name=$1, description=$2, status=$3, channel=$4,
		segment_id=$5, scheduled_at=$6, sent_count=$7, open_rate=$8, click_rate=$9,
		updated_at=$10, version=$11
		WHERE id=$12 AND version=$13`

	return sharedpg.ExecWithOptimisticLockPool(ctx, r.pool, query,
		campaign.Name, campaign.Description, campaign.Status, campaign.Channel,
		campaign.SegmentID, campaign.ScheduledAt, campaign.SentCount, campaign.OpenRate, campaign.ClickRate,
		campaign.UpdatedAt, campaign.Version,
		campaign.ID, oldVersion,
	)
}

func (r *CampaignRepository) UpdateInTx(ctx context.Context, tx pgx.Tx, campaign *domain.Campaign) error {
	oldVersion := campaign.Version
	campaign.IncrementVersion()

	query := `UPDATE campaigns SET name=$1, description=$2, status=$3, channel=$4,
		segment_id=$5, scheduled_at=$6, sent_count=$7, open_rate=$8, click_rate=$9,
		updated_at=$10, version=$11
		WHERE id=$12 AND version=$13`

	return sharedpg.ExecWithOptimisticLock(ctx, tx, query,
		campaign.Name, campaign.Description, campaign.Status, campaign.Channel,
		campaign.SegmentID, campaign.ScheduledAt, campaign.SentCount, campaign.OpenRate, campaign.ClickRate,
		campaign.UpdatedAt, campaign.Version,
		campaign.ID, oldVersion,
	)
}

func (r *CampaignRepository) List(ctx context.Context, filter domain.CampaignFilter) ([]*domain.Campaign, int, error) {
	countQuery := `SELECT COUNT(*) FROM campaigns WHERE 1=1`
	listQuery := `SELECT id, name, description, status, channel, segment_id, scheduled_at, sent_count, open_rate, click_rate, version, created_at, updated_at
		FROM campaigns WHERE 1=1`
	args := []any{}
	argIdx := 1

	if filter.Status != nil {
		clause := fmt.Sprintf(" AND status = $%d", argIdx)
		countQuery += clause
		listQuery += clause
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Channel != nil {
		clause := fmt.Sprintf(" AND channel = $%d", argIdx)
		countQuery += clause
		listQuery += clause
		args = append(args, *filter.Channel)
		argIdx++
	}
	if filter.SegmentID != nil {
		clause := fmt.Sprintf(" AND segment_id = $%d", argIdx)
		countQuery += clause
		listQuery += clause
		args = append(args, *filter.SegmentID)
		argIdx++
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count campaigns: %w", err)
	}

	listQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*domain.Campaign
	for rows.Next() {
		c := &domain.Campaign{}
		if err := rows.Scan(
			&c.ID, &c.Name, &c.Description, &c.Status, &c.Channel,
			&c.SegmentID, &c.ScheduledAt, &c.SentCount, &c.OpenRate, &c.ClickRate,
			&c.Version, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan campaign: %w", err)
		}
		campaigns = append(campaigns, c)
	}

	return campaigns, total, nil
}
