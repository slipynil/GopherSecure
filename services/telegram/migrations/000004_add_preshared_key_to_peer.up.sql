ALTER TABLE peer ADD COLUMN preshared_key TEXT;
CREATE INDEX idx_peer_preshared_key ON peer(preshared_key) WHERE preshared_key IS NOT NULL;
