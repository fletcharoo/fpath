## Project Overview
`fpath` is a micro evaluation language library for the Go programming language. It compiles a small evaluation script which can be reused by providing input data, which it will then return the final evaluation.

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
