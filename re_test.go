package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestSplitArguments(t *testing.T) {
	t.Run("arguments have less than two element should be error", func(t *testing.T) {
		args := []string{"re"}

		_, _, err := splitCommand(args)

		eErr := errors.New("you should add command after re [command], e.g. 're go test -v .'")
		if eErr.Error() != err.Error() {
			t.Errorf("expect error is '%v' but got '%v'", eErr, err)
		}
	})

	t.Run("arguments should have two element at least", func(t *testing.T) {
		args := []string{"re", "go"}

		_, _, err := splitCommand(args)

		if err != nil {
			t.Errorf("expect with out any error but got %v", err)
		}
	})

	t.Run("arguments should have two element at least", func(t *testing.T) {
		args := []string{"re", "go"}

		_, _, err := splitCommand(args)

		if err != nil {
			t.Errorf("expect with out any error but got %v", err)
		}
	})

	t.Run("arguments index 1 should be command", func(t *testing.T) {
		args := []string{"re", "go", "test", "-v", "."}

		prog, _, _ := splitCommand(args)

		if "go" != prog {
			t.Errorf("expect prog is %q but got %q", "go", prog)
		}
	})

	t.Run("arguments index 2 until the end should be params of the command", func(t *testing.T) {
		args := []string{"re", "go", "test", "-v", "."}

		_, params, _ := splitCommand(args)

		eParams := []string{"test", "-v", "."}
		if !reflect.DeepEqual(eParams, params) {
			t.Errorf("expect params is %q but got %q", eParams, params)
		}

	})
}
