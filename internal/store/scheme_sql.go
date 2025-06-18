package store

const (
	InsertReq         = `INSERT INTO shortener_urls(original, short, user_id) VALUES ($1, $2, $3)`
	selectShortReq    = `SELECT short FROM shortener_urls WHERE original = $1`
	SelectOriginalReq = `SELECT original, is_deleted FROM shortener_urls WHERE short = $1`
	SelectUserURLsReq = `SELECT short,original FROM shortener_urls WHERE user_id = $1`
)
