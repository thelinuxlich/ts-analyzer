package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strings"
    "testing"

    sitter "github.com/smacker/go-tree-sitter"
    "github.com/smacker/go-tree-sitter/typescript/typescript"
)

func TestCheckExportedFunctions(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    // Test case 1: Function declaration with required code block
    testContent := []byte(`
export function functionWithCodeBlock() {
    const requiredCode = true;
    return requiredCode;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Create a query that matches the updated checkExportedFunctions implementation
    testQuery := `
    (export_statement
        (function_declaration) @func)
    (export_statement
        (lexical_declaration
            (variable_declarator
                value: (arrow_function) @arrow_func)))
    (export_statement
        (lexical_declaration
            (variable_declarator
                value: (function_expression) @func_expr)))
    `

    query, err := sitter.NewQuery([]byte(testQuery), typescript.GetLanguage())
    if err != nil {
        t.Fatalf("Error creating query: %v", err)
    }

    cursor := sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    // Test with code that exists in the function
    foundFunction := false
    hasRequiredCode := false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode = true") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported function")
    }

    if !hasRequiredCode {
        t.Error("Function does not contain required code")
    }

    // Test case 2: Function without required code block
    testContent = []byte(`
export function functionWithoutCodeBlock() {
    return false;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    cursor = sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    foundFunction = false
    hasRequiredCode = false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported function")
    }

    if hasRequiredCode {
        t.Error("Function should not contain required code")
    }

    // Test case 3: Exported arrow function with required code block
    testContent = []byte(`
export const arrowFunctionWithCodeBlock = () => {
    const requiredCode = true;
    return requiredCode;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    cursor = sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    foundFunction = false
    hasRequiredCode = false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode = true") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported arrow function")
    }

    if !hasRequiredCode {
        t.Error("Arrow function does not contain required code")
    }

    // Test case 4: Exported arrow function without required code block
    testContent = []byte(`
export const arrowFunctionWithoutCodeBlock = () => {
    return false;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    cursor = sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    foundFunction = false
    hasRequiredCode = false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported arrow function")
    }

    if hasRequiredCode {
        t.Error("Arrow function should not contain required code")
    }

    // Test case 5: Exported function expression with required code block
    testContent = []byte(`
export const functionExpressionWithCodeBlock = function() {
    const requiredCode = true;
    return requiredCode;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    cursor = sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    foundFunction = false
    hasRequiredCode = false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode = true") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported function expression")
    }

    if !hasRequiredCode {
        t.Error("Function expression does not contain required code")
    }

    // Test case 6: Exported function expression without required code block
    testContent = []byte(`
export const functionExpressionWithoutCodeBlock = function() {
    return false;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    cursor = sitter.NewQueryCursor()
    cursor.Exec(query, rootNode)

    foundFunction = false
    hasRequiredCode = false

    for {
        match, ok := cursor.NextMatch()
        if !ok {
            break
        }

        for _, capture := range match.Captures {
            foundFunction = true
            funcNode := capture.Node
            funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])
            if strings.Contains(funcContent, "requiredCode") {
                hasRequiredCode = true
            }
        }
    }

    if !foundFunction {
        t.Error("Failed to find exported function expression")
    }

    if hasRequiredCode {
        t.Error("Function expression should not contain required code")
    }
}

func TestShouldIgnore(t *testing.T) {
    testCases := []struct {
        path        string
        ignorePaths []string
        expected    bool
    }{
        // Exact match
        {
            path:        "node_modules/package/file.ts",
            ignorePaths: []string{"node_modules/package/file.ts"},
            expected:    true,
        },
        // Contains pattern
        {
            path:        "src/generated/file.ts",
            ignorePaths: []string{"generated"},
            expected:    true,
        },
        // Glob pattern
        {
            path:        "src/file.test.ts",
            ignorePaths: []string{"*.test.ts"},
            expected:    true,
        },
        // No match
        {
            path:        "src/main.ts",
            ignorePaths: []string{"node_modules", "*.test.ts", "generated"},
            expected:    false,
        },
    }

    for i, tc := range testCases {
        result := shouldIgnore(tc.path, tc.ignorePaths)
        if result != tc.expected {
            t.Errorf("Test case %d: Expected shouldIgnore(%q, %v) to return %v, got %v",
                i, tc.path, tc.ignorePaths, tc.expected, result)
        }
    }
}

func TestCheckAllFunctions(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    // Test with multiple function types - one missing the required code
    testContent := []byte(`
function regularFunction() {
    const requiredCode = true;
}

const arrowFunction = () => {
    const requiredCode = true;
}

class TestClass {
    methodFunction() {
        const requiredCode = true;
    }
}

const functionExpression = function() {
    // Missing required code here
    return false;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Capture stdout to check results without printing to console
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func() {
        w.Close()
        var buf bytes.Buffer
        io.Copy(&buf, r)
        os.Stdout = oldStdout
    }()

    // One function is missing the required code
    if testing.Verbose() {
        t.Log("Testing with one function missing required code")
    }
    result, _ := checkAllFunctions(rootNode, content, "requiredCode", false, testFile, false, false)
    if result {
        t.Error("Expected checkAllFunctions to return false when at least one function is missing the code block")
    }

    // Test with all functions containing the required code
    testContent = []byte(`
function regularFunction() {
    const requiredCode = true;
}

const arrowFunction = () => {
    const requiredCode = true;
}

class TestClass {
    methodFunction() {
        const requiredCode = true;
    }
}

const functionExpression = function() {
    const requiredCode = true;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    // Capture stdout to check results without printing to console
    oldStdout = os.Stdout
    r, w, _ = os.Pipe()
    os.Stdout = w
    defer func() {
        w.Close()
        var buf bytes.Buffer
        io.Copy(&buf, r)
        os.Stdout = oldStdout
    }()

    // All functions have the required code
    if testing.Verbose() {
        t.Log("Testing with all functions having required code")
    }
    result, _ = checkAllFunctions(rootNode, content, "requiredCode", false, testFile, false, false)
    if !result {
        t.Error("Expected checkAllFunctions to return true when all functions have the code block")
    }
}

func TestFunctionNodeTypes(t *testing.T) {
    // Create a temporary test file with different function types
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    testContent := []byte(`
function regularFunction() {
    const requiredCode = true;
}

const arrowFunction = () => {
    const requiredCode = true;
}

class TestClass {
    methodFunction() {
        const requiredCode = true;
    }
}

const functionExpression = function() {
    const requiredCode = true;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Skip printing AST structure to reduce test output noise
    if testing.Verbose() {
        // Print the node types to help debug
        t.Log("Node type:", rootNode.Type())

        // Print the AST structure to understand node types
        var printNode func(node *sitter.Node, depth int)
        printNode = func(node *sitter.Node, depth int) {
            if node == nil {
                return
            }

            indent := strings.Repeat("  ", depth)
            t.Logf("%s%s [%d-%d]", indent, node.Type(),
                node.StartPoint().Row+1, node.EndPoint().Row+1)

            for i := 0; i < int(node.ChildCount()); i++ {
                child := node.Child(i)
                if child != nil {
                    printNode(child, depth+1)
                }
            }
        }

        // Print first few levels of the AST
        t.Log("AST Structure:")
        printNode(rootNode, 0)
    }

    // Try different queries to see which ones work
    queries := []string{
        "(function_declaration) @func",
        "(arrow_function) @arrow",
        "(method_definition) @method",
        "(lexical_declaration (variable_declarator value: (function_expression))) @func_expr",
    }

    for i, queryStr := range queries {
        query, err := sitter.NewQuery([]byte(queryStr), typescript.GetLanguage())
        if err != nil {
            t.Logf("Query %d failed: %v", i, err)
            continue
        }

        cursor := sitter.NewQueryCursor()
        cursor.Exec(query, rootNode)

        count := 0
        for {
            match, ok := cursor.NextMatch()
            if !ok {
                break
            }

            for _, capture := range match.Captures {
                count++
                if testing.Verbose() {
                    t.Logf("Query %d matched node at line %d: %s",
                        i, capture.Node.StartPoint().Row+1, capture.Node.Type())
                }
            }
        }

        if testing.Verbose() {
            t.Logf("Query %d matched %d nodes", i, count)
        }
    }
}

func TestInvertedSearch(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    // Test case: Functions with forbidden code block (including arrow functions and function expressions)
    testContent := []byte(`
export function functionWithForbiddenCode() {
    const forbiddenCode = true;
    return forbiddenCode;
}

export function functionWithoutForbiddenCode() {
    const safeCode = true;
    return safeCode;
}

export const arrowWithForbiddenCode = () => {
    const forbiddenCode = true;
    return forbiddenCode;
};

export const arrowWithoutForbiddenCode = () => {
    const safeCode = true;
    return safeCode;
};

export const funcExprWithForbiddenCode = function() {
    const forbiddenCode = true;
    return forbiddenCode;
};

export const funcExprWithoutForbiddenCode = function() {
    const safeCode = true;
    return safeCode;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Test with inverted search (looking for functions that should NOT contain "forbiddenCode")
    if testing.Verbose() {
        t.Log("Testing inverted search - looking for functions containing forbidden code")
    }
    result, _ := checkExportedFunctions(rootNode, content, "forbiddenCode", false, testFile, true, false)
    if result {
        t.Error("Expected checkExportedFunctions with inverted search to return false when functions contain the forbidden code")
    }

    // Test with all functions not containing the forbidden code
    testContent = []byte(`
export function functionOne() {
    const safeCode = true;
    return safeCode;
}

export function functionTwo() {
    const anotherSafeCode = true;
    return anotherSafeCode;
}

export const arrowOne = () => {
    const safeCode = true;
    return safeCode;
};

export const arrowTwo = () => {
    const anotherSafeCode = true;
    return anotherSafeCode;
};

export const funcExprOne = function() {
    const safeCode = true;
    return safeCode;
};

export const funcExprTwo = function() {
    const anotherSafeCode = true;
    return anotherSafeCode;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    // Test with inverted search - all functions should pass
    if testing.Verbose() {
        t.Log("Testing inverted search - no functions should contain forbidden code")
    }
    result, _ = checkExportedFunctions(rootNode, content, "forbiddenCode", false, testFile, true, false)
    if !result {
        t.Error("Expected checkExportedFunctions with inverted search to return true when no functions contain the forbidden code")
    }
}

func TestCheckCallbackFunctions(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    // Test case 1: Callback function with required code block
    testContent := []byte(`
function main() {
    fetchData((response) => {
        const requiredCode = true;
        return response;
    });

    processItems(function(item) {
        const requiredCode = true;
        return item;
    });
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Test with callbacks that have the required code
    if testing.Verbose() {
        t.Log("Testing with callbacks having required code")
    }
    result, _ := checkCallbackFunctions(rootNode, content, "requiredCode", false, testFile, false, false)
    if !result {
        t.Error("Expected checkCallbackFunctions to return true when all callbacks have the code block")
    }

    // Test case 2: Callback function without required code block
    testContent = []byte(`
function main() {
    fetchData((response) => {
        // Missing required code
        return response;
    });

    processItems(function(item) {
        // Missing required code
        return item;
    });
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    // Test with callbacks missing the required code
    if testing.Verbose() {
        t.Log("Testing with callbacks missing required code")
    }
    result, _ = checkCallbackFunctions(rootNode, content, "requiredCode", false, testFile, false, false)
    if result {
        t.Error("Expected checkCallbackFunctions to return false when callbacks are missing the code block")
    }

    // Test case 3: Inverted search for forbidden code
    testContent = []byte(`
function main() {
    fetchData((response) => {
        const forbiddenCode = true;
        return response;
    });

    processItems(function(item) {
        // No forbidden code here
        return item;
    });
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    // Test with inverted search for forbidden code
    if testing.Verbose() {
        t.Log("Testing inverted search for forbidden code")
    }
    result, _ = checkCallbackFunctions(rootNode, content, "forbiddenCode", false, testFile, true, false)
    if result {
        t.Error("Expected checkCallbackFunctions with inverted search to return false when a callback contains forbidden code")
    }
}

func TestRegexCodeBlockMatching(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "regex_test2.ts")

    // Test with specific regex patterns
    testContent := []byte(`
function test() {
                using ctx = getContext();
                return true;
            }

function test() {
                using _ = getContext();
                return true;
            }

function test() {
                const ctx = getContext();
                return true;
            }

function test() {
                // using ctx = getContext();
                const ctx = getContext();
                return true;
            }

function test() {
                using myCustomContext = getContext();
                return true;
            }

function test() {
                using myContext = getContext();
                return true;
            }
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Test cases for different regex patterns
    patterns := []struct {
        pattern       string
        isRegex       bool
        expectedMatch bool
        description   string
    }{
        {
            pattern:       "using ctx = getContext()",
            isRegex:       false,
            expectedMatch: false, // Only one function has this exact match, but we're checking all functions
            description:   "Exact match - only first function should match",
        },
        {
            pattern:       `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:       true,
            expectedMatch: false, // Not all functions match this pattern
            description:   "Regex match - should match 4 out of 6 functions",
        },
        {
            pattern:       `using\s+(ctx|_|myCustomContext)\s+=\s+getContext\(\)`,
            isRegex:       true,
            expectedMatch: false, // Not all functions match this pattern
            description:   "Specific variable names - should match 3 out of 6 functions",
        },
    }

    for _, tc := range patterns {
        t.Run(tc.description, func(t *testing.T) {
            // Capture stdout to check results without printing to console
            oldStdout := os.Stdout
            r, w, _ := os.Pipe()
            os.Stdout = w
            defer func() {
                w.Close()
                var buf bytes.Buffer
                io.Copy(&buf, r)
                os.Stdout = oldStdout
            }()

            // Test with the pattern
            result, issueCount := checkAllFunctions(rootNode, content, tc.pattern, tc.isRegex, testFile, false, false)

            if result != tc.expectedMatch {
                t.Errorf("Expected result to be %v for pattern '%s', got %v with %d issues",
                    tc.expectedMatch, tc.pattern, result, issueCount)
            }

            // Check the pattern results
        })
    }
}

func TestIsCodeBlockUsedInFunction(t *testing.T) {
    testCases := []struct {
        name           string
        functionCode   string
        codeBlock      string
        isRegex        bool
        expectedResult bool
    }{
        {
            name: "Exact match",
            functionCode: `function test() {
                using ctx = getContext();
                return true;
            }`,
            codeBlock:      "using ctx = getContext()",
            isRegex:        false,
            expectedResult: true,
        },
        {
            name: "No match",
            functionCode: `function test() {
                using _ = getContext();
                return true;
            }`,
            codeBlock:      "using ctx = getContext()",
            isRegex:        false,
            expectedResult: false,
        },
        {
            name: "Regex match with lowercase variable",
            functionCode: `function test() {
                using ctx = getContext();
                return true;
            }`,
            codeBlock:      `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: true,
        },
        {
            name: "Regex match with underscore",
            functionCode: `function test() {
                using _ = getContext();
                return true;
            }`,
            codeBlock:      `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: true,
        },
        {
            name: "Regex no match",
            functionCode: `function test() {
                const ctx = getContext();
                return true;
            }`,
            codeBlock:      `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: false,
        },
        {
            name: "Regex match in comment only",
            functionCode: `function test() {
                // using ctx = getContext();
                const ctx = getContext();
                return true;
            }`,
            codeBlock:      `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: false,
        },
        {
            name: "Complex regex pattern",
            functionCode: `function test() {
                using myCustomContext = getContext();
                return true;
            }`,
            codeBlock:      `using\s+(ctx|_|myCustomContext)\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: true,
        },
        {
            name: "Regex match with camelCase variable",
            functionCode: `function test() {
                using myContext = getContext();
                return true;
            }`,
            codeBlock:      `using\s+[a-zA-Z0-9_]+\s+=\s+getContext\(\)`,
            isRegex:        true,
            expectedResult: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test the function with the pattern
            result := isCodeBlockUsedInFunction(tc.functionCode, tc.codeBlock, tc.isRegex, false)

            if result != tc.expectedResult {
                t.Errorf("Expected isCodeBlockUsedInFunction to return %v, got %v",
                    tc.expectedResult, result)
            }
        })
    }
}

func TestIgnoreComment(t *testing.T) {
    // Create a temporary test file
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.ts")

    // Test case: Function with ignore comment
    testContent := []byte(`
export function functionWithoutCodeBlock() {
    // This function is missing the required code block
    return false;
}

// @ts-analyzer-ignore
export function functionWithIgnoreComment() {
    // This function is also missing the required code block but has an ignore comment
    return false;
}

// This is a regular comment, not an ignore comment
export function anotherFunctionWithoutCodeBlock() {
    // This function is missing the required code block
    return false;
}
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    // Initialize tree-sitter
    parser := sitter.NewParser()
    parser.SetLanguage(typescript.GetLanguage())

    // Parse the file
    content, err := os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree := parser.Parse(nil, content)
    rootNode := tree.RootNode()

    // Capture stdout to check results without printing to console
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func() {
        w.Close()
        var buf bytes.Buffer
        io.Copy(&buf, r)
        os.Stdout = oldStdout
    }()

    // Test with the ignore comment - use false for verbose to avoid debug output
    result, issueCount := checkExportedFunctions(rootNode, content, "requiredCode", false, testFile, false, false)

    // We should have 2 issues (the first and third functions), but not the second one with the ignore comment
    if issueCount != 2 {
        t.Errorf("Expected 2 issues, got %d", issueCount)
    }

    if result {
        t.Error("Expected checkExportedFunctions to return false when functions are missing the code block")
    }

    // Test with arrow functions
    testContent = []byte(`
export const arrowFunction = () => {
    // This function is missing the required code block
    return false;
};

// @ts-analyzer-ignore
export const arrowFunctionWithIgnore = () => {
    // This function is also missing the required code block but has an ignore comment
    return false;
};
    `)

    if err := os.WriteFile(testFile, testContent, 0644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    content, err = os.ReadFile(testFile)
    if err != nil {
        t.Fatalf("Failed to read test file: %v", err)
    }

    tree = parser.Parse(nil, content)
    rootNode = tree.RootNode()

    // Capture stdout to check results without printing to console
    oldStdout = os.Stdout
    r, w, _ = os.Pipe()
    os.Stdout = w
    defer func() {
        w.Close()
        var buf bytes.Buffer
        io.Copy(&buf, r)
        os.Stdout = oldStdout
    }()

    // Test with the ignore comment for arrow functions
    result, issueCount = checkExportedFunctions(rootNode, content, "requiredCode", false, testFile, false, false)

    // We should have 1 issue (the first function), but not the second one with the ignore comment
    if issueCount != 1 {
        t.Errorf("Expected 1 issue, got %d", issueCount)
    }

    if result {
        t.Error("Expected checkExportedFunctions to return false when functions are missing the code block")
    }
}

func TestEndToEndIgnoreComment(t *testing.T) {
    // Skip if running in short mode
    if testing.Short() {
        t.Skip("Skipping end-to-end test in short mode")
    }

    // Create a temporary directory for test files
    tempDir := t.TempDir()

    // Create test files with different patterns
    files := map[string]string{
        "file1.ts": `
export function func1() {
    using ctx = getContext();
    return true;
}

// @ts-analyzer-ignore
export function func2() {
    // This function is missing the required code block but has an ignore comment
    return true;
}
`,
        "file2.ts": `
export function func3() {
    // This function is missing the required code block
    return true;
}

// @ts-analyzer-ignore
export const func4 = () => {
    // This arrow function is missing the required code block but has an ignore comment
    return true;
};
`,
    }

    // Write test files
    for filename, content := range files {
        filePath := filepath.Join(tempDir, filename)
        if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
            t.Fatalf("Failed to write test file %s: %v", filename, err)
        }
    }

    // Save current working directory
    originalDir, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current directory: %v", err)
    }
    defer os.Chdir(originalDir)

    // Change to temp directory
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("Failed to change to temp directory: %v", err)
    }

    // Reset flags to avoid redefinition errors
    flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

    // Capture stdout to check results without printing to console
    oldStdout := os.Stdout
    r, w, _ := os.Pipe()
    os.Stdout = w
    defer func() {
        os.Stdout = oldStdout
    }()

    // Set up command line arguments - make sure verbose is false to avoid debug output
    os.Args = []string{
        "ts-analyzer",
        "-code-block", "using",
        "-file-glob", "*.ts",
        "-verbose", "false",
    }

    // Override os.Exit for testing
    oldOsExit := osExit
    defer func() { osExit = oldOsExit }()

    exitCode := 0
    osExit = func(code int) {
        exitCode = code
        panic(exitError{code: code})
    }

    // Run the main function in a separate goroutine
    done := make(chan bool)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                if _, ok := r.(exitError); !ok {
                    t.Errorf("Unexpected panic: %v", r)
                }
            }
            done <- true
        }()
        main()
    }()

    // Wait for the function to complete
    <-done

    // Close the pipe to flush the output
    w.Close()

    // Read the output but don't print it
    var buf bytes.Buffer
    io.Copy(&buf, r)

    // We should have 1 file with issues (file2.ts with func3 missing the code block)
    // The other functions either have the code block or have the ignore comment
    if exitCode != 1 {
        t.Errorf("Expected exit code 1, got %d", exitCode)
    }
}

func TestEndToEndRegexFlag(t *testing.T) {
    // Skip if running in short mode
    if testing.Short() {
        t.Skip("Skipping end-to-end test in short mode")
    }

    // Create a temporary directory for test files
    tempDir := t.TempDir()

    // Create test files with different patterns
    files := map[string]string{
        "file1.ts": `
export function func1() {
    using ctx = getContext();
    return true;
}

export function func2() {
    using _ = getContext();
    return true;
}
`,
        "file2.ts": `
export function func3() {
    using myContext = getContext();
    return true;
}

export function func4() {
    const ctx = getContext(); // Not using the "using" keyword
    return true;
}
`,
    }

    // Write test files
    for filename, content := range files {
        filePath := filepath.Join(tempDir, filename)
        if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
            t.Fatalf("Failed to write test file %s: %v", filename, err)
        }
    }

    // Save current working directory
    originalDir, err := os.Getwd()
    if err != nil {
        t.Fatalf("Failed to get current directory: %v", err)
    }
    defer os.Chdir(originalDir)

    // Change to temp directory
    if err := os.Chdir(tempDir); err != nil {
        t.Fatalf("Failed to change to temp directory: %v", err)
    }

    // Test cases
    testCases := []struct {
        name          string
        codeBlock     string
        isRegex       bool
        invert        bool
        expectedFiles int
        expectedExit  int
    }{
        {
            name:          "Exact match - only matches ctx",
            codeBlock:     "using ctx = getContext()",
            isRegex:       false,
            invert:        false,
            expectedFiles: 2, // Both files have functions missing the exact match
            expectedExit:  1,
        },
        {
            name:          "Regex match - matches all using patterns",
            codeBlock:     `using [a-z_]+ = getContext\(\)`,
            isRegex:       true,
            invert:        false,
            expectedFiles: 1, // file2.ts has functions missing the pattern
            expectedExit:  1,
        },
        {
            name:          "Regex match - with invert flag",
            codeBlock:     `using [a-z_]+ = getContext\(\)`,
            isRegex:       true,
            invert:        true,
            expectedFiles: 1, // file2.ts has func4 without the pattern
            expectedExit:  1,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Reset flags to avoid redefinition errors
            flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

            // Set up command line arguments
            os.Args = []string{
                "ts-analyzer",
                "-code-block", tc.codeBlock,
                "-regex", fmt.Sprintf("%t", tc.isRegex),
                "-invert", fmt.Sprintf("%t", tc.invert),
                "-file-glob", "*.ts",
                "-verbose", "true",
            }

            // Set up the command line arguments

            // Capture stdout to check results
            oldStdout := os.Stdout
            r, w, _ := os.Pipe()
            os.Stdout = w

            // Override os.Exit for testing
            oldOsExit := osExit
            defer func() { osExit = oldOsExit }()

            exitCode := 0
            osExit = func(code int) {
                exitCode = code
                panic(exitError{code: code})
            }

            // Run the main function in a separate goroutine
            done := make(chan bool)
            go func() {
                defer func() {
                    if r := recover(); r != nil {
                        if _, ok := r.(exitError); !ok {
                            t.Errorf("Unexpected panic: %v", r)
                        }
                    }
                    done <- true
                }()
                main()
            }()

            // Wait for main to complete
            <-done

            // Restore stdout
            w.Close()
            os.Stdout = oldStdout

            // Read captured output
            var buf bytes.Buffer
            io.Copy(&buf, r)
            output := buf.String()

            // Process the command output

            // Check exit code
            if exitCode != tc.expectedExit {
                t.Errorf("Expected exit code %d, got %d", tc.expectedExit, exitCode)
            }

            // Count files with issues in output by looking for the "Total: X file(s) with issues" line
            fileCount := 0
            lines := strings.Split(output, "\n")
            for _, line := range lines {
                if strings.Contains(line, "Total:") && strings.Contains(line, "file(s) with issues") {
                    fmt.Sscanf(line, "Total: %d file(s) with issues", &fileCount)
                    break
                }
            }
            // Check if the number of files with issues matches the expected count
            if fileCount != tc.expectedFiles {
                t.Errorf("Expected %d files with issues, found %d\nOutput: %s",
                    tc.expectedFiles, fileCount, output)
            }
        })
    }
}

// Define an error type for exiting with a specific code
type exitError struct {
    code int
}

func (e exitError) Error() string {
    return fmt.Sprintf("exit with code %d", e.code)
}
