package database

import (
	"testing"
)

func TestCloseNilDB(t *testing.T) {
	err := Close(nil)
	if err != nil {
		t.Errorf("expected no error for nil db, got: %v", err)
	}
}
