## Project Overview
`fpath` is a micro evaluation language library for the Go programming language. It compiles a small evaluation script which can be reused by providing input data, which it will then return the final evaluation.

### Operator Precedence
fpath deliberately implements no operator precedence, evaluating all arithmetic expressions strictly left to right. This design choice simplifies the language and ensures predictable evaluation order, so `2 + 3 * 4` evaluates to `20` rather than the mathematical result of `14`. When implementing or parsing expressions, always process operators in the order they appear without applying traditional precedence rules.

### String Indexing Support
Similar to other languages, strings in fpath are treated as a list of characters (which are also strings in fpath). This means any functions or functionality that can accept a list (like indexing individual items) will also work with strings. For example:
- `"hello"[0]` returns `"h"`
- `"world"[2]` returns `"r"`
- Indexing operations with expressions like `"hello"[1+1]` work as expected
- Chained indexing with mixed types works: `[[1, 2], "hello"][1][0]` returns `"h"`
- All error handling for out-of-bounds or invalid indices is consistent with list indexing

## Common Developer Commands
- `make test`: Run all tests
- `make testupdate`: Run all tests and update snapshots

## Important Files
- `go.mod`: Go module definition with dependencies for the fpath library
- `Makefile`: Build automation with test and snapshot update commands
- `README.md`: Project documentation describing the micro language for querying in-memory data
- `internal/lexer/lexer.go`: Lexical analyzer that tokenizes input strings for parsing
- `internal/parser/parser.go`: Parser that converts tokens into an abstract syntax tree (AST)
- `internal/parser/ast.go`: Abstract syntax tree node definitions and expression interfaces
- `internal/runtime/runtime.go`: Expression evaluator that executes parsed AST against input data
- `internal/shared.go`: Utility functions for path-based data lookup in nested structures

## Development Standards

### Error Handling in Tests
This project follows strict error handling standards to ensure test reliability and maintainability:

#### Error Type Definitions
- Export specific error types for common error conditions
- Define errors in the relevant package (e.g., `runtime.ErrIncompatibleTypes`, `parser.ErrInvalidDecode`)
- Use `errors.New()` for base errors and `fmt.Errorf("%w: ...")` for wrapped errors with context

#### Test Error Validation
- **NEVER** use string-based error checking like `require.Contains(t, err.Error(), "error text")`
- **ALWAYS** use type-based error checking with `require.ErrorIs(t, err, ExpectedErrorType)`
- This makes tests less fragile when error messages change or additional context is added

#### Examples
```go
// ❌ BAD - Fragile string matching
require.Contains(t, err.Error(), "incompatible types")

// ✅ GOOD - Type-based checking  
require.ErrorIs(t, err, runtime.ErrIncompatibleTypes)
```

#### Current Error Types
- `runtime.ErrIncompatibleTypes`: Type mismatch errors in operations
- `runtime.ErrDivisionByZero`: Division by zero errors
- `runtime.ErrBooleanOperation`: Invalid boolean operation errors
- `parser.ErrInvalidDecode`: Expression decoding errors
- `parser.ErrUndefinedToken`: Undefined token errors
- `parser.ErrExpectedToken`: Missing expected token errors

### Debugging Practices
This project follows a test-driven debugging approach:

#### NO Debugging Files
- **NEVER** create separate debug files (e.g., `debug.go`, `debug_parser.go`, `temp.go`)
- Debug files cause issues including redefinition of helper functions across packages, syntax conflicts with existing code, additional cleanup work required, unclear separate between production and debug code.

#### Test-Based Debugging
- **ALWAYS** debug within existing test files (`*_test.go`)
- Add temporary test cases with descriptive names like `TestDebug_SpecificIssue`
- Use `t.Logf()` for debug output instead of `fmt.Println()`
- Mark debug tests clearly with comments: `// DEBUG: Investigate XYZ issue`
- Remove or convert debug tests to proper test cases after debugging is complete

#### Examples
```go
// ❌ BAD - Creating a debug file
// File: internal/runtime/debug.go
func debugHelper() { /* ... */ }

// ✅ GOOD - Debug within existing test file
// File: internal/runtime/runtime_test.go
func TestDebug_TypeMismatchIssue(t *testing.T) {
    // DEBUG: Investigating type conversion edge case
    result := evaluateExpression(...)
    t.Logf("Result type: %T, value: %v", result, result)
    require.Equal(t, expected, result)
}
```

#### Temporary Debug Code
- If you need helper functions for debugging, define them inline within the test function
- Use `t.Helper()` to mark helper functions properly
- Always remove temporary debug code before considering work complete

## Project Structure
```
fpath/
├── internal/
│   ├── lexer/
│   │   ├── lexer.go          # Tokenizes input strings into lexical tokens
│   │   └── lexer_test.go     # Tests for lexer functionality
│   ├── parser/
│   │   ├── ast.go            # AST node definitions and expression interfaces
│   │   ├── parser.go         # Parses tokens into executable AST
│   │   └── parser_test.go    # Tests for parser functionality
│   ├── runtime/
│   │   ├── runtime.go        # Evaluates AST expressions against input data
│   │   ├── runtime_test.go   # Tests for runtime evaluation
│   │   └── __snapshots__/    # Snapshot test files
│   ├── shared.go             # Shared utilities for data path lookup
│   └── shared_test.go        # Tests for shared utilities
├── AGENTS.md                 # AI assistant guidance documentation
├── go.mod                    # Go module definition and dependencies
├── go.sum                    # Go module dependency checksums
├── LICENSE                   # Project license file
├── Makefile                  # Build automation and test commands
└── README.md                 # Project overview and documentation
```
