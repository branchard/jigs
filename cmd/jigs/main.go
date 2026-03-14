package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	"github.com/branchard/jigs/internal/dotenv"
	"github.com/branchard/jigs/internal/prompt"
)

// version is set at build time via -ldflags.
var version = "dev"

func printUsage(w *os.File) {
	fmt.Fprintf(w, "jigs - interactively populate Dotenv files from templates\n\n")
	fmt.Fprintf(w, "Usage:\n")
	fmt.Fprintf(w, "  jigs <file1> [file2 ...]\n\n")
	fmt.Fprintf(w, "Arguments:\n")
	fmt.Fprintf(w, "  file1, file2, ...  Template files (e.g. .env.dist, .env.dev)\n\n")
	fmt.Fprintf(w, "Options:\n")
	fmt.Fprintf(w, "  -h, --help         Show this help message and exit\n")
	fmt.Fprintf(w, "  -v, --version      Show version and exit\n\n")
	fmt.Fprintf(w, "Example:\n")
	fmt.Fprintf(w, "  jigs .env.dist .env.dev\n")
}

func main() {
	outputPath, err := filepath.Abs(".env") // TODO: can be changed with parameter --output -o
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error determining output path: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		printUsage(os.Stderr)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-h", "--help":
		printUsage(os.Stdout)
		return
	case "-v", "--version":
		fmt.Println(version)
		return
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
	alreadyExists := false
	if _, err := os.Stat(outputPath); err == nil {
		alreadyExists = true
		fmt.Printf("Output file already exists in \"%s\".\n", outputPath)
		fmt.Println("Only the missing variables will be asked.")
		existingFile, err := dotenv.Parse(outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading existing %s: %v\n", outputPath, err)
			os.Exit(1)
		}
		existing = existingFile.VarMap()
	} else {
		fmt.Printf("Output file does not exist. Creating a new one in \"%s\".\n", outputPath)
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
		fmt.Printf("All variables are already defined in \"%s\".\n", outputPath)
		return
	}

	// Prompt the user for missing values
	fmt.Printf("Please provide values for %d variable(s) (press enter to keep the default value in [brackets]):\n\n", len(toPrompt))

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
			fmt.Fprintf(os.Stderr, "Error reading existing \"%s\": %v\n", outputPath, err)
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
		fmt.Fprintf(os.Stderr, "Error writing \"%s\": %v\n", outputPath, err)
		os.Exit(1)
	}

	if alreadyExists {
		fmt.Printf("\n%d variable(s) has been added to \"%s\".\n", len(toPrompt), outputPath)
	} else {
		fmt.Printf("\n%s has been created with %d variable(s).\n", outputPath, len(toPrompt))
	}
}
