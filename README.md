# Сервис ротации баннеров

Это сервис для показа баннеров с использованием алгоритма многорукого бандита.

## Запуск

```sh
make run
```

## Остановка

```sh
make down
```

## Примеры запросов к API

### Добавить баннер в слот
```
POST /api/v1/banner_slot
{
  "slot_id": 1,
  "banner_id": 100
}
```

### Удалить баннер из слота
```
DELETE /api/v1/banner_slot
{
  "slot_id": 1,
  "banner_id": 100
}
```

### Засчитать клик
```
POST /api/v1/register_click
{
  "slot_id": 1,
  "banner_id": 100,
  "group_id": 1
}
```

### Выбрать баннер для показа
```
POST /api/v1/choose_banner
{
  "slot_id": 1,
  "group_id": 1
}
Ответ: { "banner_id": 100 }
```

[![CI Status](https://github.com/roots-catcher/banner-rotation/actions/workflows/ci.yml/badge.svg)](https://github.com/roots-catcher/banner-rotation/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/roots-catcher/banner-rotation)](https://goreportcard.com/report/github.com/roots-catcher/banner-rotation)