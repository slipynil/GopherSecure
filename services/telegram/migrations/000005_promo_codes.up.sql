CREATE TABLE promo_codes (
    id          SERIAL PRIMARY KEY,
    code        TEXT UNIQUE NOT NULL,
    bonus_days  INT NOT NULL,                 -- бонусные дни подписки
    max_uses    INT DEFAULT 0,                -- 0 = безлимит
    used_count  INT DEFAULT 0,
    is_active   BOOLEAN DEFAULT TRUE,         -- мягкое удаление
    expires_at  TIMESTAMP NOT NULL,           -- когда истекает сам код
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE promo_activations (
    id              SERIAL PRIMARY KEY,
    promo_code_id   INT NOT NULL REFERENCES promo_codes(id),
    chat_id         BIGINT NOT NULL REFERENCES client(chat_id) ON DELETE CASCADE,
    activated_at    TIMESTAMP DEFAULT NOW(),
    UNIQUE (promo_code_id, chat_id)           -- один пользователь = одна активация кода
);

-- Индексы для производительности
CREATE INDEX idx_promo_activations_chat_id ON promo_activations(chat_id);
CREATE INDEX idx_promo_codes_expires_at ON promo_codes(expires_at);
