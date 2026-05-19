package merge

import (
	"context"
	"fmt"
	"path/filepath"

	"tin/internal/ghostscript"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

const defaultOutputName = "merged.pdf"

type Options struct {
	Inputs []string
	Output string
}

type Result struct {
	Inputs     []string
	Output     string
	BeforeSize int64
	AfterSize  int64
}

func Merge(ctx context.Context, opts Options) (Result, error) {
	if len(opts.Inputs) < 2 {
		return Result{}, fmt.Errorf("at least two input files are required")
	}

	for _, input := range opts.Inputs {
		if ext := fileutil.LowerExt(input); ext != ".pdf" {
			return Result{}, fmt.Errorf("unsupported file type %q; only .pdf is supported", ext)
		}
	}

	output := opts.Output
	if output == "" {
		output = filepath.Join(fileutil.Parent(opts.Inputs[0]), defaultOutputName)
	}
	if ext := fileutil.LowerExt(output); ext != ".pdf" {
		return Result{}, fmt.Errorf("output must have .pdf extension, got %q", ext)
	}

	var before int64
	for _, input := range opts.Inputs {
		size, err := fileutil.FileSize(input)
		if err != nil {
			return Result{}, err
		}
		before += size
	}

	exists, err := fsx.Exists(output)
	if err != nil {
		return Result{}, err
	}
	if exists {
		return Result{}, fmt.Errorf("output %q already exists", output)
	}

	if err := ghostscript.MergePDFs(ctx, opts.Inputs, output); err != nil {
		return Result{}, err
	}

	after, err := fileutil.FileSize(output)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Inputs:     opts.Inputs,
		Output:     output,
		BeforeSize: before,
		AfterSize:  after,
	}, nil
}
