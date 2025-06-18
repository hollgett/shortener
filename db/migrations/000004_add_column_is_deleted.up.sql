ALTER TABLE shortener_urls 
    ADD COLUMN is_deleted BOOLEAN DEFAULT false;

ALTER TABLE shortener_urls
    ALTER COLUMN is_deleted SET NOT NULL;