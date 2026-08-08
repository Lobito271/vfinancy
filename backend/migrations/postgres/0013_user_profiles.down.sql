-- 0013_user_profiles.down.sql

DROP TRIGGER IF EXISTS trg_profiles_set_updated_at ON user_profiles;
DROP TABLE IF EXISTS user_profiles;
