package main

import (
	"bufio"
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed cv.schema.json
var cvSchemaJSON []byte

var cvTemplate = CV{
	Personal:   Personal{Address: Address{}},
	Summary:    []string{},
	Private:    []string{},
	Experience: []Experience{},
	Education:  []Education{},
	Skills:     []SkillGroup{},
}

func runInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	var force bool
	fs.BoolVar(&force, "force", false, "overwrite existing files with default ones")
	fs.BoolVar(&force, "f", false, "overwrite existing files with default ones")

	cfg := configTemplate
	bindConfigFlags(fs, &cfg)

	if err := fs.Parse(args); err != nil {
		return err
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading current directory: %w", err)
	}

	if len(entries) > 0 {
		fmt.Printf("warning: current folder is not empty (%d existing entries)\n", len(entries))
		if !confirm("continue anyway?") {
			fmt.Println("aborted")
			return nil
		}
	}

	cvJSON, err := marshalJSONIndent(struct {
		Schema string `json:"$schema"`
		CV
	}{
		Schema: "./cv.schema.json",
		CV:     cvTemplate,
	})
	if err != nil {
		return fmt.Errorf("marshaling cv template: %w", err)
	}

	configJSON, err := marshalJSONIndent(struct {
		Schema string `json:"$schema"`
		Config
	}{
		Schema: "./config.schema.json",
		Config: cfg,
	})
	if err != nil {
		return fmt.Errorf("marshaling config template: %w", err)
	}

	files := []struct {
		name    string
		content []byte
		mode    os.FileMode
	}{
		{"cv.json", cvJSON, 0644},
		{"cv.schema.json", cvSchemaJSON, 0644},
		{"config.json", configJSON, 0644},
		{"config.schema.json", configSchemaJSON, 0644},
	}

	for _, f := range files {
		path := filepath.Join(".", f.name)
		if !force {
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("skipping %s (already exists)\n", f.name)
				continue
			}
		}
		if err := os.WriteFile(path, f.content, f.mode); err != nil {
			return fmt.Errorf("writing %s: %w", f.name, err)
		}
		fmt.Printf("created %s\n", f.name)
	}

	fmt.Println("project initialized")
	return nil
}

func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}
