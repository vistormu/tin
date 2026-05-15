package compress

import (
	"context"
	"fmt"
	"strconv"

	"tin/internal/cwebp"
	"tin/internal/ghostscript"
	"tin/internal/magick"
	"tin/internal/shared/fileutil"
)

type Options struct {
	Input   string
	Output  string
	Quality string
}

type Result struct {
	Input      string
	Output     string
	BeforeSize int64
	AfterSize  int64
}

func (r Result) ReductionPercent() float64 {
	if r.BeforeSize <= 0 {
		return 0
	}
	return (float64(r.BeforeSize-r.AfterSize) / float64(r.BeforeSize)) * 100
}

func Compress(ctx context.Context, opts Options) (Result, error) {
	if opts.Input == "" {
		return Result{}, fmt.Errorf("input file is required")
	}

	before, err := fileutil.FileSize(opts.Input)
	if err != nil {
		return Result{}, err
	}

	output := opts.Output
	if output == "" {
		output = DefaultOutputPath(opts.Input)
	}

	ext := fileutil.LowerExt(opts.Input)
	switch ext {
	case ".pdf":
		err = ghostscript.CompressPDF(ctx, opts.Input, output, opts.Quality)
	case ".jpg", ".jpeg", ".png":
		if err := validateImageQuality(opts.Quality); err != nil {
			return Result{}, err
		}
		err = magick.CompressImage(ctx, opts.Input, output, opts.Quality)
	case ".webp":
		if err := validateImageQuality(opts.Quality); err != nil {
			return Result{}, err
		}
		err = cwebp.CompressWebP(ctx, opts.Input, output, opts.Quality)
	default:
		return Result{}, fmt.Errorf("unsupported file type %q; supported types: .pdf, .jpg, .jpeg, .png, .webp", ext)
	}
	if err != nil {
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

func DefaultOutputPath(input string) string {
	return fileutil.DefaultOutputPath(input, "-compressed")
}

func validateImageQuality(quality string) error {
	if quality == "" {
		return nil
	}

	value, err := strconv.Atoi(quality)
	if err != nil || value < 1 || value > 100 {
		return fmt.Errorf("unsupported image quality %q; use an integer from 1 to 100", quality)
	}
	return nil
}
