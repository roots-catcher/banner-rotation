package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPostgresStorage(t *testing.T) {
	connStr := "postgres://rotation_user:rotation_pass@postgres:5432/banner_rotation?sslmode=disable"
	store, err := New(connStr)
	require.NoError(t, err)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("error closing store: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = store.db.Exec(ctx, `
        TRUNCATE TABLE groups, banner_slots, statistics, banners, slots CASCADE;
        INSERT INTO groups (id, description) VALUES (1, 'Group 1');
        INSERT INTO banners (id, description) VALUES (1, 'Banner 1');
        INSERT INTO slots (id, description) VALUES (1, 'Slot 1');
    `)
	require.NoError(t, err)

	t.Run("AddBannerToSlot", func(t *testing.T) {
		err := store.AddBannerToSlot(ctx, 1, 1)
		require.NoError(t, err)
	})

	t.Run("RecordShow", func(t *testing.T) {
		err := store.RecordShow(ctx, 1, 1, 1)
		require.NoError(t, err)
	})

	t.Run("RecordClick", func(t *testing.T) {
		err := store.RecordClick(ctx, 1, 1, 1)
		require.NoError(t, err)
	})

	t.Run("GetBannerStats", func(t *testing.T) {
		stats, err := store.GetBannerStats(ctx, 1, 1)
		require.NoError(t, err)
		require.Len(t, stats, 1)
		require.Equal(t, 1, stats[0].Shows)
		require.Equal(t, 1, stats[0].Clicks)
	})

	t.Run("RemoveBannerFromSlot", func(t *testing.T) {
		err := store.RemoveBannerFromSlot(ctx, 1, 1)
		require.NoError(t, err)
	})
}

func TestGetBannersForSlot(t *testing.T) {
	connStr := "postgres://rotation_user:rotation_pass@postgres:5432/banner_rotation?sslmode=disable"
	store, err := New(connStr)
	require.NoError(t, err)
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("error closing store: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Подготовка данных с заполнением обязательных полей
	_, err = store.db.Exec(ctx, `
        TRUNCATE TABLE banner_slots, banners, slots CASCADE;
        INSERT INTO slots (id, description) VALUES (10, 'Slot 10');
        INSERT INTO banners (id, description) VALUES (100, 'Banner 100'), (101, 'Banner 101');
        INSERT INTO banner_slots (slot_id, banner_id) VALUES (10, 100), (10, 101);
    `)
	require.NoError(t, err)

	// Тестирование
	t.Run("existing slot", func(t *testing.T) {
		ids, err := store.GetBannersForSlot(ctx, 10)
		require.NoError(t, err)
		require.ElementsMatch(t, []int{100, 101}, ids)
	})

	t.Run("empty slot", func(t *testing.T) {
		ids, err := store.GetBannersForSlot(ctx, 999)
		require.NoError(t, err)
		require.Empty(t, ids)
	})

	t.Run("invalid slot", func(t *testing.T) {
		ids, err := store.GetBannersForSlot(ctx, -1)
		require.NoError(t, err)
		require.Empty(t, ids)
	})
}
