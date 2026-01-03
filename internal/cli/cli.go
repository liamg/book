package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/liamg/book/internal/bot"
)

type Config struct {
	Source          bot.Source
	Query           string
	LimitExtensions []string
	OutputFile      string
}

func Run(cfg Config) error {

	if cfg.Query == "" {
		return fmt.Errorf("invalid search query")
	}

	b := bot.New(cfg.Source)
	if err := b.Connect(); err != nil {
		return err
	}
	defer b.Close()

	fmt.Printf("Searching for \x1b[32m%s\x1b[0m...\n", cfg.Query)
	results, err := b.Search(cfg.Query, cfg.LimitExtensions)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("\nFound %d results:\n\n", len(results))
	for i, r := range results {
		fmt.Printf("  %d. %s\n", i+1, r.Filename)
	}

	fmt.Print("\nEnter number to download (or 'q' to quit): ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "q" || input == "" {
		return nil
	}

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(results) {
		return fmt.Errorf("invalid selection: %s", input)
	}

	selected := results[choice-1]

	fmt.Printf("\nDownloading \x1b[32m%s\x1b[0m...\n", selected.Filename)
	data, err := b.Download(selected)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}

	outFile := selected.Filename
	if cfg.OutputFile != "" {
		isDir := strings.HasSuffix(outFile, string(filepath.Separator))
		if !isDir {
			if stat, err := os.Stat(cfg.OutputFile); err == nil && stat.IsDir() {
				isDir = true
			}
		}
		if isDir {
			outFile = filepath.Join(cfg.OutputFile, filepath.Base(selected.Filename))
		} else {
			outFile = cfg.OutputFile
		}
	}
	fmt.Println("Download complete, writing to disk...")
	if err := os.WriteFile(outFile, data, 0600); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}
	fmt.Printf("Saved to \x1b[32m%s\x1b[0m\n", outFile)

	return nil
}
