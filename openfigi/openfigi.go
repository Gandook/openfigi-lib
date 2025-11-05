package openfigi

import (
	"bufio"
	"context"
	"io"
	"math/rand"
	"regexp"
	"sync"
	"time"
	"unicode"
)

// chanBufferSize is used as the buffer size for the channels returned by ValidateStream and
// GenerateStream.
const chanBufferSize = 100

// A valid OpenFIGI symbol:
// 1. Starts with either "BBG" or "KKG".
// 2. Contains 8 alphanumeric characters (vowels are not allowed) after that.
// 3. Contains a checksum digit at the end.
var validFIGIPattern = regexp.MustCompile(`^(BB|KK)G[0-9BCDFGHJKLMNPQRSTVWXYZ]{8}[0-9]$`)

// ValidationResult represents the result of a validation operation.
// Input is the examined string.
// IsValid is true if and only if Input is a valid OpenFIGI symbol.
// Message is a human-readable description of the result. If Input is invalid, Message will
// contain additional information about the reason.
type ValidationResult struct {
	Input   string
	IsValid bool
	Message string
}

type FIGIService interface {
	// Validate receives a string and determines if it is a valid OpenFIGI symbol.
	Validate(figi string) (bool, string)
	// ValidateStream reads a large number of strings from an external source (e.g., a file),
	// validates them, and returns the results via a channel.
	ValidateStream(ctx context.Context, reader io.Reader) <-chan ValidationResult

	// Generate(n uint32) []string
	// GenerateStream(ctx context.Context, n uint32) <-chan string
}

// defaultFIGIService implements the FIGIService interface.
// rng is a random number generator used for generating new OpenFIGI symbols.
// rngLock is a mutex used to ensure no two goroutines use rng at the same time.
type defaultFIGIService struct {
	rng     *rand.Rand
	rngLock sync.Mutex
}

// NewService creates a new FIGIService instance.
func NewService() FIGIService {
	return &defaultFIGIService{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// charValue calculates the corresponding value of a character in Luhn's algorithm, NOT
// considering its position in the string.
// The value of a digit is the same as the digit itself.
// The value of a letter is its (0-based) position in the alphabet + 10 (A = 10, B = 11, etc.).
func charValue(r rune) int {
	if unicode.IsDigit(r) {
		return int(r - '0')
	} else {
		return int(r - 'A' + 10)
	}
}

// charValueWithPos calculates the corresponding value of a character in Luhn's algorithm,
// considering its position in the string.
// In the algorithm, the value of the characters in odd positions is doubled (except for the
// checksum digit).
// Note that the position numbers start from 0, not 1.
func charValueWithPos(r rune, pos int) int {
	if pos%2 == 0 || pos == 11 {
		return charValue(r)
	} else {
		return charValue(r) * 2
	}
}

// Validate receives a string and determines if it is a valid OpenFIGI symbol.
func (d *defaultFIGIService) Validate(figi string) (bool, string) {
	if !validFIGIPattern.MatchString(figi) {
		return false, "pattern mismatch"
	}

	digitSum := 0
	currentCharValue := 0

	for i, r := range figi {
		currentCharValue = charValueWithPos(r, i)
		digitSum += (currentCharValue / 10) + (currentCharValue % 10)
	}

	if digitSum%10 != 0 {
		return false, "invalid checksum"
	}

	return true, "valid"
}

// ValidateStream reads a large number of strings from an external source (e.g., a file),
// validates them, and returns the results via a channel.
func (d *defaultFIGIService) ValidateStream(ctx context.Context, reader io.Reader) <-chan ValidationResult {
	out := make(chan ValidationResult, chanBufferSize)

	go func() {
		defer close(out)

		scanner := bufio.NewScanner(reader)

		var input, message string
		var isValid bool
		var result ValidationResult

		for scanner.Scan() {
			input = scanner.Text()
			isValid, message = d.Validate(input)

			result = ValidationResult{
				Input:   input,
				IsValid: isValid,
				Message: message,
			}

			select {
			case <-ctx.Done():
				return
			case out <- result:
				// The result is sent.
			}
		}
	}()

	return out
}
