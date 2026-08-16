-- 0026_journal_entries.up.sql (PostgreSQL)
-- Module: Accounting
-- Double-entry bookkeeping. Posted entries are immutable; corrections
-- require a reversing entry. journal_entry_lines reference chart_of_accounts.

CREATE TABLE journal_entries (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id        UUID           NOT NULL,
    fiscal_period_id  UUID           NOT NULL,
    number            VARCHAR(30)    NOT NULL,
    entry_date        DATE           NOT NULL,
    posting_date      DATE,
    description       TEXT,
    source            VARCHAR(20)    NOT NULL
        CHECK (source IN ('sale', 'purchase', 'payment', 'manual', 'adjustment', 'closing', 'opening')),
    source_id         UUID,
    status            VARCHAR(20)    NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'posted', 'reversed')),
    reverses_entry_id UUID,
    reversed_by_entry_id UUID,
    posted_at         TIMESTAMPTZ,
    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    created_by        UUID,
    posted_by         UUID,

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
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id      UUID           NOT NULL,
    line_number           INTEGER        NOT NULL DEFAULT 1
        CHECK (line_number >= 1),
    account_id            UUID           NOT NULL,
    description           TEXT,
    debit                 NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (debit >= 0),
    credit                NUMERIC(18,2)  NOT NULL DEFAULT 0
        CHECK (credit >= 0),
    currency_code         VARCHAR(3)     NOT NULL DEFAULT 'PEN',
    exchange_rate         NUMERIC(18,6)  NOT NULL DEFAULT 1,
    amount_in_txn_currency NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_journal_entry_lines_entry
        FOREIGN KEY (journal_entry_id) REFERENCES journal_entries(id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_journal_entry_lines_account
        FOREIGN KEY (account_id) REFERENCES chart_of_accounts(id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT ck_journal_entry_lines_one_side
        CHECK ((debit > 0) <> (credit > 0))
);

CREATE INDEX idx_journal_entry_lines_entry
    ON journal_entry_lines (journal_entry_id, line_number);

CREATE INDEX idx_journal_entry_lines_account
    ON journal_entry_lines (account_id);

CREATE TRIGGER trg_journal_entries_set_updated_at
    BEFORE UPDATE ON journal_entries
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
