# TypeScript Code Analyzer

This tool recursively walks through TypeScript files in a directory and checks if functions contain a specific code block.

## Prerequisites

- Go 1.18 or higher
- Tree-sitter TypeScript package

## Installation

```bash
# Install the required Go package
go get github.com/smacker/go-tree-sitter@latest

# Build the analyzer
go build -o ./bin/ts-analyzer main.go
```

## Usage

```bash
# Run directly with the compiled binary
./bin/ts-analyzer -dir="/path/to/typescript/files" -code-block="using ctx = getContext()" [-regex=true|false] [-fn-types="exported,internal,callback"] [-file-glob="*.ts"] [-invert=true|false] [-verbose=true|false]

# Or build and run in one step
go run main.go -dir="/path/to/typescript/files" -code-block="using ctx = getContext()" [-regex=true|false] [-fn-types="exported,internal,callback"] [-file-glob="*.ts"] [-invert=true|false] [-verbose=true|false]
```

### Parameters

- `-dir`: Directory to recursively search for TypeScript files
- `-code-block`: Code block that should exist in each function
- `-regex`: (Optional) Treat the code block as a regular expression. Default is false.
- `-fn-types`: (Optional) Function types to check: 'exported', 'internal', 'callback', or a comma-separated combination. Default is "exported".
- `-file-glob`: (Optional) Pattern to match files to analyze. Default is "**/*.ts".
- `-invert`: (Optional) Invert the search to find functions that should NOT contain the code block. Default is false.
- `-verbose`: (Optional) Enable verbose output for debugging. Default is false.

## Examples

Check if all exported functions in the repositories package use the context with any variable name:

```bash
./bin/ts-analyzer -dir="./packages/repositories/src" -code-block="using [a-z_]+ = getContext\\(\\)" -regex=true
```

Check if any internal functions in the utils package contain a deprecated API call:

```bash
./bin/ts-analyzer -dir="./packages/utils/src" -code-block="deprecatedAPI()" -fn-types="internal" -invert=true
```

Check if all callback functions include error handling:

```bash
./bin/ts-analyzer -dir="./packages/repositories/src" -code-block="try {" -fn-types="callback"
```

Check all function types (exported, internal, and callbacks) for a specific pattern:

```bash
./bin/ts-analyzer -dir="./packages/repositories/src" -code-block="required()" -fn-types="exported,internal,callback"
```

## Use Cases

1. **Enforce coding standards**: Ensure all repository functions use context tracking
2. **Find deprecated code**: Identify functions using outdated APIs or patterns
3. **Security checks**: Locate functions missing required security checks
4. **Detect anti-patterns**: Find code that should not be using certain patterns
5. **Callback validation**: Ensure callback functions include proper error handling

## Output

The tool will list all TypeScript files that have functions missing the required code block (or containing forbidden code blocks when using `-invert=true`). For each issue, it will show the file path and line number.

Example output:
```
/path/to/file.ts:42 - Missing required code block
/path/to/another.ts:156 - Missing required code block
```

When using `-invert=true`:
```
/path/to/file.ts:42 - Contains forbidden code block
```

## How It Works

The analyzer uses Tree-sitter to parse TypeScript files and identify different types of functions:
- **Exported functions**: Functions that are explicitly exported from a module
- **Internal functions**: Functions that are defined but not exported
- **Callback functions**: Functions passed as arguments to other functions

It then checks if each function contains the specified code block. The tool is particularly useful for enforcing coding standards across large codebases.

## Ignoring Functions

You can add a special comment above any function to make the analyzer ignore it:

```typescript
// @ts-analyzer-ignore
export function functionToIgnore() {
    // This function will be ignored by the analyzer
    // even if it's missing the required code block
}
```

This is useful for:
- Legacy code that can't be immediately updated
- Functions that legitimately don't need the required code block
- Special cases where the standard pattern doesn't apply
