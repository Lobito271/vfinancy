-- 0026_journal_entries.up.sql (SQLite)
-- Module: Accounting
-- Double-entry bookkeeping. Posted entries are immutable; corrections
-- require a reversing entry. journal_entry_lines reference chart_of_accounts.

CREATE TABLE journal_entries (
    id                TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    company_id        TEXT           NOT NULL,
    fiscal_period_id  TEXT           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    entry_date        TIMESTAMP      NOT NULL,
    posting_date      TIMESTAMP,
    description       TEXT,
    source            VARCHAR(20)    NOT NULL
        CHECK (source IN ('sale', 'purchase', 'payment', 'manual', 'adjustment', 'closing', 'opening')),
    source_id         TEXT,
    status            VARCHAR(20)    NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'posted', 'reversed')),
    reverses_entry_id TEXT,
    reversed_by_entry_id TEXT,
    posted_at         TIMESTAMP,
    created_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    updated_at        TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),
    created_by        TEXT,
    posted_by         TEXT,

    CONSTRAINT fk_journal_entries_company
        FOREIGN KEY (company_id) REFERENCES companies(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entries_period
        FOREIGN KEY (fiscal_period_id) REFERENCES fiscal_periods(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT fk_journal_entries_reverses
        FOREIGN KEY (reverses_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE SET NULL
);

CREATE UNIQUE INDEX uq_journal_entries_number
    ON journal_entries (company_id, number);

CREATE INDEX idx_journal_entries_period
    ON journal_entries (company_id, fiscal_period_id, entry_date);

CREATE INDEX idx_journal_entries_source
    ON journal_entries (source, source_id);

CREATE TABLE journal_entry_lines (
    id                    TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    journal_entry_id      TEXT           NOT NULL,
    line_number           INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    account_id            TEXT           NOT NULL,
    description           TEXT,
    debit                 TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(debit AS REAL) >= 0),
    credit                TEXT           NOT NULL DEFAULT '0.00'
        CHECK (CAST(credit AS REAL) >= 0),
    currency_code         VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate         TEXT           NOT NULL DEFAULT '1.000000',
    amount_in_txn_currency TEXT          NOT NULL DEFAULT '0.00',
    created_at            TIMESTAMP      NOT NULL DEFAULT (CAST(unixepoch('subsec') * 1000 AS INTEGER)),

    CONSTRAINT fk_journal_entry_lines_entry
        FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entry_lines_account
        FOREIGN KEY (account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_journal_entry_lines_one_side
        CHECK ((CAST(debit AS REAL) > 0) <> (CAST(credit AS REAL) > 0))
);

CREATE INDEX idx_journal_entry_lines_entry
    ON journal_entry_lines (journal_entry_id, line_number);

CREATE INDEX idx_journal_entry_lines_account
    ON journal_entry_lines (account_id);
