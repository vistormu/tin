package convert

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"tin/internal/ffmpeg"
	"tin/internal/magick"
	"tin/internal/shared/fileutil"
	"tin/internal/shared/fsx"
)

type Family int

const (
	FamilyUnknown Family = iota
	FamilyImage
	FamilyPDF
	FamilyVideo
	FamilyAudio
)

var familyByExt = map[string]Family{
	".jpg":  FamilyImage,
	".jpeg": FamilyImage,
	".png":  FamilyImage,
	".webp": FamilyImage,
	".pdf":  FamilyPDF,
	".mp4":  FamilyVideo,
	".mov":  FamilyVideo,
	".mkv":  FamilyVideo,
	".webm": FamilyVideo,
	".avi":  FamilyVideo,
	".mp3":  FamilyAudio,
	".wav":  FamilyAudio,
	".m4a":  FamilyAudio,
	".flac": FamilyAudio,
	".ogg":  FamilyAudio,
	".opus": FamilyAudio,
}

type Options struct {
	Inputs []string
	Output string
	Target string
}

type Result struct {
	Inputs     []string
	Outputs    []string
	BeforeSize int64
	AfterSize  int64
}

func Convert(ctx context.Context, opts Options) ([]Result, error) {
	if len(opts.Inputs) == 0 {
		return nil, fmt.Errorf("input file is required")
	}

	toExt, err := resolveTargetExt(opts)
	if err != nil {
		return nil, err
	}
	toFam, ok := familyByExt[toExt]
	if !ok {
		return nil, fmt.Errorf("unsupported target type %q; supported types: %s", toExt, supportedExts())
	}

	fromFam := FamilyUnknown
	for _, input := range opts.Inputs {
		ext := fileutil.LowerExt(input)
		fam, ok := familyByExt[ext]
		if !ok {
			return nil, fmt.Errorf("unsupported input type %q; supported types: %s", ext, supportedExts())
		}
		if fromFam == FamilyUnknown {
			fromFam = fam
		} else if fam != fromFam {
			return nil, fmt.Errorf("all inputs must share the same family; mixed extensions found")
		}
	}

	if opts.Output != "" {
		outputExt := fileutil.LowerExt(opts.Output)
		if normalizeExt(outputExt) != normalizeExt(toExt) {
			return nil, fmt.Errorf("output extension %q does not match target type %q", outputExt, toExt)
		}
	}

	switch {
	case fromFam == FamilyImage && toFam == FamilyImage:
		return convertImageToImage(ctx, opts, toExt)
	case fromFam == FamilyImage && toFam == FamilyPDF:
		return convertImageToPDF(ctx, opts)
	case fromFam == FamilyPDF && toFam == FamilyImage:
		return convertPDFToImage(ctx, opts, toExt)
	case fromFam == FamilyVideo && toFam == FamilyVideo:
		return convertWithFFmpeg(ctx, opts, toExt, false)
	case fromFam == FamilyVideo && toFam == FamilyAudio:
		return convertWithFFmpeg(ctx, opts, toExt, true)
	case fromFam == FamilyAudio && toFam == FamilyAudio:
		return convertWithFFmpeg(ctx, opts, toExt, false)
	default:
		return nil, fmt.Errorf("unsupported conversion from %s to %s", familyName(fromFam), familyName(toFam))
	}
}

func convertImageToImage(ctx context.Context, opts Options, toExt string) ([]Result, error) {
	if len(opts.Inputs) > 1 && opts.Output != "" {
		return nil, fmt.Errorf("--output is only allowed with a single input")
	}

	results := make([]Result, 0, len(opts.Inputs))
	for _, input := range opts.Inputs {
		fromExt := fileutil.LowerExt(input)
		if normalizeExt(fromExt) == normalizeExt(toExt) {
			return nil, fmt.Errorf("input %q is already %s", input, toExt)
		}

		output := opts.Output
		if output == "" {
			output = fileutil.ReplaceExt(input, toExt)
		}
		if err := refuseOverwrite(output); err != nil {
			return nil, err
		}

		before, err := fileutil.FileSize(input)
		if err != nil {
			return nil, err
		}
		if err := magick.ConvertImage(ctx, input, output); err != nil {
			return nil, err
		}
		after, err := fileutil.FileSize(output)
		if err != nil {
			return nil, err
		}

		results = append(results, Result{
			Inputs:     []string{input},
			Outputs:    []string{output},
			BeforeSize: before,
			AfterSize:  after,
		})
	}
	return results, nil
}

