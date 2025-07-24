package app

import (
	"banner-rotation/internal/pkg/events"
	"banner-rotation/internal/storage"
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage - полная реализация mock-хранилища
type MockStorage struct {
	mu          sync.RWMutex
	stats       map[string]storage.BannerStat // ключ: "slotID_groupID_bannerID"
	bannerSlots map[int]map[int]bool          // slotID -> bannerID -> true
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		stats:       make(map[string]storage.BannerStat),
		bannerSlots: make(map[int]map[int]bool),
	}
}

func (m *MockStorage) key(slotID, groupID, bannerID int) string {
	return fmt.Sprintf("%d_%d_%d", slotID, groupID, bannerID)
}

func (m *MockStorage) AddBannerToSlot(ctx context.Context, slotID, bannerID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.bannerSlots[slotID]; !ok {
		m.bannerSlots[slotID] = make(map[int]bool)
	}
	m.bannerSlots[slotID][bannerID] = true
	return nil
}

func (m *MockStorage) RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if banners, ok := m.bannerSlots[slotID]; ok {
		delete(banners, bannerID)
	}
	return nil
}

func (m *MockStorage) RecordShow(ctx context.Context, slotID, bannerID, groupID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(slotID, groupID, bannerID)
	stat := m.stats[key]
	stat.BannerID = bannerID
	stat.Shows++
	m.stats[key] = stat
	return nil
}

func (m *MockStorage) RecordClick(ctx context.Context, slotID, bannerID, groupID int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.key(slotID, groupID, bannerID)
	stat := m.stats[key]
	stat.BannerID = bannerID
	stat.Clicks++
	m.stats[key] = stat
	return nil
}

func (m *MockStorage) GetBannerStats(ctx context.Context, slotID, groupID int) ([]storage.BannerStat, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stats []storage.BannerStat
	for key, stat := range m.stats {
		var sID, gID, bID int
		_, err := fmt.Sscanf(key, "%d_%d_%d", &sID, &gID, &bID)
		if err != nil {
			continue
		}
		if sID == slotID && gID == groupID {
			// Проверяем что баннер все еще в слоте
			if banners, ok := m.bannerSlots[slotID]; ok {
				if _, exists := banners[bID]; exists {
					stats = append(stats, stat)
				}
			}
		}
	}
	return stats, nil
}

func (m *MockStorage) GetBannersForSlot(ctx context.Context, slotID int) ([]int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if banners, ok := m.bannerSlots[slotID]; ok {
		ids := make([]int, 0, len(banners))
		for id := range banners {
			ids = append(ids, id)
		}
		return ids, nil
	}
	return nil, nil
}

func (m *MockStorage) Close() error {
	return nil
}

type MockProducer struct{}

func (m *MockProducer) SendEvent(ctx context.Context, eventType events.EventType, slotID, bannerID, groupID int) error {
	return nil
}

func (m *MockProducer) Close() error {
	return nil
}

func TestBandit_New(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	require.NotNil(t, bandit)
}

func TestBandit_ChooseBanner_NoBanners(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	_, err := bandit.ChooseBanner(ctx, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no banners in rotation")
}

func TestBandit_ChooseBanner_NewBanners(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 2))

	selected := make(map[int]bool)
	for i := 0; i < 100; i++ {
		bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
		require.NoError(t, err)
		selected[bannerID] = true
	}

	assert.True(t, selected[1])
	assert.True(t, selected[2])
}

func TestBandit_ChooseBanner_PrefersBetter(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 2))

	// Имитируем статистику: баннер 2 лучше
	for i := 0; i < 100; i++ {
		// Показы
		require.NoError(t, store.RecordShow(ctx, 1, 1, 1))
		require.NoError(t, store.RecordShow(ctx, 1, 2, 1))
		// Клики: баннер 1 - 10, баннер 2 - 30
		if i < 10 {
			require.NoError(t, store.RecordClick(ctx, 1, 1, 1))
		}
		if i < 30 {
			require.NoError(t, store.RecordClick(ctx, 1, 2, 1))
		}
	}

	// Пересоздаем bandit, чтобы он перечитал статистику
	bandit = NewBandit(store, producer)

	counts := make(map[int]int)
	for i := 0; i < 1000; i++ {
		bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
		require.NoError(t, err)
		counts[bannerID]++
	}

	assert.Greater(t, counts[2], counts[1])
}

func TestBandit_ChooseBanner_NewBannerGetsChance(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	// Имитируем 100 показов и кликов для баннера 1
	for i := 0; i < 100; i++ {
		require.NoError(t, store.RecordShow(ctx, 1, 1, 1))
		if i < 30 { // 30 кликов
			require.NoError(t, store.RecordClick(ctx, 1, 1, 1))
		}
	}

	// Добавляем новый баннер 2
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 2))

	// Пересоздаем bandit, чтобы он перечитал список баннеров и статистику
	bandit = NewBandit(store, producer)

	found := false
	for i := 0; i < 100; i++ {
		bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
		require.NoError(t, err)
		if bannerID == 2 {
			found = true
			break
		}
	}

	assert.True(t, found, "new banner should be selected at least once")
}

