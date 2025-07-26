-- Удаление старых таблиц
DROP TABLE IF EXISTS statistics;
DROP TABLE IF EXISTS banner_slots;
DROP TABLE IF EXISTS banners;
DROP TABLE IF EXISTS slots;
DROP TABLE IF EXISTS groups;

-- Создание таблиц с обязательными полями
CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE banners (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE slots (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL
);

CREATE TABLE banner_slots (
    slot_id INT NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    banner_id INT NOT NULL REFERENCES banners(id) ON DELETE CASCADE,
    PRIMARY KEY (slot_id, banner_id)
);

CREATE TABLE statistics (
    slot_id INT NOT NULL,
    banner_id INT NOT NULL,
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    shows INT DEFAULT 0,
    clicks INT DEFAULT 0,
    PRIMARY KEY (slot_id, banner_id, group_id),
    FOREIGN KEY (slot_id, banner_id) REFERENCES banner_slots(slot_id, banner_id) ON DELETE CASCADE
);