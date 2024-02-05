package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	testfixtures "github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func prepareTestDatabase(t *testing.T, dsn string) {
	t.Helper()

	var err error

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	require.NoError(t, db.Ping())

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "migrations"), "postgres", driver)
	require.NoError(t, err)

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("testdata/fixtures"),
	)
	require.NoError(t, err)

	require.NoError(t, fixtures.Load())
}

// setupDatabase spins up a new Postgres container and returns a closure
// please always make sure to call the closure as it is the teardown function
func setupDatabase(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	var dsn string

	containerReq := testcontainers.ContainerRequest{
		Image:        "postgres:latest",
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp"),
		Env: map[string]string{
			"POSTGRES_DB":       "betterstacktest",
			"POSTGRES_PASSWORD": "betterstack",
			"POSTGRES_USER":     "betterstack",
		},
	}

	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: containerReq,
			Started:          true,
		})

	require.NoError(t, err)

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	require.NoError(t, err)

	dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "betterstack", "betterstack",
		fmt.Sprintf("localhost:%s", port.Port()), "betterstacktest")

	prepareTestDatabase(t, dsn)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)

	require.NoError(t, db.Ping())

	return db, func() {
		err := dbContainer.Terminate(context.Background())
		require.NoError(t, err)
	}
}

func TestURLRepositoryTable_Create(t *testing.T) {
	client, teardownFunc := setupDatabase(t)
	defer teardownFunc()

	userDB := NewUserRepository(client)

	require.NoError(t, userDB.Create(context.Background(), &User{
		Email:    "ken@unix.org",
		FullName: "Ken Thompson",
	}))
	//
}

func TestURLRepositoryTable_Get(t *testing.T) {
	client, teardownFunc := setupDatabase(t)
	defer teardownFunc()

	userDB := NewUserRepository(client)

	// take a look at testdata/fxtures/users.yml
	// this email exists there so we must be able to fetch it
	_, err := userDB.Get(context.Background(), "john.doe@gmail.com")
	require.NoError(t, err)

	email := "test@test.com"
	firstName := "Ken Thompson"

	// email does not exist here
	_, err = userDB.Get(context.Background(), email)
	require.Error(t, err)

	require.NoError(t, userDB.Create(context.Background(), &User{
		Email:    email,
		FullName: firstName,
	}))

	// fetch the same email again
	user, err := userDB.Get(context.Background(), email)
	require.NoError(t, err)

	require.Equal(t, email, user.Email)
	require.Equal(t, firstName, user.FullName)
}