func convertImageToPDF(ctx context.Context, opts Options) ([]Result, error) {
	if len(opts.Inputs) == 1 {
		input := opts.Inputs[0]
		output := opts.Output
		if output == "" {
			output = fileutil.ReplaceExt(input, ".pdf")
		}
		if err := refuseOverwrite(output); err != nil {
			return nil, err
		}

		before, err := fileutil.FileSize(input)
		if err != nil {
			return nil, err
		}
		if err := magick.ImagesToPDF(ctx, []string{input}, output); err != nil {
			return nil, err
		}
		after, err := fileutil.FileSize(output)
		if err != nil {
			return nil, err
		}
		return []Result{{
			Inputs:     []string{input},
			Outputs:    []string{output},
			BeforeSize: before,
			AfterSize:  after,
		}}, nil
	}

	output := opts.Output
	if output == "" {
		output = filepath.Join(fileutil.Parent(opts.Inputs[0]), "combined.pdf")
	}
	if err := refuseOverwrite(output); err != nil {
		return nil, err
	}

	var before int64
	for _, input := range opts.Inputs {
		size, err := fileutil.FileSize(input)
		if err != nil {
			return nil, err
		}
		before += size
	}

	if err := magick.ImagesToPDF(ctx, opts.Inputs, output); err != nil {
		return nil, err
	}
	after, err := fileutil.FileSize(output)
	if err != nil {
		return nil, err
	}

	return []Result{{
		Inputs:     append([]string{}, opts.Inputs...),
		Outputs:    []string{output},
		BeforeSize: before,
		AfterSize:  after,
	}}, nil
}

func convertPDFToImage(ctx context.Context, opts Options, toExt string) ([]Result, error) {
	if opts.Output != "" {
		return nil, fmt.Errorf("--output is not supported when converting pdf to images; pages are written next to each input")
	}

	results := make([]Result, 0, len(opts.Inputs))
	for _, input := range opts.Inputs {
		parent := fileutil.Parent(input)
		stem := fileutil.Stem(input)
		pattern := filepath.Join(parent, stem+"-%03d"+toExt)
		globPattern := filepath.Join(parent, stem+"-[0-9][0-9][0-9]"+toExt)

		existing, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, err
		}
		if len(existing) > 0 {
			return nil, fmt.Errorf("output files already exist matching %q", globPattern)
		}

		before, err := fileutil.FileSize(input)
		if err != nil {
			return nil, err
		}
		if err := magick.PDFToImages(ctx, input, pattern); err != nil {
			return nil, err
		}

		outputs, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, err
		}
		if len(outputs) == 0 {
			return nil, fmt.Errorf("no pages were produced for %q", input)
		}

		var after int64
		for _, p := range outputs {
			size, err := fileutil.FileSize(p)
			if err != nil {
				return nil, err
			}
			after += size
		}

		results = append(results, Result{
			Inputs:     []string{input},
			Outputs:    outputs,
			BeforeSize: before,
			AfterSize:  after,
		})
	}
	return results, nil
}

func convertWithFFmpeg(ctx context.Context, opts Options, toExt string, stripVideo bool) ([]Result, error) {
	if len(opts.Inputs) > 1 && opts.Output != "" {
		return nil, fmt.Errorf("--output is only allowed with a single input")
	}

	results := make([]Result, 0, len(opts.Inputs))
	for _, input := range opts.Inputs {
		fromExt := fileutil.LowerExt(input)
		if !stripVideo && normalizeExt(fromExt) == normalizeExt(toExt) {
			return nil, fmt.Errorf("input %q is already %s", input, toExt)
		}

		output := opts.Output
		if output == "" {
			output = fileutil.ReplaceExt(input, toExt)
		}
		if err := refuseOverwrite(output); err != nil {
			return nil, err
		}

		before, err := fileutil.FileSize(input)
		if err != nil {
			return nil, err
		}

		var runErr error
		if stripVideo {
			runErr = ffmpeg.StripVideo(ctx, input, output)
		} else {
			runErr = ffmpeg.Convert(ctx, input, output)
		}
		if runErr != nil {
			return nil, runErr
		}

		after, err := fileutil.FileSize(output)
		if err != nil {
			return nil, err
		}

		results = append(results, Result{
			Inputs:     []string{input},
			Outputs:    []string{output},
			BeforeSize: before,
			AfterSize:  after,
		})
	}
	return results, nil
}

func DefaultOutputPath(input string, targetExt string) string {
	return fileutil.ReplaceExt(input, targetExt)
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

func refuseOverwrite(path string) error {
	exists, err := fsx.Exists(path)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("output %q already exists", path)
	}
	return nil
}

func supportedExts() string {
	exts := make([]string, 0, len(familyByExt))
	for ext := range familyByExt {
		exts = append(exts, ext)
	}
	return strings.Join(sortedDistinct(exts), ", ")
}

func sortedDistinct(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func familyName(f Family) string {
	switch f {
	case FamilyImage:
		return "image"
	case FamilyPDF:
		return "pdf"
	case FamilyVideo:
		return "video"
	case FamilyAudio:
		return "audio"
	default:
		return "unknown"
	}
}
