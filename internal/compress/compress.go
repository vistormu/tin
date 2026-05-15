package compress

import (
	"context"
	"fmt"

	"tin/internal/cwebp"
	"tin/internal/ghostscript"
	"tin/internal/magick"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

type Options struct {
	Input     string
	Output    string
	Intensity string
	InPlace   bool
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
	if opts.InPlace && opts.Output != "" {
		return Result{}, fmt.Errorf("--in-place cannot be combined with --output")
	}

	before, err := fileutil.FileSize(opts.Input)
	if err != nil {
		return Result{}, err
	}

	ext := fileutil.LowerExt(opts.Input)
	run := func(target string) error {
		switch ext {
		case ".pdf":
			preset, err := pdfPresetForIntensity(opts.Intensity)
			if err != nil {
				return err
			}
			return ghostscript.CompressPDF(ctx, opts.Input, target, preset)
		case ".jpg", ".jpeg", ".png":
			quality, err := imageQualityForIntensity(opts.Intensity)
			if err != nil {
				return err
			}
			return magick.CompressImage(ctx, opts.Input, target, quality)
		case ".webp":
			quality, err := imageQualityForIntensity(opts.Intensity)
			if err != nil {
				return err
			}
			return cwebp.CompressWebP(ctx, opts.Input, target, quality)
		default:
			return fmt.Errorf("unsupported file type %q; supported types: .pdf, .jpg, .jpeg, .png, .webp", ext)
		}
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
	return fileutil.DefaultOutputPath(input, "-compressed")
}
