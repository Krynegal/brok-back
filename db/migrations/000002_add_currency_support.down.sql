-- Удаляем таблицу курсов валют
DROP TABLE IF EXISTS exchange_rates;

-- Удаляем колонки валют из таблиц
ALTER TABLE assets DROP COLUMN IF EXISTS currency;
ALTER TABLE transactions DROP COLUMN IF EXISTS currency;
ALTER TABLE users DROP COLUMN IF EXISTS base_currency; 