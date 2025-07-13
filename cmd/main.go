package main

import (
	"banner-rotation/internal/app"
	"fmt"
)

func main() {
	fmt.Println("Banner Rotation Service")

	// Создаем "бандита" для слота 1 и группы пользователей 1
	bandit := app.NewBandit()

	// Добавляем баннеры
	bandit.RecordShow(1)  // Показ баннера 1
	bandit.RecordShow(2)  // Показ баннера 2
	bandit.RecordClick(2) // Клик по баннеру 2

	// Выбираем баннер для показа
	bannerID := bandit.ChooseBanner()
	fmt.Printf("Выбран баннер: %d\n", bannerID)

	// Добавляем новый баннер
	bandit.RecordShow(3)

	// Снова выбираем
	bannerID = bandit.ChooseBanner()
	fmt.Printf("Выбран баннер: %d\n", bannerID)
}
