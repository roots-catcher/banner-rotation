package postgres

import (
	"banner-rotation/internal/storage"
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	db *pgxpool.Pool
	mu sync.RWMutex
}

func New(connString string) (*PostgresStorage, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse conn string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &PostgresStorage{db: pool}, nil
}

func (s *PostgresStorage) AddBannerToSlot(ctx context.Context, slotID, bannerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(ctx, `
		INSERT INTO banner_slots (slot_id, banner_id)
		VALUES ($1, $2)
		ON CONFLICT (slot_id, banner_id) DO NOTHING`,
		slotID, bannerID,
	)

	return err
}

func (s *PostgresStorage) RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(ctx, `
		DELETE FROM banner_slots 
		WHERE slot_id = $1 AND banner_id = $2`,
		slotID, bannerID,
	)

	return err
}

func (s *PostgresStorage) RecordShow(ctx context.Context, slotID, bannerID, groupID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(ctx, `
		INSERT INTO statistics (slot_id, banner_id, group_id, shows)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (slot_id, banner_id, group_id)
		DO UPDATE SET shows = statistics.shows + 1`,
		slotID, bannerID, groupID,
	)

	return err
}

func (s *PostgresStorage) RecordClick(ctx context.Context, slotID, bannerID, groupID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(ctx, `
		INSERT INTO statistics (slot_id, banner_id, group_id, clicks)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (slot_id, banner_id, group_id)
		DO UPDATE SET clicks = statistics.clicks + 1`,
		slotID, bannerID, groupID,
	)

	return err
}

func (s *PostgresStorage) GetBannerStats(ctx context.Context, slotID, groupID int) ([]storage.BannerStat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(ctx, `
		SELECT banner_id, shows, clicks
		FROM statistics
		WHERE slot_id = $1 AND group_id = $2`,
		slotID, groupID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []storage.BannerStat
	for rows.Next() {
		var stat storage.BannerStat
		if err := rows.Scan(&stat.BannerID, &stat.Shows, &stat.Clicks); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *PostgresStorage) Close() error {
	s.db.Close()
	return nil
}

func (s *PostgresStorage) GetBannersForSlot(ctx context.Context, slotID int) ([]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(ctx, `
        SELECT banner_id 
        FROM banner_slots
        WHERE slot_id = $1`,
		slotID,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to query banners for slot: %w", err)
	}
	defer rows.Close()

	var bannerIDs []int
	for rows.Next() {
		var bannerID int
		if err := rows.Scan(&bannerID); err != nil {
			return nil, fmt.Errorf("failed to scan banner ID: %w", err)
		}
		bannerIDs = append(bannerIDs, bannerID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return bannerIDs, nil
}
