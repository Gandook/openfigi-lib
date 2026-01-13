package openfigi

import (
	"context"
	"errors"
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

func TestGetDigitSum(t *testing.T) {
	var sum int

	sum = getDigitSum("BBG00HLH6Y37")
	if sum != 60 {
		t.Errorf("Expected 60 for 'BBG00HLH6Y37', got %d.", sum)
	}

	sum = getDigitSum("KKG012C5GMZ0")
	if sum != 45 {
		t.Errorf("Expected 45 for 'KKG012C5GMZ0', got %d.", sum)
	}
}

func TestValidate(t *testing.T) {
	t.Run("Valid Symbols", func(t *testing.T) {
		service := NewService()

		var err error

		if err = service.Validate("BBG00HLH6Y37"); err != nil {
			t.Errorf("Expected 'BBG00HLH6Y37' to be valid, got invalid with the message: %v.", err)
		}

		if err = service.Validate("BBG003B5WQD2"); err != nil {
			t.Errorf("Expected 'BBG003B5WQD2' to be valid, got invalid with the message: %v.", err)
		}

		if err = service.Validate("KKG012C5GMZ5"); err != nil {
			t.Errorf("Expected 'KKG012C5GMZ5' to be valid, got invalid with the message: %v.", err)
		}
	})

	t.Run("Invalid Inputs (Pattern Mismatch)", func(t *testing.T) {
		service := NewService()

		var err error

		if err = service.Validate("BKG00HLH6Y37"); !errors.Is(err, errPatternMismatch) {
			t.Errorf("Expected 'BKG00HLH6Y37' to be a pattern mismatch, got %v.", err)
		}

		if err = service.Validate("BBG00HLH6E37"); !errors.Is(err, errPatternMismatch) {
			t.Errorf("Expected 'BBG00HLH6E37' to be a pattern mismatch, got %v.", err)
		}

		if err = service.Validate("BBG0HLH6Y37"); !errors.Is(err, errPatternMismatch) {
			t.Errorf("Expected 'BBG0HLH6Y37' to be a pattern mismatch, got %v.", err)
		}

		if err = service.Validate("BBG00HLH6Y3H"); !errors.Is(err, errPatternMismatch) {
			t.Errorf("Expected 'BBG00HLH6Y3H' to be a pattern mismatch, got %v.", err)
		}
	})

	t.Run("Invalid Inputs (Checksum Mismatch)", func(t *testing.T) {
		service := NewService()

		var err error

		if err = service.Validate("BBG0088JSC34"); !errors.Is(err, errInvalidChecksum) {
			t.Errorf("Expected 'BBG0088JSC34' to be a checksum mismatch, got %v.", err)
		}

		if err = service.Validate("BBG01J952TC0"); !errors.Is(err, errInvalidChecksum) {
			t.Errorf("Expected 'BBG01J952TC0' to be a checksum mismatch, got %v.", err)
		}

		if err = service.Validate("KKG019FZ8N78"); !errors.Is(err, errInvalidChecksum) {
			t.Errorf("Expected 'KKG019FZ8N78' to be a checksum mismatch, got %v.", err)
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
			if result.Error != nil {
				t.Errorf("Expected '%s' to be valid, got invalid with the message: %v.",
					result.Input, result.Error)
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
			if !errors.Is(result.Error, errInvalidChecksum) {
				t.Errorf("Expected '%s' to be a checksum mismatch, got %v.",
					result.Input, result.Error)
			}
		}
	})
}

func TestGenerate(t *testing.T) {
	var symbolsNeeded uint = 10
	service := NewService()
	symbols := service.Generate(symbolsNeeded)

	if uint(len(symbols)) != symbolsNeeded {
		t.Errorf("Expected %d symbol(s), got %d.", symbolsNeeded, len(symbols))
	}

	var err error

	for _, symbol := range symbols {
		if err = service.Validate(symbol); err != nil {
			t.Errorf("Expected '%s' to be a valid OpenFIGI symbol, "+
				"got invalid with the message: %v.", symbol, err)
		}
	}
}

func TestGenerateStream(t *testing.T) {
	var symbolsNeeded uint = 100
	ctx := context.Background()
	service := NewService()
	var symbolCount uint = 0

	symbolsChan := service.GenerateStream(ctx, symbolsNeeded)
	for symbol := range symbolsChan {
		if err := service.Validate(symbol); err != nil {
			t.Errorf("Expected '%s' to be a valid OpenFIGI symbol, "+
				"got invalid with the message: %s.", symbol, err)
		}

		symbolCount++
	}

	if symbolCount != symbolsNeeded {
		t.Errorf("Expected %d symbol(s), got %d.", symbolsNeeded, symbolCount)
	}
}
