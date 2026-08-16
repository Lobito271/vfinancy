-- 0024_fiscal_periods.up.sql (PostgreSQL)
-- Module: Accounting
-- Fiscal periods gate journal posting (open / closing / closed).

CREATE TABLE fiscal_periods (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id   UUID           NOT NULL,
    name         VARCHAR(100)   NOT NULL,
    period_start DATE           NOT NULL,
    period_end   DATE           NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closing', 'closed')),
    closed_at    TIMESTAMPTZ,
    closed_by    UUID,
    created_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by   UUID,
    updated_by   UUID,

    CONSTRAINT fk_fiscal_periods_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT ck_fiscal_periods_range
        CHECK (period_end >= period_start)
);

CREATE UNIQUE INDEX uq_fiscal_periods_name
    ON fiscal_periods (company_id, name);

CREATE UNIQUE INDEX uq_fiscal_periods_range
    ON fiscal_periods (company_id, period_start, period_end);

CREATE TRIGGER trg_fiscal_periods_set_updated_at
    BEFORE UPDATE ON fiscal_periods
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
