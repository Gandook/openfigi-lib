package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Gandook/openfigi-lib/openfigi"
)

// printUsageGuide prints a usage guide to stderr.
func printUsageGuide() {
	_, err := fmt.Fprintf(os.Stderr, `Usage: openfigi-cli <command> [options]

Commands:
	generate	Generate new OpenFIGI symbols and return them all at once
	genstream	Generate new OpenFIGI symbols and return them as a stream (one-by-one)
	validate	Check if a given string is a valid OpenFIGI symbol
	valstream	Validate existing OpenFIGI symbols from a file or stdin and return the 
			results as a stream (one-by-one)
`)
	if err != nil {
		return
	}
}

// runGenerate executes a "generate" command to generate valid OpenFIGI symbols.
func runGenerate(svc openfigi.FIGIService, args []string) error {
	command := flag.NewFlagSet("generate", flag.ExitOnError)
	n := command.Uint("n", 1, "Number of symbols to generate")
	err := command.Parse(args)
	if err != nil {
		return err
	}

	symbols := svc.Generate(*n)
	for _, symbol := range symbols {
		fmt.Println(symbol)
	}

	return nil
}

// runGenstream executes a "genstream" command to generate valid OpenFIGI symbols.
func runGenstream(ctx context.Context, svc openfigi.FIGIService, args []string) error {
	command := flag.NewFlagSet("genstream", flag.ExitOnError)
	n := command.Uint("n", 1, "Number of symbols to generate")
	err := command.Parse(args)
	if err != nil {
		return err
	}

	rcvChan := svc.GenerateStream(ctx, *n)

	for {
		select {
		case <-ctx.Done(): // Unexpected interruption
			return ctx.Err()
		case symbol, ok := <-rcvChan:
			if !ok {
				return nil
			}
			fmt.Println(symbol)
		}
	}
}

// runValidate executes a "validate" command to validate a given string.
func runValidate(svc openfigi.FIGIService, args []string) error {
	command := flag.NewFlagSet("validate", flag.ExitOnError)
	s := command.String("s", "", "String to validate")
	err := command.Parse(args)
	if err != nil {
		return err
	}

	isValid, message := svc.Validate(*s)

	if isValid {
		fmt.Println("Valid")
	} else {
		fmt.Printf("Invalid (Reason: %s)\n", message)
	}

	return nil
}

// runValstream executes a "valstream" command to validate OpenFIGI symbols from a file or stdin.
func runValstream(ctx context.Context, svc openfigi.FIGIService, args []string) error {
	command := flag.NewFlagSet("valstream", flag.ExitOnError)
	err := command.Parse(args)
	if err != nil {
		return err
	}

	var reader io.Reader

	if command.NArg() == 0 { // No file provided, use stdin instead
		reader = os.Stdin
	} else {
		file, openingErr := os.Open(command.Arg(0))
		if openingErr != nil {
			return openingErr
		}

		defer func(file *os.File) {
			err = file.Close()
			if err != nil {
				log.Fatalf("Error closing file: %v", err)
			}
		}(file)

		reader = file
	}

	rcvChan := svc.ValidateStream(ctx, reader)

	for {
		select {
		case <-ctx.Done(): // Unexpected interruption
			return ctx.Err()
		case result, ok := <-rcvChan:
			if !ok {
				return nil
			}

			fmt.Printf("%s is ", result.Input)
			if result.IsValid {
				fmt.Println("valid")
			} else {
				fmt.Printf("invalid (reason: %s)\n", result.Message)
			}
		}
	}
}

func main() {
	ctx := context.Background()
	service := openfigi.NewService()

	// Printing a usage guide if no arguments are provided.
	if len(os.Args) < 2 {
		printUsageGuide()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		if err := runGenerate(service, os.Args[2:]); err != nil {
			log.Fatalf("Error in generate command: %v", err)
		}
	case "genstream":
		if err := runGenstream(ctx, service, os.Args[2:]); err != nil {
			log.Fatalf("Error in genstream command: %v", err)
		}
	case "validate":
		if err := runValidate(service, os.Args[2:]); err != nil {
			log.Fatalf("Error in validate command: %v", err)
		}
	case "valstream":
		if err := runValstream(ctx, service, os.Args[2:]); err != nil {
			log.Fatalf("Error in valstream command: %v", err)
		}
	default:
		printUsageGuide()
		os.Exit(1)
	}
}
