-- Добавляем поддержку валют в таблицы
ALTER TABLE assets ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';
ALTER TABLE transactions ADD COLUMN currency VARCHAR(3) NOT NULL DEFAULT 'USD';
ALTER TABLE users ADD COLUMN base_currency VARCHAR(3) NOT NULL DEFAULT 'USD';

-- Создаем таблицу курсов валют
CREATE TABLE exchange_rates (
    id SERIAL PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(20,10) NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(from_currency, to_currency, timestamp)
);

-- Создаем индексы для быстрого поиска курсов
CREATE INDEX idx_exchange_rates_from_to ON exchange_rates(from_currency, to_currency);
CREATE INDEX idx_exchange_rates_timestamp ON exchange_rates(timestamp);

-- Добавляем ограничения для валидации валют
ALTER TABLE assets ADD CONSTRAINT check_currency_format CHECK (currency ~ '^[A-Z]{3}$');
ALTER TABLE transactions ADD CONSTRAINT check_transaction_currency_format CHECK (currency ~ '^[A-Z]{3}$');
ALTER TABLE users ADD CONSTRAINT check_user_currency_format CHECK (base_currency ~ '^[A-Z]{3}$');
ALTER TABLE exchange_rates ADD CONSTRAINT check_from_currency_format CHECK (from_currency ~ '^[A-Z]{3}$');
ALTER TABLE exchange_rates ADD CONSTRAINT check_to_currency_format CHECK (to_currency ~ '^[A-Z]{3}$');

-- Добавляем комментарии к таблицам
COMMENT ON COLUMN assets.currency IS 'Валюта актива (ISO 4217 код)';
COMMENT ON COLUMN transactions.currency IS 'Валюта транзакции (ISO 4217 код)';
COMMENT ON COLUMN users.base_currency IS 'Базовая валюта пользователя для отображения (ISO 4217 код)';
COMMENT ON COLUMN exchange_rates.from_currency IS 'Исходная валюта';
COMMENT ON COLUMN exchange_rates.to_currency IS 'Целевая валюта';
COMMENT ON COLUMN exchange_rates.rate IS 'Курс обмена (сколько целевой валюты за 1 исходную)';
COMMENT ON COLUMN exchange_rates.timestamp IS 'Время действия курса'; 