package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/yogzblr/probe"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-probe <workflow.yaml>")
		os.Exit(1)
	}

	workflowFile := os.Args[1]

	// Read workflow file
	data, err := os.ReadFile(workflowFile)
	if err != nil {
		log.Fatalf("Failed to read workflow file: %v", err)
	}

	fmt.Printf("=== Testing Workflow: %s ===\n\n", filepath.Base(workflowFile))

	// Create probe instance
	p := probe.New()

	// Execute workflow
	result, err := p.ExecuteYAML(context.Background(), data)

	// Print results
	fmt.Printf("Workflow: %s\n", result.Name)
	fmt.Printf("Overall Success: %v\n\n", result.Success)

	for i, task := range result.Tasks {
		fmt.Printf("Task %d: %s (%s)\n", i+1, task.Name, task.Type)
		if task.Success {
			fmt.Printf("  ✓ SUCCESS\n")
		} else {
			fmt.Printf("  ✗ FAILED\n")
			if task.Error != "" {
				fmt.Printf("  Error: %s\n", task.Error)
			}
		}
		if task.Output != nil {
			fmt.Printf("  Output: %+v\n", task.Output)
		}
		fmt.Println()
	}

	if err != nil {
		fmt.Printf("\nWorkflow Error: %v\n", err)
		os.Exit(1)
	}

	if !result.Success {
		os.Exit(1)
	}

	fmt.Println("=== All tasks completed successfully ===")
}
