-- 0014_application_settings.down.sql

DROP TRIGGER IF EXISTS trg_settings_set_updated_at ON application_settings;
DROP TABLE IF EXISTS application_settings;
