DROP INDEX IF EXISTS idx_peer_preshared_key;
ALTER TABLE peer DROP COLUMN preshared_key;
