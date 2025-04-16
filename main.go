package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
	"github.com/bmatcuk/doublestar/v4"
)

// For testing purposes
var osExit = os.Exit

func main() {
	// Parse command line arguments
	var (
		codeBlock string
		isRegex   bool
		invert    bool
		fileGlob  string
		directory string
		fnTypes   string
		verbose   bool
	)

	flag.StringVar(&codeBlock, "code-block", "", "Code block to check for")
	flag.BoolVar(&isRegex, "regex", false, "Treat code-block as a regular expression")
	flag.BoolVar(&invert, "invert", false, "Invert the check (find functions that DO have the code block)")
	flag.StringVar(&fileGlob, "file-glob", "**/*.ts", "File glob pattern to search")
	flag.StringVar(&directory, "dir", ".", "Directory to search in")
	flag.StringVar(&fnTypes, "fn-types", "exported", "Function types to check: 'exported', 'internal', 'callback', or comma-separated combination")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.Parse()

	// Validate function types
	fnTypesMap := parseFunctionTypes(fnTypes)
	if len(fnTypesMap) == 0 {
		fmt.Println("Error: Invalid function types. Use 'exported', 'internal', 'callback', or a comma-separated combination")
		flag.Usage()
		os.Exit(1)
	}

	if codeBlock == "" {
		fmt.Println("Error: code-block is required")
		flag.Usage()
		os.Exit(1)
	}

	// Change to the specified directory
	if directory != "." {
		err := os.Chdir(directory)
		if err != nil {
			fmt.Printf("Error changing to directory %s: %v\n", directory, err)
			os.Exit(1)
		}
	}

	// Find all files matching the glob pattern
	files, err := findFiles(fileGlob)
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Printf("No files found matching pattern: %s\n", fileGlob)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("Found %d files to check\n", len(files))
	}

	allFilesValid := true
	invalidFiles := make(map[string]int) // Track files with issues and count of issues

	for _, file := range files {
		// Skip node_modules
		if strings.Contains(file, "node_modules") {
			continue
		}

		// Process TypeScript files
		if strings.HasSuffix(file, ".ts") || strings.HasSuffix(file, ".tsx") {
			// Get absolute path
			absPath, err := filepath.Abs(file)
			if err != nil {
				absPath = file // Fallback to original path
			}

			if verbose {
				fmt.Printf("Checking file: %s\n", absPath)
			}

			valid, issueCount := processTypeScriptFile(file, codeBlock, isRegex, invert, fnTypesMap, verbose)
			if !valid {
				allFilesValid = false
				invalidFiles[absPath] = issueCount
			}
		}
	}

	// Print summary
	if !allFilesValid {
		fmt.Println("\nSummary of files with issues:")

		// Get sorted list of filepaths
		var sortedPaths []string
		for filePath := range invalidFiles {
			// Convert to absolute path if not already
			absPath, err := filepath.Abs(filePath)
			if err == nil {
				sortedPaths = append(sortedPaths, absPath)
			} else {
				sortedPaths = append(sortedPaths, filePath)
			}
		}

		// Sort the filepaths alphabetically
		sort.Strings(sortedPaths)

		// Print issues in sorted order
		for _, absPath := range sortedPaths {
			count := invalidFiles[absPath]
			if invert {
				fmt.Printf("%s: %d function(s) containing forbidden code block\n", absPath, count)
			} else {
				fmt.Printf("%s: %d function(s) missing required code block\n", absPath, count)
			}
		}

		fmt.Printf("\nTotal: %d file(s) with issues\n", len(invalidFiles))
		osExit(1) // Use the variable instead of direct call
	} else if verbose {
		fmt.Println("All functions contain the required code block")
	}
}

