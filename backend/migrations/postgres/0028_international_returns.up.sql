-- 0028_international_returns.up.sql (PostgreSQL)
-- Module: Purchasing
-- Returns of products to international suppliers. The supplier is billed
-- a USD credit (accounts payable negative / contra-AP). Returns to local
-- suppliers reuse the regular purchase-return flow (0029 movements).

CREATE TABLE international_returns (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id       UUID           NOT NULL,
    supplier_id      UUID           NOT NULL,
    number           VARCHAR(30)    NOT NULL,
    return_date      DATE           NOT NULL,
    currency_code    VARCHAR(3)     NOT NULL DEFAULT 'USD',
    exchange_rate    NUMERIC(18,6)  NOT NULL DEFAULT 1,
    subtotal         NUMERIC(18,2)  NOT NULL DEFAULT 0,
    total            NUMERIC(18,2)  NOT NULL DEFAULT 0,
    reason           TEXT,
    status           VARCHAR(20)    NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'authorized', 'received', 'reconciled', 'cancelled')),
    authorized_at    TIMESTAMPTZ,
    authorized_by    UUID,
    received_at      TIMESTAMPTZ,
    reconciled_at    TIMESTAMPTZ,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by       UUID,
    updated_by       UUID,

    CONSTRAINT fk_international_returns_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_international_returns_supplier
        FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_international_returns_total
        CHECK (total >= 0)
);

CREATE UNIQUE INDEX uq_international_returns_number
    ON international_returns (company_id, number);

CREATE TABLE international_return_items (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    international_return_id UUID           NOT NULL,
    product_id              UUID           NOT NULL,
    quantity                NUMERIC(18,4)  NOT NULL DEFAULT 0
        CHECK (quantity > 0),
    unit_cost               NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (unit_cost >= 0),
    line_total              NUMERIC(18,2)  NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

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

CREATE TRIGGER trg_international_returns_set_updated_at
    BEFORE UPDATE ON international_returns
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
