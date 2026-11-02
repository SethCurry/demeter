package dscript

import (
	"errors"
	"strings"
	"testing"

	"github.com/glycerine/zygomys/v9/zygo"
)

func TestInstanceRun(t *testing.T) {
	instance := NewInstance()
	t.Cleanup(func() {
		_ = instance.env.Close()
	})

	if err := instance.env.LoadString(`(+ 1 2)`); err != nil {
		t.Fatalf("failed to load expression: %v", err)
	}

	got, err := instance.Run()
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}

	if got != "3" {
		t.Errorf("Run() = %q, want %q", got, "3")
	}
}

func TestInstanceRunWrapsErrorsWithStackTrace(t *testing.T) {
	instance := NewInstance()
	t.Cleanup(func() {
		_ = instance.env.Close()
	})

	if err := instance.env.LoadString(`(example)`); err != nil {
		t.Fatalf("failed to load expression: %v", err)
	}

	_, err := instance.Run()
	if err == nil {
		t.Fatal("Run() succeeded, want an error from the example function")
	}

	var scriptErr *Error
	if !errors.As(err, &scriptErr) {
		t.Fatalf("Run() error = %T, want *Error", err)
	}

	if !strings.Contains(err.Error(), "function not implemented: example") {
		t.Errorf("unexpected error message: %v", err)
	}

	if scriptErr.StackTrace == "" {
		t.Error("expected a non-empty stack trace")
	}
}

func TestExampleFunctionReturnsNotImplementedError(t *testing.T) {
	_, err := ExampleFunction(nil, "example", []zygo.Sexp{})
	if err == nil {
		t.Fatal("ExampleFunction() succeeded, want an error")
	}

	if !strings.Contains(err.Error(), "function not implemented: example") {
		t.Errorf("unexpected error message: %v", err)
	}
}
