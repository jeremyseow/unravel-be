package postgres

import (
	. "github.com/go-jet/jet/v2/postgres"
	"github.com/google/uuid"
)

// TODO: replace with tenant from auth context (Phase 6)
var defaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// uuidStr wraps a UUID value in an explicit CAST to avoid the "uuid = text"
// operator error. pq sends string params as text OID and PostgreSQL has no
// implicit uuid = text operator.
func uuidStr(id uuid.UUID) StringExpression {
	return StringExp(CAST(String(id.String())).AS("uuid"))
}
