-- 0024_fiscal_periods.up.sql (SQLite)
-- Module: Accounting
-- Fiscal periods gate journal posting (open / closing / closed).

CREATE TABLE fiscal_periods (
    id           TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id   TEXT           NOT NULL,
    name         VARCHAR(100)   NOT NULL,
    period_start TIMESTAMP      NOT NULL,
    period_end   TIMESTAMP      NOT NULL,
    status       VARCHAR(20)    NOT NULL DEFAULT 'open'
        CHECK (status IN ('open', 'closing', 'closed')),
    closed_at    TIMESTAMP,
    closed_by    TEXT,
    created_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at   TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by   TEXT,
    updated_by   TEXT,

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
