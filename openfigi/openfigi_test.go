package openfigi

import (
	"context"
	"os"
	"testing"
)

func TestCharValue(t *testing.T) {
	t.Run("Digit", func(t *testing.T) {
		value := charValue('3')
		if value != 3 {
			t.Errorf("Expected 3 for '3', got %d.", value)
		}
	})

	t.Run("Letter", func(t *testing.T) {
		value := charValue('D')
		if value != 13 {
			t.Errorf("Expected 13 for 'D', got %d.", value)
		}
	})
}

func TestCharValueWithPos(t *testing.T) {
	t.Run("Odd Position", func(t *testing.T) {
		value := charValueWithPos('7', 5)
		if value != 14 {
			t.Errorf("Expected 14 for '7' at position 5, got %d.", value)
		}
	})

	t.Run("Even Position", func(t *testing.T) {
		value := charValueWithPos('Y', 8)
		if value != 34 {
			t.Errorf("Expected 34 for 'Y' at position 8, got %d.", value)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("Valid Symbols", func(t *testing.T) {
		service := NewService()

		var valid bool
		var message string

		if valid, message = service.Validate("BBG00HLH6Y37"); !valid {
			t.Errorf("Expected 'BBG00HLH6Y37' to be valid, got invalid with the message: %s.", message)
		}

		if valid, message = service.Validate("BBG003B5WQD2"); !valid {
			t.Errorf("Expected 'BBG003B5WQD2' to be valid, got invalid with the message: %s.", message)
		}

		if valid, message = service.Validate("KKG012C5GMZ5"); !valid {
			t.Errorf("Expected 'KKG012C5GMZ5' to be valid, got invalid with the message: %s.", message)
		}
	})

	t.Run("Invalid Inputs (Pattern Mismatch)", func(t *testing.T) {
		service := NewService()

		var valid bool
		var message string

		if valid, message = service.Validate("BKG00HLH6Y37"); message != "pattern mismatch" {
			t.Errorf("Expected 'BKG00HLH6Y37' to be a pattern mismatch, "+
				"got %v with the message: %s.", valid, message)
		}

		if valid, message = service.Validate("BBG00HLH6E37"); message != "pattern mismatch" {
			t.Errorf("Expected 'BBG00HLH6E37' to be a pattern mismatch, "+
				"got %v with the message: %s.", valid, message)
		}

		if valid, message = service.Validate("BBG0HLH6Y37"); message != "pattern mismatch" {
			t.Errorf("Expected 'BBG0HLH6Y37' to be a pattern mismatch, "+
				"got %v with the message: %s.", valid, message)
		}

		if valid, message = service.Validate("BBG00HLH6Y3H"); message != "pattern mismatch" {
			t.Errorf("Expected 'BBG00HLH6Y3H' to be a pattern mismatch, "+
				"got %v with the message: %s.", valid, message)
		}
	})

	t.Run("Invalid Inputs (Checksum Mismatch)", func(t *testing.T) {
		service := NewService()

		var valid bool
		var message string

		if valid, message = service.Validate("BBG0088JSC34"); message != "invalid checksum" {
			t.Errorf("Expected 'BBG0088JSC34' to be a checksum mismatch, "+
				"got %v with the message: %s.", valid, message)
		}

		if valid, message = service.Validate("BBG01J952TC0"); message != "invalid checksum" {
			t.Errorf("Expected 'BBG01J952TC0' to be a checksum mismatch, "+
				"got %v with the message: %s.", valid, message)
		}

		if valid, message = service.Validate("KKG019FZ8N78"); message != "invalid checksum" {
			t.Errorf("Expected 'KKG019FZ8N78' to be a checksum mismatch, "+
				"got %v with the message: %s.", valid, message)
		}
	})
}

func TestValidateStream(t *testing.T) {
	t.Run("Valid Symbols", func(t *testing.T) {
		file, err := os.Open("testdata/valid_symbols.txt")
		if err != nil {
			t.Fatalf("Failed to open the test data file: %v", err)
		}
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				t.Fatalf("Failed to close the test data file: %v", err)
			}
		}(file)

		ctx := context.Background()
		service := NewService()

		resultsChan := service.ValidateStream(ctx, file)
		for result := range resultsChan {
			if !result.IsValid {
				t.Errorf("Expected '%s' to be valid, got invalid with the message: %s.",
					result.Input, result.Message)
			}
		}
	})

	t.Run("Invalid Inputs", func(t *testing.T) {
		file, err := os.Open("testdata/invalid_inputs.txt")
		if err != nil {
			t.Fatalf("Failed to open the test data file: %v", err)
		}
		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				t.Fatalf("Failed to close the test data file: %v", err)
			}
		}(file)

		ctx := context.Background()
		service := NewService()

		resultsChan := service.ValidateStream(ctx, file)
		for result := range resultsChan {
			if result.Message != "invalid checksum" {
				t.Errorf("Expected '%s' to be a checksum mismatch, "+
					"got %v with the message: %s.", result.Input, result.IsValid, result.Message)
			}
		}
	})
}
