package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/branchard/jigs/internal/dotenv"
	"github.com/branchard/jigs/internal/prompt"
)

const outputPath = ".env"

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: jigs <file1> [file2 ...]\n")
		fmt.Fprintf(os.Stderr, "Example: jigs .env.dist .env.dev\n")
		os.Exit(1)
	}

	sourceFiles := os.Args[1:]

	// Collect all variables from source files in order, preserving first
	// occurrence and structure from the first file that defines each key.
	type varDef struct {
		key     string
		value   string // default value from source
		comment string // preceding comment block, if any
	}

	seen := make(map[string]bool)
	var orderedVars []varDef

	for _, path := range sourceFiles {
		src, err := dotenv.Parse(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		for _, entry := range src.Entries {
			if !entry.IsVar {
				continue
			}
			if seen[entry.Key] {
				continue
			}
			seen[entry.Key] = true
			orderedVars = append(orderedVars, varDef{
				key:   entry.Key,
				value: entry.Value,
			})
		}
	}

	if len(orderedVars) == 0 {
		fmt.Fprintln(os.Stderr, "No variables found in the provided files.")
		os.Exit(1)
	}

	// Load existing .env if it exists
	existing := make(map[string]string)
	if _, err := os.Stat(outputPath); err == nil {
		existingFile, err := dotenv.Parse(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading existing %s: %v\n", outputPath, err)
			os.Exit(1)
		}
		existing = existingFile.VarMap()
	}

	// Determine which variables need prompting
	var toPrompt []varDef
	for _, v := range orderedVars {
		if _, ok := existing[v.key]; ok {
			continue // already set in existing .env
		}
		toPrompt = append(toPrompt, v)
	}

	if len(toPrompt) == 0 {
		fmt.Println("All variables are already defined in", outputPath)
		return
	}

	// Prompt the user for missing values
	fmt.Printf("Please provide values for %d variable(s):\n\n", len(toPrompt))

	reader := bufio.NewReader(os.Stdin)
	results := make(map[string]string)
	for _, v := range toPrompt {
		val, err := prompt.ForValue(reader, os.Stdout, v.key, v.value)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nError reading input: %v\n", err)
			os.Exit(1)
		}
		results[v.key] = val
	}

	// Build the output file: start from existing .env or from scratch
	var output *dotenv.File
	if _, err := os.Stat(outputPath); err == nil {
		output, err = dotenv.Parse(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading existing %s: %v\n", outputPath, err)
			os.Exit(1)
		}
	} else {
		output = &dotenv.File{}
	}

	// Append new variables in order
	for _, v := range toPrompt {
		output.Set(v.key, results[v.key])
	}

	if err := output.Write(outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}

	fmt.Printf("\n%s has been updated with %d variable(s).\n", outputPath, len(toPrompt))
}
