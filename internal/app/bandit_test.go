package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBandit(t *testing.T) {
	b := NewBandit()
	assert.NotNil(t, b)
	assert.Equal(t, 0, b.TotalShows)
	assert.Empty(t, b.Banners)
}

func TestRecordShow(t *testing.T) {
	b := NewBandit()

	// Первый показ баннера 1
	b.RecordShow(1)
	assert.Equal(t, 1, b.TotalShows)
	assert.Equal(t, 1, b.Banners[1].Shows)
	assert.Equal(t, 0, b.Banners[1].Clicks)

	// Второй показ баннера 1
	b.RecordShow(1)
	assert.Equal(t, 2, b.TotalShows)
	assert.Equal(t, 2, b.Banners[1].Shows)

	// Показ нового баннера 2
	b.RecordShow(2)
	assert.Equal(t, 3, b.TotalShows)
	assert.Equal(t, 1, b.Banners[2].Shows)
}

func TestRecordClick(t *testing.T) {
	b := NewBandit()
	b.RecordShow(1) // Сначала показ

	b.RecordClick(1)
	assert.Equal(t, 1, b.Banners[1].Clicks)

	b.RecordClick(1)
	assert.Equal(t, 2, b.Banners[1].Clicks)
}

func TestChooseBanner_NewBanners(t *testing.T) {
	b := NewBandit()

	// Добавляем баннеры но без показов
	b.Banners[1] = &BannerStat{ID: 1}
	b.Banners[2] = &BannerStat{ID: 2}

	// Должны выбираться оба так как у них максимальный приоритет
	selected := make(map[int]bool)
	for i := 0; i < 100; i++ {
		id := b.ChooseBanner()
		selected[id] = true
	}

	assert.True(t, selected[1])
	assert.True(t, selected[2])
}

func TestChooseBanner_PrefersBetter(t *testing.T) {
	b := NewBandit()

	// Баннер 1: 100 показов, 10 кликов (CTR=0.1)
	b.Banners[1] = &BannerStat{ID: 1, Shows: 100, Clicks: 10}

	// Баннер 2: 100 показов, 30 кликов (CTR=0.3)
	b.Banners[2] = &BannerStat{ID: 2, Shows: 100, Clicks: 30}

	b.TotalShows = 200

	// Выбираем 1000 раз
	counts := make(map[int]int)
	for i := 0; i < 1000; i++ {
		id := b.ChooseBanner()
		counts[id]++
	}

	// Баннер 2 должен выбираться чаще
	assert.Greater(t, counts[2], counts[1])
}

func TestChooseBanner_NewBannerGetsChance(t *testing.T) {
	b := NewBandit()

	// Старый баннер с хорошим CTR
	b.Banners[1] = &BannerStat{ID: 1, Shows: 1000, Clicks: 300} // CTR=0.3
	b.TotalShows = 1000

	// Добавляем новый баннер
	b.Banners[2] = &BannerStat{ID: 2} // Ни разу не показывали

	// Выбираем 100 раз
	foundNew := false
	for i := 0; i < 100; i++ {
		if b.ChooseBanner() == 2 {
			foundNew = true
			break
		}
	}

	assert.True(t, foundNew, "New banner should be selected at least once")
}
