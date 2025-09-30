// Copyright (c) 2025 glassBead-tc and contributors
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const version = "1.0.0"

func main() {
	var (
		filePath  = flag.String("file", "", "Go file to migrate")
		dryRun    = flag.Bool("dry-run", false, "Show changes without modifying file")
		showHelp  = flag.Bool("help", false, "Show help")
		backupExt = flag.String("backup", ".backup", "Backup file extension")
	)

	flag.Parse()

	if *showHelp || *filePath == "" {
		printHelp()
		os.Exit(0)
	}

	// Check file exists
	if _, err := os.Stat(*filePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File not found: %s\n", *filePath)
		os.Exit(1)
	}

	// Read file
	content, err := os.ReadFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Check if it uses mark3labs
	if !strings.Contains(string(content), "github.com/mark3labs/mcp-go") {
		fmt.Println("✓ File doesn't use mark3labs SDK, no migration needed")
		os.Exit(0)
	}

	// Perform migration
	migrated, changes := migrateContent(string(content))

	if *dryRun {
		fmt.Println("🔍 Dry run - changes that would be made:")
		for i, change := range changes {
			fmt.Printf("%d. %s\n", i+1, change)
		}
		fmt.Println("\n--- Migrated Content ---")
		fmt.Println(migrated)
		return
	}

	// Create backup
	backupPath := *filePath + *backupExt
	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating backup: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📦 Backup created: %s\n", backupPath)

	// Write migrated content
	if err := os.WriteFile(*filePath, []byte(migrated), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing migrated file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Successfully migrated %s\n", *filePath)
	fmt.Printf("   %d changes applied\n", len(changes))
	for i, change := range changes {
		fmt.Printf("   %d. %s\n", i+1, change)
	}
}

func migrateContent(content string) (string, []string) {
	var changes []string
	migrated := content

	// 1. Update imports
	if strings.Contains(migrated, `"github.com/mark3labs/mcp-go/mcp"`) {
		migrated = strings.ReplaceAll(migrated, `"github.com/mark3labs/mcp-go/mcp"`, `"github.com/modelcontextprotocol/go-sdk/mcp"`)
		changes = append(changes, "Updated mcp import to official SDK")
	}

	if strings.Contains(migrated, `"github.com/mark3labs/mcp-go/server"`) {
		// Remove server import, it's not needed with official SDK
		migrated = regexp.MustCompile(`\s*"github.com/mark3labs/mcp-go/server"\n`).ReplaceAllString(migrated, "")
		changes = append(changes, "Removed mark3labs server import")
	}

	// 2. Update server creation
	serverPattern := regexp.MustCompile(`server\.NewMCPServer\(\s*"([^"]+)",\s*"([^"]+)"[^)]*\)`)
	if serverPattern.MatchString(migrated) {
		migrated = serverPattern.ReplaceAllStringFunc(migrated, func(match string) string {
			matches := serverPattern.FindStringSubmatch(match)
			if len(matches) >= 3 {
				name, version := matches[1], matches[2]
				return fmt.Sprintf(`mcp.NewServer(&mcp.Implementation{
		Name:    "%s",
		Version: "%s",
	}, nil)`, name, version)
			}
			return match
		})
		changes = append(changes, "Updated server creation to official SDK pattern")
	}

	// 3. Update tool result constructors
	if strings.Contains(migrated, "mcp.NewToolResultText(") {
		migrated = regexp.MustCompile(`mcp\.NewToolResultText\(([^)]+)\)`).ReplaceAllStringFunc(migrated, func(match string) string {
			// Extract the text argument
			textArg := strings.TrimPrefix(match, "mcp.NewToolResultText(")
			textArg = strings.TrimSuffix(textArg, ")")
			return fmt.Sprintf(`&mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: %s},
			},
		}`, textArg)
		})
		changes = append(changes, "Updated NewToolResultText calls")
	}

	if strings.Contains(migrated, "mcp.NewToolResultError(") {
		migrated = regexp.MustCompile(`mcp\.NewToolResultError\(([^)]+)\)`).ReplaceAllStringFunc(migrated, func(match string) string {
			textArg := strings.TrimPrefix(match, "mcp.NewToolResultError(")
			textArg = strings.TrimSuffix(textArg, ")")
			return fmt.Sprintf(`&mcp.CallToolResult{
			IsError: true,
			Content: []mcp.Content{
				&mcp.TextContent{Text: %s},
			},
		}`, textArg)
		})
		changes = append(changes, "Updated NewToolResultError calls")
	}

	// 4. Update ServeStdio
	if strings.Contains(migrated, "server.ServeStdio(") {
		migrated = regexp.MustCompile(`(?:if err := )?server\.ServeStdio\(([^)]+)\)(?:\s*;\s*err != nil \{[^}]+\})?`).ReplaceAllString(
			migrated,
			`if err := $1.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}`,
		)
		changes = append(changes, "Updated ServeStdio to Run with StdioTransport")
	}

	// 5. Add context import if not present
	if len(changes) > 0 && !strings.Contains(migrated, `"context"`) {
		// Find the import block and add context
		importPattern := regexp.MustCompile(`import \(\n`)
		if importPattern.MatchString(migrated) {
			migrated = importPattern.ReplaceAllString(migrated, `import (
	"context"
`)
			changes = append(changes, "Added context import")
		}
	}

	if len(changes) == 0 {
		changes = append(changes, "No automatic migrations needed - manual review required")
	}

	return migrated, changes
}

func printHelp() {
	fmt.Printf(`MCP SDK Migrator v%s
Converts Go MCP server code from mark3labs SDK to official modelcontextprotocol SDK

Usage:
  mcp-sdk-migrate -file <path> [options]

Options:
  -file string
        Go file to migrate (required)
  -dry-run
        Show changes without modifying file
  -backup string
        Backup file extension (default ".backup")
  -help
        Show this help message

Examples:
  # Migrate a file
  mcp-sdk-migrate -file cmd/my-server/main.go

  # Preview changes without modifying
  mcp-sdk-migrate -file server.go -dry-run

  # Custom backup extension
  mcp-sdk-migrate -file server.go -backup .old

What gets migrated:
  ✓ Import statements
  ✓ Server creation (NewMCPServer → NewServer)
  ✓ Tool result constructors (NewToolResultText, NewToolResultError)
  ✓ Server start (ServeStdio → Run with StdioTransport)

⚠️  Manual changes needed:
  - Tool registration (declarative → AddTool with typed args)
  - Handler signatures (add typed args parameter)
  - Parameter extraction (request.RequireString → parse from args struct)

For complete migration guide, see: docs/mcp-server-migration-spec.md
`, version)
}

