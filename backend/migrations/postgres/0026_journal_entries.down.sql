-- 0026_journal_entries.down.sql (PostgreSQL)

DROP TRIGGER IF EXISTS trg_journal_entries_set_updated_at ON journal_entries;
DROP TABLE IF EXISTS journal_entry_lines;
DROP TABLE IF EXISTS journal_entries;
