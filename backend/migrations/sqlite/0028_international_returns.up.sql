-- 0028_international_returns.up.sql (SQLite)
-- Module: Purchasing
-- Returns of products to international suppliers. The supplier is billed
-- a USD credit (accounts payable negative / contra-AP). Returns to local
-- suppliers reuse the regular purchase-return flow (0029 movements).

CREATE TABLE international_returns (
    id               TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id       TEXT           NOT NULL,
    supplier_id      TEXT           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    return_date      TIMESTAMP      NOT NULL,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'USD',
    exchange_rate    TEXT           NOT NULL DEFAULT '1.000000',
    subtotal         TEXT           NOT NULL DEFAULT '0.00',
    total            TEXT           NOT NULL DEFAULT '0.00',
    reason           TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'authorized', 'received', 'reconciled', 'cancelled')),
    authorized_at    TIMESTAMP,
    authorized_by    TEXT,
    received_at      TIMESTAMP,
    reconciled_at    TIMESTAMP,
    created_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at       TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by       TEXT,
    updated_by       TEXT,

    CONSTRAINT fk_international_returns_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_international_returns_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_international_returns_total
        CHECK (CAST(total AS REAL) >= 0)
);

CREATE UNIQUE INDEX uq_international_returns_number
    ON international_returns (company_id, number);

CREATE TABLE international_return_items (
    id                      TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    international_return_id TEXT           NOT NULL,
    product_id              TEXT           NOT NULL,
    quantity                TEXT           NOT NULL DEFAULT '0.0000'
        CHECK (CAST(quantity AS REAL) > 0),
    unit_cost               TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(unit_cost AS REAL) >= 0),
    line_total               TEXT          NOT NULL DEFAULT '0.00',
    created_at              TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_intl_return_items_return
        FOREIGN KEY (international_return_id) REFERENCES international_returns(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_intl_return_items_product
        FOREIGN KEY (product_id) REFERENCES products(id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE INDEX idx_intl_return_items_return
    ON international_return_items (international_return_id);

CREATE INDEX idx_intl_return_items_product
    ON international_return_items (product_id);
