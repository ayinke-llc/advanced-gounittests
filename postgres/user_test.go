package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

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
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		t.Fatal(err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", "migrations"), "postgres", driver)
	if err != nil {
		t.Fatal(err)
	}

	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatal(err)
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("testdata/fixtures"),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = fixtures.Load()
	if err != nil {
		t.Fatal(err)
	}
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
	if err != nil {
		t.Fatal(err)
	}

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", "betterstack", "betterstack",
		fmt.Sprintf("localhost:%s", port.Port()), "betterstacktest")

	prepareTestDatabase(t, dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	return db, func() {
		err := dbContainer.Terminate(context.Background())
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestURLRepositoryTable_Create(t *testing.T) {
	client, teardownFunc := setupDatabase(t)
	defer teardownFunc()

	userDB := NewUserRepository(client)

	err := userDB.Create(context.Background(), &User{
		Email:    "ken@unix.org",
		FullName: "Ken Thompson",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestURLRepositoryTable_Get(t *testing.T) {
	client, teardownFunc := setupDatabase(t)
	defer teardownFunc()

	userDB := NewUserRepository(client)

	// take a look at testdata/fxtures/users.yml
	// this email exists there so we must be able to fetch it
	_, err := userDB.Get(context.Background(), "john.doe@gmail.com")
	if err != nil {
		t.Fatal(err)
	}

	email := "test@test.com"
	firstName := "Ken Thompson"

	// email does not exist here
	_, err = userDB.Get(context.Background(), email)
	if err == nil {
		t.Fatal(errors.New("expected an error here. Email should not be found"))
	}

	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Unexpected database error. Expected %v got %v", sql.ErrNoRows, err)
	}

	err = userDB.Create(context.Background(), &User{
		Email:    email,
		FullName: firstName,
	})
	if err != nil {
		t.Fatal(err)
	}

	// fetch the same email again
	user, err := userDB.Get(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(email, user.Email) {
		t.Fatalf("retrieved values do not match. Expected %s, got %s", email, user.Email)
	}

	if !strings.EqualFold(firstName, user.FullName) {
		t.Fatalf("retrieved values do not match. Expected %s, got %s", firstName, user.FullName)
	}
}
