package app

import (
	"banner-rotation/internal/kafka"
	"banner-rotation/internal/pkg/events"
	"banner-rotation/internal/storage"
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
)

// BanditInterface определяет контракт для работы с ротацией баннеров
type BanditInterface interface {
	AddBannerToSlot(ctx context.Context, slotID, bannerID int) error
	RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error
	ChooseBanner(ctx context.Context, slotID, groupID int) (int, error)
	RecordClick(ctx context.Context, slotID, bannerID, groupID int) error
}

var _ BanditInterface = (*Bandit)(nil)

// Bandit - основной объект для управления ротацией баннеров
type Bandit struct {
	mu       sync.RWMutex
	store    storage.Storage
	cache    map[string]banditCache
	producer kafka.ProducerInterface
}

// banditCache - кешированная статистика для комбинации слот+группа
type banditCache struct {
	totalShows int
	banners    map[int]*BannerStat
}

// BannerStat - статистика для одного баннера
type BannerStat struct {
	Shows  int
	Clicks int
}

// NewBandit создает новый экземпляр Bandit
func NewBandit(store storage.Storage, producer kafka.ProducerInterface) *Bandit {
	return &Bandit{
		store:    store,
		cache:    make(map[string]banditCache),
		producer: producer,
	}
}

// getCacheKey генерирует ключ кеша для комбинации слот+группа
func (b *Bandit) getCacheKey(slotID, groupID int) string {
	return fmt.Sprintf("%d_%d", slotID, groupID)
}

var (
	ErrNoBanners = errors.New("no banners in rotation for slot")
)

// loadStats загружает статистику из хранилища или кеша
func (b *Bandit) loadStats(ctx context.Context, slotID, groupID int) (banditCache, error) {
	bannerIDs, err := b.store.GetBannersForSlot(ctx, slotID)
	if err != nil {
		return banditCache{}, fmt.Errorf("failed to get banners: %w", err)
	}

	cache := banditCache{
		banners:    make(map[int]*BannerStat),
		totalShows: 0,
	}

	// Инициализация для всех баннеров
	for _, id := range bannerIDs {
		cache.banners[id] = &BannerStat{}
	}

	// 3. Если в слоте нет баннеров - возвращаем ошибку
	if len(bannerIDs) == 0 {
		return banditCache{}, ErrNoBanners
	}

	// 4. Инициализация временного кеша
	tempCache := banditCache{
		banners:    make(map[int]*BannerStat, len(bannerIDs)),
		totalShows: 0,
	}

	// 5. Создаем записи для всех баннеров (даже без статистики)
	for _, id := range bannerIDs {
		tempCache.banners[id] = &BannerStat{}
	}

	// 6. Загрузка статистики из хранилища
	stats, err := b.store.GetBannerStats(ctx, slotID, groupID)
	if err != nil {
		return banditCache{}, fmt.Errorf("failed to get stats for slot %d group %d: %w",
			slotID, groupID, err)
	}

	// 7. Обновление кеша на основе загруженной статистики
	for _, stat := range stats {
		if banner, exists := tempCache.banners[stat.BannerID]; exists {
			banner.Shows = stat.Shows
			banner.Clicks = stat.Clicks
			tempCache.totalShows += stat.Shows
		}
	}

	// 8. Сохранение в основное хранилище кеша
	b.mu.Lock()
	key := b.getCacheKey(slotID, groupID)
	b.cache[key] = tempCache
	b.mu.Unlock()

	return tempCache, nil
}

func (b *Bandit) sendEvent(eventType events.EventType, slotID, bannerID, groupID int) {
	if b.producer == nil {
		return
	}

	go func() {
		_ = b.producer.SendEvent(context.Background(), eventType, slotID, bannerID, groupID)
	}()
}

