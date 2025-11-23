# fpath

fpath is a micro evaluation language library for the Go programming language that provides a powerful yet simple way to query and transform in-memory data structures using a concise expression language. It compiles a small evaluation script which can be reused by providing input data, which it will then return the final evaluation. The library allows you to compile queries once and evaluate them multiple times with different input data, making it efficient for repeated operations.

The language supports arithmetic operations, comparisons, logical operations, conditional expressions, data access, and built-in functions - all with a deliberate left-to-right evaluation order that ensures predictable behavior.

## Features

- **Compile-once, evaluate-many**: Compile queries once and reuse with different input data
- **Left-to-right evaluation**: No operator precedence - expressions evaluate strictly left-to-right
- **Rich data type support**: Numbers, strings, booleans, lists, and maps
- **Comprehensive operators**: Arithmetic, comparison, logical, and ternary operations
- **Data access**: Indexing and slicing for lists, strings, and maps
- **Built-in functions**: Mathematical, utility, and sorting functions
- **String indexing**: Treat strings as lists of characters
- **Error handling**: Clear error messages for invalid operations

## Installation

```bash
go get github.com/fletcharoo/fpath
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/fletcharoo/fpath"
)

func main() {
    // Compile a query once
    query, err := fpath.Compile("2 + 3 * 4")
    if err != nil {
        panic(err)
    }

    // Evaluate with no input data needed
    result, _ := query.Evaluate(nil)
    fmt.Printf("Result: %.0f\n", result) // Result: 20 (left-to-right evaluation)
}
```

## Syntax Guide

### Data Types

| Type | Description | Example |
|------|-------------|---------|
| Numbers | Integer and floating-point numbers | `42`, `3.14`, `-5.2` |
| Strings | Text values in quotes | `"hello world"` |
| Booleans | True/false values | `true`, `false` |
| Lists | Ordered collections of values | `[1, 2, 3]`, `["a", "b", "c"]` |
| Maps | Key-value pairs | `{"key": "value", "count": 10}` |
| Input reference | Refers to the input data | `$` |

### Arithmetic Operators

All arithmetic operations are evaluated **left-to-right** (no operator precedence):

| Operator | Description | Example | Result |
|----------|-------------|---------|--------|
| `+` | Addition | `2 + 3` | `5` |
| `-` | Subtraction | `10 - 3 - 2` | `5` |
| `*` | Multiplication | `3 * 4` | `12` |
| `/` | Division | `7 / 2` | `3.5` |
| `//` | Integer division | `7 // 2` | `3` |
| `%` | Modulo | `7 % 3` | `1` |
| `^` | Exponentiation | `2 ^ 3` | `8` |

### Comparison Operators

| Operator | Description | Example | Result |
|----------|-------------|---------|---------|
| `==` | Equal to | `5 == 5` | `true` |
| `!=` | Not equal to | `"hello" != "world"` | `true` |
| `<` | Less than | `3 < 10` | `true` |
| `<=` | Less than or equal to | `5 <= 5` | `true` |
| `>` | Greater than | `10 > 5` | `true` |
| `>=` | Greater than or equal to | `10 >= 10` | `true` |

### Logical Operators

| Operator | Description | Example | Result |
|----------|-------------|---------|---------|
| `&&` | Logical AND | `true && true` | `true` |
| `\|\|` | Logical OR | `true \|\| false` | `true` |

### Ternary Operator

```
condition ? true_expression : false_expression
```

Example: `5 > 3 ? "greater" : "less"` evalutes to `"greater"`

### Indexing and Slicing

| Operation | Description | Example | Result |
|-----------|-------------|---------|---------|
| List indexing | Access list element by index | `[1, 2, 3][0]` | `1` |
| String indexing | Access character by index | `"hello"[0]` | `"h"` |
| Map indexing | Access map value by key | `{"name": "Alice"}["name"]` | `"Alice"` |
| List slicing | Slice list from start to end | `[1, 2, 3, 4, 5][1:3]` | `[1, 2]` |
| String slicing | Slice string from start to end | `"hello"[1:4]` | `"ell"` |

### Built-in Functions

| Function | Description | Example | Result |
|----------|-------------|---------|---------|
| `len(value)` | Get length of string, list, or map | `len("hello")` | `5` |
| `filter(list, condition)` | Filter list items by condition | `filter([1, 2, 3, 4, 5], _ > 3)` | `[4, 5]` |
| `contains(haystack, needle)` | Check if list/string/map contains value | `contains([1, 2, 3], 2)` | `true` |
| `abs(number)` | Absolute value | `abs(-5)` | `5` |
| `min(values...)` | Minimum value | `min(1, 5, 3)` | `1` |
| `max(values...)` | Maximum value | `max(1, 5, 3)` | `5` |
| `round(number)` | Round to nearest integer | `round(3.7)` | `4` |
| `floor(number)` | Round down to integer | `floor(3.7)` | `3` |
| `ceil(number)` | Round up to integer | `ceil(3.2)` | `4` |
| `sort(value)` | Sort lists and strings in ascending order | `sort([3, 1, 2])` | `[1, 2, 3]` |

