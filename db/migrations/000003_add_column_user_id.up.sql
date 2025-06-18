ALTER TABLE shortener_urls
    ADD COLUMN user_id VARCHAR(8) DEFAULT '';

ALTER TABLE shortener_urls
    ALTER COLUMN user_id SET NOT NULL;