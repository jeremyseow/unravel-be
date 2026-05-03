package postgres

import (
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

// uuidStr wraps a UUID value in an explicit CAST to avoid the "uuid = text"
// operator error: go-jet maps PostgreSQL uuid columns to ColumnString, so pq
// sends the parameter as text OID, but PostgreSQL has no uuid = text operator.
func uuidStr(id uuid.UUID) StringExpression {
	return StringExp(CAST(String(id.String())).AS("uuid"))
}
