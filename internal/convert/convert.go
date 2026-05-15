package convert

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"tin/internal/magick"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

type Options struct {
	Input  string
	Output string
	Target string
}

type Result struct {
	Input      string
	Output     string
	BeforeSize int64
	AfterSize  int64
}

func Convert(ctx context.Context, opts Options) (Result, error) {
	if opts.Input == "" {
		return Result{}, fmt.Errorf("input file is required")
	}

	fromExt := fileutil.LowerExt(opts.Input)
	if !fileutil.IsSupportedImageExt(fromExt) {
		return Result{}, fmt.Errorf("unsupported input type %q; supported types: .jpg, .jpeg, .png, .webp", fromExt)
	}

	toExt, err := resolveTargetExt(opts)
	if err != nil {
		return Result{}, err
	}
	if !fileutil.IsSupportedImageExt(toExt) {
		return Result{}, fmt.Errorf("unsupported target type %q; supported types: .jpg, .jpeg, .png, .webp", toExt)
	}
	if opts.Output != "" {
		outputExt := fileutil.LowerExt(opts.Output)
		if normalizeExt(outputExt) != normalizeExt(toExt) {
			return Result{}, fmt.Errorf("output extension %q does not match target type %q", outputExt, toExt)
		}
	}
	if normalizeExt(fromExt) == normalizeExt(toExt) {
		return Result{}, fmt.Errorf("input is already %s", toExt)
	}

	output := opts.Output
	if output == "" {
		output = DefaultOutputPath(opts.Input, toExt)
	}

	exists, err := fsx.Exists(output)
	if err != nil {
		return Result{}, err
	}
	if exists {
		return Result{}, fmt.Errorf("output %q already exists", output)
	}

	before, err := fileutil.FileSize(opts.Input)
	if err != nil {
		return Result{}, err
	}

	if err := magick.ConvertImage(ctx, opts.Input, output); err != nil {
		return Result{}, err
	}

	after, err := fileutil.FileSize(output)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Input:      opts.Input,
		Output:     output,
		BeforeSize: before,
		AfterSize:  after,
	}, nil
}

func DefaultOutputPath(input string, targetExt string) string {
	ext := filepath.Ext(input)
	stem := strings.TrimSuffix(input, ext)
	return stem + targetExt
}

func resolveTargetExt(opts Options) (string, error) {
	if opts.Target != "" {
		ext := opts.Target
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		return strings.ToLower(ext), nil
	}
	if opts.Output != "" {
		return fileutil.LowerExt(opts.Output), nil
	}
	return "", fmt.Errorf("target type is required; pass -t <ext> or -o <output>")
}

func normalizeExt(ext string) string {
	if ext == ".jpeg" {
		return ".jpg"
	}
	return ext
}
