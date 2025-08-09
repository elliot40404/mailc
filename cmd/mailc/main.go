package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/elliot40404/mailc/internal/generator"
	"github.com/elliot40404/mailc/internal/parser"
)

const VERSION = "DEBUG"

func printHelp() {
	fmt.Println(`mailc - Type-safe email templates

Usage:
  mailc [command] [flags]

Commands:
  generate   Parse HTML templates and generate Go code
  help       Show this help message
  version    Show the current mailc version

Flags (for generate command):
  -input     Directory containing HTML email templates (default: ./emails)
  -output    Directory to write generated Go code (default: ./internal/emails)

Examples:
  mailc generate -input ./emails -output ./internal/emails
  mailc version`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		printHelp()
		return

	case "version":
		fmt.Printf("mailc version %s\n", VERSION)
		return

	case "generate":
		fs := flag.NewFlagSet("generate", flag.ExitOnError)
		inputDir := fs.String("input", "./emails", "Directory containing HTML email templates")
		outputDir := fs.String("output", "./internal/emails", "Directory to write generated Go code")
		version := fs.String("version", VERSION, "Version string to embed in generated files")
		err := fs.Parse(os.Args[2:])
		if err != nil {
			log.Fatalf("Error parsing cli flags")
		}

		if _, err := os.Stat(*inputDir); os.IsNotExist(err) {
			log.Fatalf("Input directory does not exist: %s", *inputDir)
		}

		if err := os.MkdirAll(*outputDir, 0o755); err != nil {
			log.Fatalf("Failed to create output directory: %v", err)
		}

		files, err := filepath.Glob(filepath.Join(*inputDir, "*.html"))
		if err != nil {
			log.Fatalf("Failed to list template files: %v", err)
		}
		if len(files) == 0 {
			log.Fatalf("No .html files found in input directory: %s", *inputDir)
		}

		// Parse all templates
		var templates []*parser.ParsedTemplate
		for _, file := range files {
			pt, err := parser.ParseFile(file)
			if err != nil {
				log.Fatalf("Failed to parse %s: %v", file, err)
			}
			templates = append(templates, pt)
		}

		// Generate code
		if err := generator.GenerateCode(templates, *outputDir, *version); err != nil {
			log.Fatalf("Code generation failed: %v", err)
		}

		fmt.Printf("✅ Generated %d email templates into %s\n", len(templates), *outputDir)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}
