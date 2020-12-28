package repository

const (
	createEmailQuery = `INSERT INTO emails ("to", "from", subject, body, content_type) VALUES ($1, $2, $3, $4, $5) RETURNING email_id`

	findEmailByIdQuery = `SELECT email_id, "to", "from", subject, body, content_type, created_at FROM emails WHERE email_id = $1`

	totalCountQuery = `SELECT COUNT (email_id) as totalCount FROM emails WHERE "to" ILIKE '%' || $1 || '%'`

	findEmailByReceiverQuery = `SELECT email_id, "to", "from", subject, body, content_type, created_at 
	FROM emails WHERE "to" ILIKE '%' || $1 || '%' ORDER BY created_at OFFSET $2 LIMIT $3`
)