**Note**: For mixed-type lists, `sort()` uses type hierarchy: numbers < strings < booleans

**Note**: In `filter()`, the underscore `_` represents the current item being evaluated.

## Examples

### Data Filtering

```go
// Filter numbers greater than 3
query, _ := fpath.Compile("filter([1, 2, 3, 4, 5], _ > 3)")
result, _ := query.Evaluate(nil)
// Result: [4, 5]

// Filter strings with length 5
query, _ := fpath.Compile("filter([\"hello\", \"world\", \"hi\", \"test\"], len(_) == 5)")
result, _ := query.Evaluate(nil)
// Result: ["hello", "world"]
```

### Conditional Logic

```go
// Simple conditional
query, _ := fpath.Compile("5 > 3 ? \"greater\" : \"less\"")
result, _ := query.Evaluate(nil)
// Result: "greater"

// Nested conditional
query, _ := fpath.Compile("10 > 5 ? (5 > 2 ? \"yes\" : \"no\") : \"never\"")
result, _ := query.Evaluate(nil)
// Result: "yes"
```

### Mathematical Calculations

```go
// Left-to-right arithmetic (no precedence)
query, _ := fpath.Compile("2 + 3 * 4")
result, _ := query.Evaluate(nil)
// Result: 20 (not 14, because (2 + 3) * 4)

// Complex calculation
query, _ := fpath.Compile("(10 * 1.1) - 2")
result, _ := query.Evaluate(nil)
// Result: 9.0
```

### String Operations

```go
// String indexing
query, _ := fpath.Compile("\"hello\"[0]")
result, _ := query.Evaluate(nil)
// Result: "h"

// String slicing
query, _ := fpath.Compile("\"hello world\"[1:5]")
result, _ := query.Evaluate(nil)
// Result: "ello"

// String contains
query, _ := fpath.Compile("contains(\"hello world\", \"world\")")
result, _ := query.Evaluate(nil)
// Result: true
```

### List Operations

```go
// List indexing
query, _ := fpath.Compile("[1, 2, 3, 4, 5][2]")
result, _ := query.Evaluate(nil)
// Result: 3

// List slicing
query, _ := fpath.Compile("[\"a\", \"b\", \"c\", \"d\", \"e\"][1:3]")
result, _ := query.Evaluate(nil)
// Result: ["b", "c"]

// List calculations
query, _ := fpath.Compile("([10, 20, 30][0] + [10, 20, 30][1] + [10, 20, 30][2]) / 3")
result, _ := query.Evaluate(nil)
// Result: 20.0
```

### Map Operations

```go
// Map indexing
query, _ := fpath.Compile("{\"name\": \"Alice\", \"age\": 25}[\"name\"]")
result, _ := query.Evaluate(nil)
// Result: "Alice"

// Check map contains key
query, _ := fpath.Compile("contains({\"a\": 1, \"b\": 2}, \"a\")")
result, _ := query.Evaluate(nil)
// Result: true
```

### Mathematical Functions

```go
// Absolute value
query, _ := fpath.Compile("abs(-5)")
result, _ := query.Evaluate(nil)
// Result: 5

// Min and max
query, _ := fpath.Compile("min(1, 5, 3)")
result, _ := query.Evaluate(nil)
// Result: 1

query, _ := fpath.Compile("max(1, 5, 3)")
result, _ := query.Evaluate(nil)
// Result: 5

// Rounding
query, _ := fpath.Compile("round(3.7)")
result, _ := query.Evaluate(nil)
// Result: 4

query, _ := fpath.Compile("floor(3.7)")
result, _ := query.Evaluate(nil)
// Result: 3

query, _ := fpath.Compile("ceil(3.2)")
result, _ := query.Evaluate(nil)
// Result: 4
```

### Sorting Operations

```go
// Sort lists
query, _ := fpath.Compile("sort([3, 1, 2])")
result, _ := query.Evaluate(nil)
// Result: [1, 2, 3]

// Sort strings (character sorting)
query, _ := fpath.Compile("sort(\"cba\")")
result, _ := query.Evaluate(nil)
// Result: "abc"

// Sort mixed-type lists (numbers < strings < booleans)
query, _ := fpath.Compile("sort([true, \"hello\", 42])")
result, _ := query.Evaluate(nil)
// Result: [42, "hello", true]

// Sort with input data
query, _ := fpath.Compile("sort($)")
result, _ := query.Evaluate([3, 1, 2])
// Result: [1, 2, 3]
```

### Error Handling

```go
query, err := fpath.Compile("2 + + 3")
if err != nil {
    // Handle compilation error
    log.Fatal(err)
}

result, err := query.Evaluate(nil)
if err != nil {
    // Handle evaluation error (e.g., division by zero)
    log.Fatal(err)
}
```

