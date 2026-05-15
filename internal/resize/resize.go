package resize

import (
	"context"
	"fmt"

	"tin/internal/magick"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

type Options struct {
	Input   string
	Output  string
	Width   int
	Height  int
	InPlace bool
}

type Result struct {
	Input      string
	Output     string
	BeforeSize int64
	AfterSize  int64
}

func Resize(ctx context.Context, opts Options) (Result, error) {
	if opts.Input == "" {
		return Result{}, fmt.Errorf("input file is required")
	}
	if opts.Width <= 0 {
		return Result{}, fmt.Errorf("width must be greater than 0")
	}
	if opts.Height < 0 {
		return Result{}, fmt.Errorf("height must be greater than or equal to 0")
	}
	if opts.InPlace && opts.Output != "" {
		return Result{}, fmt.Errorf("--in-place cannot be combined with --output")
	}

	ext := fileutil.LowerExt(opts.Input)
	if !fileutil.IsSupportedImageExt(ext) {
		return Result{}, fmt.Errorf("unsupported file type %q; supported types: .jpg, .jpeg, .png, .webp", ext)
	}

	before, err := fileutil.FileSize(opts.Input)
	if err != nil {
		return Result{}, err
	}

	run := func(target string) error {
		return magick.ResizeImage(ctx, opts.Input, target, opts.Width, opts.Height)
	}

	var output string
	if opts.InPlace {
		output = opts.Input
		if err := fsx.ReplaceInPlace(opts.Input, run); err != nil {
			return Result{}, err
		}
	} else {
		output = opts.Output
		if output == "" {
			output = DefaultOutputPath(opts.Input)
		}
		if err := run(output); err != nil {
			return Result{}, err
		}
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

func DefaultOutputPath(input string) string {
	return fileutil.DefaultOutputPath(input, "-resized")
}
