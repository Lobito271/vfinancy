-- 0018_exchange_rates.up.sql
-- Module 2: Administration — Exchange Rates
-- Daily exchange rates between currency pairs per company.
-- Rate represents: 1 unit of from_currency = rate units of to_currency.
-- Example: 1 USD = 3.75 PEN → from=USD, to=PEN, rate=3.750000

CREATE TABLE exchange_rates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id      UUID           NOT NULL,
    from_currency   VARCHAR(3)     NOT NULL,
    to_currency     VARCHAR(3)     NOT NULL,
    rate_date       DATE           NOT NULL,
    rate            NUMERIC(18,6)  NOT NULL
        CHECK (rate > 0),
    source          VARCHAR(50)    NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'central_bank', 'sunat', 'bloomberg', 'other')),
    created_at      TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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
