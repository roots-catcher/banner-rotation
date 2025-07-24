-- Создание таблиц
CREATE TABLE IF NOT EXISTS slots (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS banners (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

-- Таблица для связи баннеров и слотов
CREATE TABLE IF NOT EXISTS banner_slots (
    slot_id INT REFERENCES slots(id) ON DELETE CASCADE,
    banner_id INT REFERENCES banners(id) ON DELETE CASCADE,
    PRIMARY KEY (slot_id, banner_id)
);

-- Таблица статистики
CREATE TABLE IF NOT EXISTS statistics (
    slot_id INT NOT NULL,
    banner_id INT NOT NULL,
    group_id INT NOT NULL,
    shows BIGINT NOT NULL DEFAULT 0,
    clicks BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY (slot_id, banner_id, group_id),
    FOREIGN KEY (slot_id, banner_id) REFERENCES banner_slots(slot_id, banner_id) ON DELETE CASCADE
);