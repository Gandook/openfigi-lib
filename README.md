# openfigi-lib

This project is a Go library and command-line tool for validating and generating OpenFIGI symbols.

## Features

* **Validate:** Checks if a symbol (or a stream of symbols from a file/stdin) adheres to the OpenFIGI format, including prefixes (`BBG`/`KKG`), allowed characters, and the checksum.
* **Generate:** Generates any number of valid, random OpenFIGI symbols.

## Project Structure

The project is split into a reusable library and a simple command-line wrapper.

```text
/
├── go.mod
├── /openfigi/           <-- The core, reusable library (package openfigi)
└── /cmd/
    └── /openfigi-cli/   <-- The command-line application (package main)
```

## How to Build

1.  Clone the repository:
    ```shell
    git clone https://github.com/Gandook/openfigi-lib.git
    cd openfigi-lib
    ```

2.  Build the `openfigi-cli` executable. This command compiles the application and places the binary in your current directory.
    ```shell
    go build ./cmd/openfigi-cli/
    ```
    This will create an executable file named `openfigi-cli` (or `openfigi-cli.exe` on Windows).

## How to Run

The application has four commands: `generate`, `genstream`, `validate`, and `valstream`.

### Generate Symbols

Use the `generate` command to create new symbols. The generated symbols are returned all at once.

**Flags:**
* `-n <number>`: Number of symbols to generate (default: 1).

**Example (Linux/macOS/PowerShell):**
```shell
# Generate 5 unique symbols
./openfigi-cli generate -n 5
```

**Example (Windows Command Prompt):**
```shell
# Generate 10 unique symbols
.\openfigi-cli.exe generate -n 10
```

The `genstream` command is similar, but instead of returning all symbols at once, it returns them one at a time (as a stream). This makes it ideal for creating a large number of symbols.

**Flags:**
* `-n <number>`: Number of symbols to generate (default: 1).

**Example (Linux/macOS/PowerShell):**
```shell
# Generate 5 unique symbols
./openfigi-cli genstream -n 5
```

**Example (Windows Command Prompt):**
```shell
# Generate 10 unique symbols
.\openfigi-cli.exe genstream -n 10
```

### Validate Symbols

Use the `validate` command to validate a single string.

**Flags:**
* `-s <string>`: The string to validate.

**Example (Linux/macOS/PowerShell):**
```shell
# Validate a single string
./openfigi-cli validate -s KKGW2Q0XJFQ5
```

Use the `valstream` command to check existing symbols. It can read from a specified file or from standard input (stdin). Similar to `genstream`, it returns the results as a stream.

**Example (Read from a file):**
```shell
# Validate all symbols in a file
./openfigi-cli valstream path/to/your/symbols.txt
```

**Example (Read from standard input):**
```shell
# Pipe input from another command (e.g., cat)
cat path/to/your/symbols.txt | ./openfigi-cli valstream
```

### Running Tests

To run the included unit tests for the library:
```shell
go test ./openfigi/
```