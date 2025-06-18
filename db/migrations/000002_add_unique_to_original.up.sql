ALTER TABLE shortener_urls 
    ADD CONSTRAINT original_unique UNIQUE(original);