package split

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"tin/internal/ghostscript"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

type Options struct {
	Input     string
	OutputDir string
}

type Result struct {
	Input     string
	OutputDir string
	Files     []string
}

func Split(ctx context.Context, opts Options) (Result, error) {
	if opts.Input == "" {
		return Result{}, fmt.Errorf("input file is required")
	}
	if ext := fileutil.LowerExt(opts.Input); ext != ".pdf" {
		return Result{}, fmt.Errorf("unsupported file type %q; only .pdf is supported", ext)
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = DefaultOutputDir(opts.Input)
	}

	if err := ensureEmptyDir(outputDir); err != nil {
		return Result{}, err
	}

	pattern := filepath.Join(outputDir, "page-%03d.pdf")
	if err := ghostscript.SplitPDF(ctx, opts.Input, pattern); err != nil {
		return Result{}, err
	}

	files, err := listPages(outputDir)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Input:     opts.Input,
		OutputDir: outputDir,
		Files:     files,
	}, nil
}

func DefaultOutputDir(input string) string {
	return filepath.Join(fileutil.Parent(input), fileutil.Stem(input)+"-pages")
}

func ensureEmptyDir(dir string) error {
	exists, err := fsx.Exists(dir)
	if err != nil {
		return err
	}
	if !exists {
		return os.MkdirAll(dir, 0o755)
	}

	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output %q exists and is not a directory", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("output directory %q is not empty", dir)
	}
	return nil
}

func listPages(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}