func TestBandit_AddNewBanner(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{}
	bandit := NewBandit(store, producer)
	ctx := context.Background()

	// Добавляем баннер 1
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	// Добавляем баннер 2 после 100 итераций
	for i := 0; i < 100; i++ {
		_, _ = bandit.ChooseBanner(ctx, 1, 1)
	}
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 2))

	// Проверяем, что новый баннер появляется в выборе
	found := false
	for i := 0; i < 50; i++ {
		bID, _ := bandit.ChooseBanner(ctx, 1, 1)
		if bID == 2 {
			found = true
			break
		}
	}
	assert.True(t, found, "New banner should be selected")
}

func TestBandit_RecordClick(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, bannerID)

	require.NoError(t, bandit.RecordClick(ctx, 1, 1, 1))

	stats, err := store.GetBannerStats(ctx, 1, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, 1, stats[0].Clicks)
}

func TestBandit_CacheUpdate(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, bannerID)

	key := bandit.getCacheKey(1, 1)
	bandit.mu.RLock()
	cache, exists := bandit.cache[key]
	bandit.mu.RUnlock()

	require.True(t, exists)
	require.Contains(t, cache.banners, 1)
	assert.Equal(t, 1, cache.banners[1].Shows)
	assert.Equal(t, 0, cache.banners[1].Clicks)

	// Клик
	require.NoError(t, bandit.RecordClick(ctx, 1, 1, 1))

	bandit.mu.RLock()
	cache = bandit.cache[key]
	bandit.mu.RUnlock()

	assert.Equal(t, 1, cache.banners[1].Clicks)
}

func TestBandit_CacheClearOnRemove(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	_, err := bandit.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)

	key := bandit.getCacheKey(1, 1)
	bandit.mu.RLock()
	_, exists := bandit.cache[key]
	bandit.mu.RUnlock()
	require.True(t, exists)

	require.NoError(t, bandit.RemoveBannerFromSlot(ctx, 1, 1))

	bandit.mu.RLock()
	_, exists = bandit.cache[key]
	bandit.mu.RUnlock()
	assert.False(t, exists, "cache should be cleared for slot")
}

func TestBandit_ConcurrentAccess(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
			require.NoError(t, err)
			assert.Equal(t, 1, bannerID)
		}()
	}
	wg.Wait()

	stats, err := store.GetBannerStats(ctx, 1, 1)
	require.NoError(t, err)
	require.Len(t, stats, 1)
	assert.Equal(t, 100, stats[0].Shows)
}

func TestBandit_AddRemoveBanners(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	// Добавляем баннер
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))

	// Выбираем баннер
	bannerID, err := bandit.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, bannerID)

	// Удаляем баннер
	require.NoError(t, bandit.RemoveBannerFromSlot(ctx, 1, 1))

	// Пытаемся выбрать снова
	_, err = bandit.ChooseBanner(ctx, 1, 1)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no banners in rotation")
}

func TestBandit_StatsPersistence(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit1 := NewBandit(store, producer)
	ctx := context.Background()

	// Добавляем баннер
	require.NoError(t, bandit1.AddBannerToSlot(ctx, 1, 1))

	// Регистрируем действия
	_, err := bandit1.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	require.NoError(t, bandit1.RecordClick(ctx, 1, 1, 1))

	// Создаем новый bandit с тем же хранилищем
	bandit2 := NewBandit(store, producer)

	// Выбираем баннер
	bannerID, err := bandit2.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, 1, bannerID)

	// Проверяем кеш
	key := bandit2.getCacheKey(1, 1)
	bandit2.mu.RLock()
	cache, exists := bandit2.cache[key]
	bandit2.mu.RUnlock()

	require.True(t, exists)
	assert.Equal(t, 2, cache.banners[1].Shows) // 1 от bandit1 + 1 от bandit2
	assert.Equal(t, 1, cache.banners[1].Clicks)
}

func TestBandit_MultipleSlotsGroups(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	// Настраиваем слот 1
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 1))
	require.NoError(t, bandit.AddBannerToSlot(ctx, 1, 2))

	// Настраиваем слот 2
	require.NoError(t, bandit.AddBannerToSlot(ctx, 2, 3))
	require.NoError(t, bandit.AddBannerToSlot(ctx, 2, 4))

	// Группа 1
	banner1, err := bandit.ChooseBanner(ctx, 1, 1)
	require.NoError(t, err)
	assert.True(t, banner1 == 1 || banner1 == 2)

	// Группа 2
	banner2, err := bandit.ChooseBanner(ctx, 1, 2)
	require.NoError(t, err)
	assert.True(t, banner2 == 1 || banner2 == 2)

	// Слот 2, группа 1
	banner3, err := bandit.ChooseBanner(ctx, 2, 1)
	require.NoError(t, err)
	assert.True(t, banner3 == 3 || banner3 == 4)

	// Проверяем кеш
	require.Len(t, bandit.cache, 3) // 3 комбинации: (1,1), (1,2), (2,1)
}

func TestBandit_Performance(t *testing.T) {
	store := NewMockStorage()
	producer := &MockProducer{} // Используем mock, реализующий интерфейс

	bandit := NewBandit(store, producer)
	ctx := context.Background()

	// Добавляем 100 баннеров
	for i := 1; i <= 100; i++ {
		require.NoError(t, bandit.AddBannerToSlot(ctx, 1, i))
	}

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, err := bandit.ChooseBanner(ctx, 1, 1)
		require.NoError(t, err)
	}
	duration := time.Since(start)

	t.Logf("Processed 1000 requests in %v", duration)
	assert.Less(t, duration, time.Second, "should handle 1000 requests quickly")
}
