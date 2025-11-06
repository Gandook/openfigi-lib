package openfigi

import (
	"bufio"
	"context"
	"io"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	// chanBufferSize is used as the buffer size for the channels returned by ValidateStream and
	// GenerateStream.
	chanBufferSize = 100
	// figiChars is the list of all valid characters in a valid OpenFIGI symbol.
	// generateChar uses this list to randomly generate a valid character.
	figiChars = "0123456789BCDFGHJKLMNPQRSTVWXYZ"
)

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

	// Generate generates n new valid OpenFIGI symbols.
	// Using this method to create a large number of symbols is NOT recommended.
	Generate(n uint) []string
	// GenerateStream generates n new valid OpenFIGI symbols and returns them via a channel.
	// This makes it ideal for creating a large number of symbols.
	GenerateStream(ctx context.Context, n uint) <-chan string
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

// getDigitSum calculates the sum of the digits in Luhn's algorithm.
func getDigitSum(s string) int {
	digitSum := 0
	currentCharValue := 0

	for i, r := range s {
		currentCharValue = charValueWithPos(r, i)
		digitSum += (currentCharValue / 10) + (currentCharValue % 10)
	}

	return digitSum
}

// Validate receives a string and determines if it is a valid OpenFIGI symbol.
func (d *defaultFIGIService) Validate(figi string) (bool, string) {
	if !validFIGIPattern.MatchString(figi) {
		return false, "pattern mismatch"
	}

	if getDigitSum(figi)%10 != 0 {
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

// generateChar randomly generates a valid OpenFIGI character.
func (d *defaultFIGIService) generateChar() byte {
	d.rngLock.Lock()
	defer d.rngLock.Unlock()

	return figiChars[d.rng.Intn(31)]
}

// generateSingle randomly generates a single valid OpenFIGI symbol.
func (d *defaultFIGIService) generateSingle() string {
	var sb strings.Builder

	d.rngLock.Lock()
	if d.rng.Intn(2) == 0 {
		sb.WriteString("BBG")
	} else {
		sb.WriteString("KKG")
	}
	d.rngLock.Unlock()

	for i := 0; i < 8; i++ {
		sb.WriteByte(d.generateChar())
	}

	digitSum := getDigitSum(sb.String())
	checksum := (10 - (digitSum % 10)) % 10
	sb.WriteByte(byte('0' + checksum))

	return sb.String()
}

// Generate generates n new valid OpenFIGI symbols.
// Using this method to create a large number of symbols is NOT recommended.
func (d *defaultFIGIService) Generate(n uint) []string {
	isGenerated := make(map[string]bool)
	result := make([]string, 0, n)

	var newSymbolCandidate string

	for uint(len(result)) < n {
		newSymbolCandidate = d.generateSingle()

		if _, exists := isGenerated[newSymbolCandidate]; exists {
			continue
		}

		isGenerated[newSymbolCandidate] = true
		result = append(result, newSymbolCandidate)
	}

	return result
}

// GenerateStream generates n new valid OpenFIGI symbols and returns them via a channel.
// This makes it ideal for creating a large number of symbols.
func (d *defaultFIGIService) GenerateStream(ctx context.Context, n uint) <-chan string {
	out := make(chan string, chanBufferSize)
	isGenerated := make(map[string]bool)

	var newSymbolCandidate string

	go func() {
		defer close(out)

		for uint(len(isGenerated)) < n {
			newSymbolCandidate = d.generateSingle()

			if _, exists := isGenerated[newSymbolCandidate]; exists {
				continue
			}

			isGenerated[newSymbolCandidate] = true

			select {
			case <-ctx.Done():
				return
			case out <- newSymbolCandidate:
				// The new symbol is sent.
			}
		}
	}()

	return out
}