// ChooseBanner выбирает баннер для показа в указанном слоте для группы
func (b *Bandit) ChooseBanner(ctx context.Context, slotID, groupID int) (int, error) {
	// Загружаем статистику
	cache, err := b.loadStats(ctx, slotID, groupID)
	if err != nil {
		return 0, err
	}

	// Если нет баннеров в ротации
	if len(cache.banners) == 0 {
		return 0, fmt.Errorf("no banners in rotation for slot %d", slotID)
	}

	// Выбираем баннер с помощью алгоритма
	bannerID := b.chooseBanner(cache)

	// Регистрируем показ в хранилище
	if err := b.store.RecordShow(ctx, slotID, bannerID, groupID); err != nil {
		return 0, fmt.Errorf("failed to record show: %w", err)
	}

	// Обновляем кеш
	b.updateCache(slotID, groupID, bannerID, true, false)

	b.sendEvent(events.EventShow, slotID, bannerID, groupID)

	return bannerID, nil
}

// chooseBanner реализует алгоритм для выбора баннера
func (b *Bandit) chooseBanner(cache banditCache) int {
	bestID := 0
	bestValue := -1.0

	for bannerID, stat := range cache.banners {
		value := b.calculateUCB(stat, cache.totalShows)
		if value > bestValue {
			bestValue = value
			bestID = bannerID
		}
	}

	return bestID
}

// calculateUCB вычисляет значение UCB для баннера
func (b *Bandit) calculateUCB(stat *BannerStat, totalShows int) float64 {
	// Если баннер еще не показывали - максимальный приоритет
	if stat.Shows == 0 {
		return math.MaxFloat64
	}

	// Вычисляем CTR (кликабельность)
	ctr := float64(stat.Clicks) / float64(stat.Shows)

	// Вычисляем "бонус исследования"
	exploration := math.Sqrt(2 * math.Log(float64(totalShows)) / float64(stat.Shows))

	return ctr + exploration
}

// RecordClick регистрирует клик по баннеру
func (b *Bandit) RecordClick(ctx context.Context, slotID, bannerID, groupID int) error {
	// Регистрируем клик в хранилище
	if err := b.store.RecordClick(ctx, slotID, bannerID, groupID); err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	// Обновляем кеш
	b.updateCache(slotID, groupID, bannerID, false, true)

	b.sendEvent(events.EventClick, slotID, bannerID, groupID)

	return nil
}

// updateCache обновляет кеш после события (показ или клик)
func (b *Bandit) updateCache(slotID, groupID, bannerID int, isShow, isClick bool) {
	key := b.getCacheKey(slotID, groupID)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Получаем кеш (если нет - создаем)
	cache, ok := b.cache[key]
	if !ok {
		cache = banditCache{
			banners: make(map[int]*BannerStat),
		}
	}

	// Получаем статистику баннера (если нет - создаем)
	stat, ok := cache.banners[bannerID]
	if !ok {
		stat = &BannerStat{}
		cache.banners[bannerID] = stat
	}

	// Обновляем статистику
	if isShow {
		stat.Shows++
		cache.totalShows++
	}
	if isClick {
		stat.Clicks++
	}

	// Сохраняем обновленный кеш
	b.cache[key] = cache
}

// AddBannerToSlot добавляет баннер в ротацию слота
func (b *Bandit) AddBannerToSlot(ctx context.Context, slotID, bannerID int) error {
	if err := b.store.AddBannerToSlot(ctx, slotID, bannerID); err != nil {
		return fmt.Errorf("failed to add banner to slot: %w", err)
	}

	// При добавлении баннера нужно обновить кеш во всех группах
	b.clearCacheForSlot(slotID)
	return nil
}

// RemoveBannerFromSlot удаляет баннер из ротации слота
func (b *Bandit) RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error {
	if err := b.store.RemoveBannerFromSlot(ctx, slotID, bannerID); err != nil {
		return fmt.Errorf("failed to remove banner from slot: %w", err)
	}

	b.clearCacheForSlot(slotID)
	return nil
}

// clearCacheForSlot очищает кеш для всех групп в указанном слоте
func (b *Bandit) clearCacheForSlot(slotID int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Удаляем все кеши, которые относятся к этому слоту
	for key := range b.cache {
		var sID int
		_, err := fmt.Sscanf(key, "%d_", &sID)
		if err == nil && sID == slotID {
			delete(b.cache, key)
		}
	}
}
