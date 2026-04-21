package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgc, err := tcpostgres.Run(ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "testcontainers: failed to start postgres container: %v\n", err)
		os.Exit(1)
	}

	defer pgc.Terminate(ctx)

	connStr, connErr := pgc.ConnectionString(ctx, "sslmode=disable")
	if connErr != nil {
		fmt.Fprintf(os.Stderr, "testcontainers: failed to get connection string: %v\n", connErr)
		os.Exit(1)
	}

	db, openErr := sql.Open("postgres", connStr)
	if openErr != nil {
		fmt.Fprintf(os.Stderr, "testcontainers: failed to open db: %v\n", openErr)
		os.Exit(1)
	}

	testDB = db
	code := m.Run()
	os.Exit(code)
}
