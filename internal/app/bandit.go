package app

import (
	"math"
	"sync"
)

// BannerStat хранит статистику для одного баннера
type BannerStat struct {
	ID     int // ID баннера
	Shows  int // количество показов
	Clicks int // количество кликов
}

// Bandit - основной объект для алгоритма
type Bandit struct {
	mu         sync.RWMutex        // защита от конкурентного доступа
	TotalShows int                 // общее количество показов
	Banners    map[int]*BannerStat // статистика по баннерам
}

// NewBandit создает новый объект Bandit
func NewBandit() *Bandit {
	return &Bandit{
		Banners: make(map[int]*BannerStat),
	}
}

// ChooseBanner выбирает баннер для показа с помощью алгоритма UCB1
func (b *Bandit) ChooseBanner() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Если баннеров нет
	if len(b.Banners) == 0 {
		return 0
	}

	bestID := 0
	bestValue := -1.0

	for id, banner := range b.Banners {
		value := b.calculateUCB(banner)
		if value > bestValue {
			bestValue = value
			bestID = id
		}
	}

	return bestID
}

// calculateUCB вычисляет значение UCB для баннера
func (b *Bandit) calculateUCB(banner *BannerStat) float64 {
	// Если баннер еще не показывали - максимальный приоритет
	if banner.Shows == 0 {
		return math.MaxFloat64
	}

	// Вычисляем CTR (кликабельность)
	ctr := float64(banner.Clicks) / float64(banner.Shows)

	// Вычисляем "бонус исследования"
	exploration := math.Sqrt(2 * math.Log(float64(b.TotalShows)) / float64(banner.Shows))

	return ctr + exploration
}

// RecordShow регистрирует показ баннера
func (b *Bandit) RecordShow(bannerID int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Если баннер новый - создаем запись
	if _, exists := b.Banners[bannerID]; !exists {
		b.Banners[bannerID] = &BannerStat{ID: bannerID}
	}

	b.Banners[bannerID].Shows++
	b.TotalShows++
}

// RecordClick регистрирует клик по баннеру
func (b *Bandit) RecordClick(bannerID int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if banner, exists := b.Banners[bannerID]; exists {
		banner.Clicks++
	}
}
