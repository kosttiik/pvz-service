package database

import (
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASSWORD", "postgres")
	os.Setenv("DB_NAME", "pvz_db")

	if err := Connect(); err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	if DB == nil {
		t.Error("DB should not be nil after successful connection")
	}

	os.Setenv("DB_PORT", "1234")
	if err := Connect(); err == nil {
		t.Error("Connect() should fail with invalid port")
	}
}