// findFiles finds all files matching the given pattern
func findFiles(pattern string) ([]string, error) {
	var files []string

	// If the pattern is an absolute path, use it directly
	if filepath.IsAbs(pattern) {
		matches, err := doublestar.Glob(os.DirFS("."), pattern[1:]) // Remove leading slash
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			files = append(files, filepath.Join("/", match))
		}
		return files, nil
	}

	// For relative paths, use Walk
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			matched, err := doublestar.Match(pattern, path)
			if err != nil {
				return err
			}

			if matched {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}

// shouldIgnore checks if a path should be ignored based on the ignore list
func shouldIgnore(path string, ignorePaths []string) bool {
	for _, ignorePath := range ignorePaths {
		// Check for exact match
		if path == ignorePath {
			return true
		}

		// Check if path contains the ignore pattern
		if strings.Contains(path, ignorePath) {
			return true
		}

		// Check if path matches a glob pattern
		matched, err := filepath.Match(ignorePath, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// Fix the checkExportedFunctions function to properly handle inverted search
func checkExportedFunctions(rootNode *sitter.Node, content []byte, codeBlock string, isRegex bool, filePath string, invertSearch bool, verbose bool) (bool, int) {
	// Create query to find exported functions, including arrow functions and function expressions
	queryStr := `
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

	query, err := sitter.NewQuery([]byte(queryStr), typescript.GetLanguage())
	if err != nil {
		fmt.Printf("Error creating query: %v\n", err)
		return false, 0
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, rootNode)

	allPass := true
	issueCount := 0
	totalFunctions := 0

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			totalFunctions++
			funcNode := capture.Node

			// Check if the function has an ignore comment
			if hasIgnoreComment(content, funcNode) {
				if verbose {
					fmt.Printf("%s:%d - Skipping function due to @ts-analyzer-ignore comment\n",
						filePath, funcNode.StartPoint().Row+1)
				}
				continue
			}

			funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])

			if verbose {
				fmt.Println("Checking function content:")
				fmt.Println(funcContent)
				fmt.Println("Looking for code block:", codeBlock)
				if isRegex {
					fmt.Println("Using regex matching")
				}
			}

			hasCodeBlock := false

			if isRegex {
				re, err := regexp.Compile(codeBlock)
				if err != nil {
					fmt.Printf("Error compiling regex: %v\n", err)
					continue
				}

				// Check each line for the regex pattern
				lines := strings.Split(funcContent, "\n")
				for _, line := range lines {
					// Skip comment lines
					trimmedLine := strings.TrimSpace(line)
					if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") {
						continue
					}

					if re.MatchString(line) {
						hasCodeBlock = true
						break
					}
				}
			} else {
				// Check each line for the exact code block
				lines := strings.Split(funcContent, "\n")
				for _, line := range lines {
					// Skip comment lines
					trimmedLine := strings.TrimSpace(line)
					if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "/*") {
						continue
					}

					if strings.Contains(line, codeBlock) {
						hasCodeBlock = true
						break
					}
				}
			}

			// For inverted search, we're looking for functions that DON'T have the code block
			// So we invert the condition
			if invertSearch {
				if hasCodeBlock {
					allPass = false
					issueCount++
					fmt.Printf("%s:%d - Contains forbidden code block\n",
						filePath, funcNode.StartPoint().Row+1)
				}
			} else {
				if !hasCodeBlock {
					allPass = false
					issueCount++
					fmt.Printf("%s:%d - Missing required code block\n",
						filePath, funcNode.StartPoint().Row+1)
				}
			}
		}
	}

	return allPass, issueCount
}

func checkAllFunctions(node *sitter.Node, content []byte, codeBlock string, isRegex bool, filename string, invert bool, verbose bool) (bool, int) {
	if node == nil {
		fmt.Println("Error: nil node passed to checkAllFunctions")
		return false, 0
	}

	allFunctionsValid := true
	issueCount := 0

	// Query to find all functions
	queryStr := `
		(function_declaration) @func
		(arrow_function) @arrow
		(method_definition) @method
		(lexical_declaration
			(variable_declarator
				value: (function_expression))) @func_var
	`

	query, err := sitter.NewQuery([]byte(queryStr), typescript.GetLanguage())
	if err != nil {
		fmt.Printf("Error creating query: %v\n", err)
		return false, 0
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, node)

	foundAnyFunction := false
	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			foundAnyFunction = true
			funcNode := capture.Node

			// Check if the function has an ignore comment
			if hasIgnoreComment(content, funcNode) {
				if verbose {
					fmt.Printf("%s:%d - Skipping function due to @ts-analyzer-ignore comment\n",
						filename, funcNode.StartPoint().Row+1)
				}
				continue
			}

			funcContent := string(content[funcNode.StartByte():funcNode.EndByte()])

			// Check if the code block is properly used
			hasCodeBlock := isCodeBlockUsedInFunction(funcContent, codeBlock, isRegex, verbose)

			// If inverted, we want functions that DON'T have the code block
			// If not inverted, we want functions that DO have the code block
			if (!invert && !hasCodeBlock) || (invert && hasCodeBlock) {
				allFunctionsValid = false
				issueCount++
				if invert {
					fmt.Printf("%s:%d - Contains forbidden code block\n",
						filename, funcNode.StartPoint().Row+1)
				} else {
					fmt.Printf("%s:%d - Missing required code block\n",
						filename, funcNode.StartPoint().Row+1)
				}
			}
		}
	}

	// If no functions were found, return true (nothing to check)
	if !foundAnyFunction && verbose {
		fmt.Println("No functions found in the file")
	}

	return allFunctionsValid || !foundAnyFunction, issueCount
}

// isCodeBlockUsedInFunction checks if a code block is properly used within a function
func isCodeBlockUsedInFunction(funcContent string, codeBlock string, isRegex bool, verbose bool) bool {
	if verbose {
		fmt.Printf("Checking function content:\n%s\n", funcContent)
		fmt.Printf("Looking for code block: %s\n", codeBlock)
		if isRegex {
			fmt.Printf("Using regex matching\n")
		}
	}

	// If using regex, compile the pattern
	if isRegex {
		pattern, err := regexp.Compile(codeBlock)
		if err != nil {
			fmt.Printf("Error compiling regex pattern: %v\n", err)
			return false
		}

		// First, check if the code block exists at all using regex
		if !pattern.MatchString(funcContent) {
			if verbose {
				fmt.Printf("Regex pattern not found in function\n")
			}
			return false
		}

		// The code block exists, now check if it's in a comment
		lines := strings.Split(funcContent, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Skip empty lines
			if trimmedLine == "" {
				continue
			}

			// Check if line matches the pattern but is not a comment
			if pattern.MatchString(line) && !strings.HasPrefix(trimmedLine, "//") && !strings.HasPrefix(trimmedLine, "/*") {
				if verbose {
					fmt.Printf("Found regex match in non-comment line: %s\n", line)
				}
				return true
			}
		}
	} else {
		// Original string-based matching
		// First, check if the code block exists at all
		if !strings.Contains(funcContent, codeBlock) {
			if verbose {
				fmt.Printf("Code block not found in function\n")
			}
			return false
		}

		// The code block exists, now check if it's in a comment
		lines := strings.Split(funcContent, "\n")
		for _, line := range lines {
			trimmedLine := strings.TrimSpace(line)

			// Skip empty lines
			if trimmedLine == "" {
				continue
			}

			// Check if line contains the code block but is not a comment
			if strings.Contains(line, codeBlock) && !strings.HasPrefix(trimmedLine, "//") && !strings.HasPrefix(trimmedLine, "/*") {
				if verbose {
					fmt.Printf("Found code block in non-comment line: %s\n", line)
				}
				return true
			}
		}
	}

	if verbose {
		fmt.Printf("Code block only found in comments or not found at all\n")
	}
	return false
}

func processTypeScriptFile(filename string, codeBlock string, isRegex bool, invert bool, fnTypes map[string]bool, verbose bool) (bool, int) {
	// Get absolute path for consistent reporting
	absPath, err := filepath.Abs(filename)
	if err != nil {
		// If we can't get absolute path, use the original filename
		absPath = filename
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", absPath, err)
		return false, 0
	}

	// Parse the file with tree-sitter
	parser := sitter.NewParser()
	parser.SetLanguage(typescript.GetLanguage())

	tree := parser.Parse(nil, content)
	rootNode := tree.RootNode()

	allValid := true
	totalIssues := 0

	// Check each requested function type
	if fnTypes["exported"] {
		valid, issues := checkExportedFunctions(rootNode, content, codeBlock, isRegex, absPath, invert, verbose)
		if !valid {
			allValid = false
		}
		totalIssues += issues
	}

	if fnTypes["internal"] {
		valid, issues := checkInternalFunctions(rootNode, content, codeBlock, isRegex, absPath, invert, verbose)
		if !valid {
			allValid = false
		}
		totalIssues += issues
	}

	if fnTypes["callback"] {
		valid, issues := checkCallbackFunctions(rootNode, content, codeBlock, isRegex, absPath, invert, verbose)
		if !valid {
			allValid = false
		}
		totalIssues += issues
	}

	return allValid, totalIssues
}

// parseFunctionTypes parses the comma-separated function types string
func parseFunctionTypes(fnTypes string) map[string]bool {
	result := make(map[string]bool)
	types := strings.Split(fnTypes, ",")

	for _, t := range types {
		t = strings.TrimSpace(t)
		if t == "exported" || t == "internal" || t == "callback" {
			result[t] = true
		}
	}

	return result
}

// Add a new function to check internal (non-exported) functions
func checkInternalFunctions(node *sitter.Node, content []byte, codeBlock string, isRegex bool, filename string, invert bool, verbose bool) (bool, int) {
	if node == nil {
		fmt.Printf("Error: nil node passed to checkInternalFunctions for file %s\n", filename)
		return false, 0
	}

	allFunctionsValid := true
	issueCount := 0

	// Query to find non-exported functions
	queryStr := `
		(function_declaration) @func
		(method_definition) @method
		(lexical_declaration
			(variable_declarator
				name: (identifier) @var_name
				value: (function_expression) @func_expr))
		(lexical_declaration
			(variable_declarator
				name: (identifier) @var_name
				value: (arrow_function) @arrow_func))
	`

	query, err := sitter.NewQuery([]byte(queryStr), typescript.GetLanguage())
	if err != nil {
		fmt.Printf("Error creating query for file %s: %v\n", filename, err)
		return false, 0
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, node)

	// Track functions we've already checked to avoid duplicates
	checkedFunctions := make(map[string]bool)

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			// Skip variable names, only process function nodes
			if capture.Node.Type() == "identifier" {
				continue
			}

			funcNode := capture.Node
			startByte := funcNode.StartByte()

			// Create a unique key for this function
			funcKey := fmt.Sprintf("%d", startByte)

			// Skip if we've already checked this function or if it's an exported function
			if checkedFunctions[funcKey] || isExportedFunction(funcNode, node) {
				continue
			}
			checkedFunctions[funcKey] = true

			// Check if the function has an ignore comment
			if hasIgnoreComment(content, funcNode) {
				if verbose {
					fmt.Printf("%s:%d - Skipping function due to @ts-analyzer-ignore comment\n",
						filename, funcNode.StartPoint().Row+1)
				}
				continue
			}

			funcContent := string(content[startByte:funcNode.EndByte()])
			lineNum := funcNode.StartPoint().Row + 1

			// Check if the code block is properly used
			hasCodeBlock := isCodeBlockUsedInFunction(funcContent, codeBlock, isRegex, verbose)

			// If inverted, we want functions that DON'T have the code block
			// If not inverted, we want functions that DO have the code block
			if (!invert && !hasCodeBlock) || (invert && hasCodeBlock) {
				allFunctionsValid = false
				issueCount++
				if invert {
					fmt.Printf("%s:%d - Contains forbidden code block\n",
						filename, lineNum)
				} else {
					fmt.Printf("%s:%d - Missing required code block\n",
						filename, lineNum)
				}
			}
		}
	}

	return allFunctionsValid, issueCount
}

// Add a new function to check callback functions
func checkCallbackFunctions(node *sitter.Node, content []byte, codeBlock string, isRegex bool, filename string, invert bool, verbose bool) (bool, int) {
	if node == nil {
		fmt.Printf("Error: nil node passed to checkCallbackFunctions for file %s\n", filename)
		return false, 0
	}

	allFunctionsValid := true
	issueCount := 0

	// Query to find callback functions (functions passed as arguments)
	queryStr := `
		(call_expression
			arguments: (arguments
				(arrow_function) @callback_arrow))
		(call_expression
			arguments: (arguments
				(function_expression) @callback_func))
	`

	query, err := sitter.NewQuery([]byte(queryStr), typescript.GetLanguage())
	if err != nil {
		fmt.Printf("Error creating query for file %s: %v\n", filename, err)
		return false, 0
	}

	cursor := sitter.NewQueryCursor()
	cursor.Exec(query, node)

	// Track functions we've already checked to avoid duplicates
	checkedFunctions := make(map[string]bool)

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			funcNode := capture.Node
			startByte := funcNode.StartByte()

			// Create a unique key for this function
			funcKey := fmt.Sprintf("%d", startByte)

			// Skip if we've already checked this function
			if checkedFunctions[funcKey] {
				continue
			}
			checkedFunctions[funcKey] = true

			// Check if the function has an ignore comment
			if hasIgnoreComment(content, funcNode) {
				if verbose {
					fmt.Printf("%s:%d - Skipping function due to @ts-analyzer-ignore comment\n",
						filename, funcNode.StartPoint().Row+1)
				}
				continue
			}

			funcContent := string(content[startByte:funcNode.EndByte()])
			lineNum := funcNode.StartPoint().Row + 1

			// Check if the code block is properly used
			hasCodeBlock := isCodeBlockUsedInFunction(funcContent, codeBlock, isRegex, verbose)

			// If inverted, we want functions that DON'T have the code block
			// If not inverted, we want functions that DO have the code block
			if (!invert && !hasCodeBlock) || (invert && hasCodeBlock) {
				allFunctionsValid = false
				issueCount++
				if invert {
					fmt.Printf("%s:%d - Contains forbidden code block\n",
						filename, lineNum)
				} else {
					fmt.Printf("%s:%d - Missing required code block\n",
						filename, lineNum)
				}
			}
		}
	}

	return allFunctionsValid, issueCount
}

// Helper function to check if a function is exported
func isExportedFunction(funcNode *sitter.Node, rootNode *sitter.Node) bool {
	// Check if the function is directly exported
	parent := funcNode.Parent()
	if parent != nil && parent.Type() == "export_statement" {
		return true
	}

	// For variable declarations, we need to check if the variable is exported
	if funcNode.Type() == "function_expression" || funcNode.Type() == "arrow_function" {
		varDecl := funcNode.Parent()
		if varDecl != nil && varDecl.Type() == "variable_declarator" {
			lexDecl := varDecl.Parent()
			if lexDecl != nil && lexDecl.Parent() != nil && lexDecl.Parent().Type() == "export_statement" {
				return true
			}
		}
	}

	return false
}

// Helper function to check if a function has an ignore comment
func hasIgnoreComment(content []byte, funcNode *sitter.Node) bool {
	// Get the start line of the function
	startLine := funcNode.StartPoint().Row

	// If the function is at the first line, there can't be a comment above it
	if startLine == 0 {
		return false
	}

	// Get the content as string and split into lines
	lines := strings.Split(string(content), "\n")

	// Check the line above the function for the ignore comment
	prevLine := lines[startLine-1]
	return strings.Contains(prevLine, "// @ts-analyzer-ignore")
}
