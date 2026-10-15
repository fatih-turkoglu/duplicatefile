package apperr

import (
	"errors"
	"fmt"
	"testing"
)

func TestError(t *testing.T) {
	base := errors.New("disk full")
	err := Errorf("db_move_failed", "could not copy database: %w", base)
	if err.Error() != "[db_move_failed] could not copy database: disk full" {
		t.Errorf("Error() = %q", err.Error())
	}
	if !errors.Is(err, base) {
		t.Error("Unwrap zinciri korunmalı")
	}
	wrapped := fmt.Errorf("context: %w", err)
	if Code(wrapped) != "db_move_failed" {
		t.Errorf("Code = %q", Code(wrapped))
	}
	if Code(base) != "" {
		t.Error("kodsuz hata boş kod vermeli")
	}
	if Message(New("busy", "x")) != "x" {
		t.Errorf("Message = %q", Message(New("busy", "x")))
	}
}
