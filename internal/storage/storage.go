package storage

import (
	"context"
)

// Storage - интерфейс для работы с хранилищем
type Storage interface {
	// Добавляет баннер в ротацию слота
	AddBannerToSlot(ctx context.Context, slotID, bannerID int) error

	// Удаляет баннер из ротации слота
	RemoveBannerFromSlot(ctx context.Context, slotID, bannerID int) error

	// Регистрирует показ баннера
	RecordShow(ctx context.Context, slotID, bannerID, groupID int) error

	// Регистрирует клик по баннеру
	RecordClick(ctx context.Context, slotID, bannerID, groupID int) error

	// Возвращает статистику для баннеров в слоте и группе
	GetBannerStats(ctx context.Context, slotID, groupID int) ([]BannerStat, error)

	// Получить все баннеры в слоте
	GetBannersForSlot(ctx context.Context, slotID int) ([]int, error)

	Close() error
}

// BannerStat - статистика баннера
type BannerStat struct {
	BannerID int
	Shows    int
	Clicks   int
}
