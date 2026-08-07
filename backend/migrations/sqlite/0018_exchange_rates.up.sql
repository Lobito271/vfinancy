-- 0018_exchange_rates.up.sql (SQLite)
-- Module 2: Administration — Exchange Rates
-- Daily exchange rates between currency pairs per company.
-- Rate represents: 1 unit of from_currency = rate units of to_currency.
-- Example: 1 USD = 3.75 PEN → from=USD, to=PEN, rate=3.750000
--
-- SQLite notes: NUMERIC(18,6) -> TEXT holding the exact decimal string,
-- rate_date DATE -> TEXT holding an ISO date string ('YYYY-MM-DD').

CREATE TABLE exchange_rates (
    id              TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id      TEXT           NOT NULL,
    from_currency   VARCHAR(3)     NOT NULL,
    to_currency     VARCHAR(3)     NOT NULL,
    rate_date       DATE           NOT NULL,
    rate            TEXT           NOT NULL
        CHECK (CAST(rate AS REAL) > 0),
    source          VARCHAR(50)    NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'central_bank', 'sunat', 'bloomberg', 'other')),
    created_at      TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_exchange_rates_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,

    CONSTRAINT fk_exchange_rates_from
        FOREIGN KEY (from_currency) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT fk_exchange_rates_to
        FOREIGN KEY (to_currency) REFERENCES currencies(code)
        ON UPDATE CASCADE ON DELETE RESTRICT,

    CONSTRAINT ck_exchange_rates_different_currencies
        CHECK (from_currency <> to_currency)
);

CREATE UNIQUE INDEX uq_exchange_rates_pair_date
    ON exchange_rates (company_id, from_currency, to_currency, rate_date);

CREATE INDEX idx_exchange_rates_lookup
    ON exchange_rates (company_id, from_currency, to_currency, rate_date DESC);
