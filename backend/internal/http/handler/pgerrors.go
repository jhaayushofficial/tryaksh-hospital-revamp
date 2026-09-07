package handler

// PostgreSQL SQLSTATE codes the handlers translate into client-facing errors.
// See https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	pgUniqueViolation = "23505"
)
