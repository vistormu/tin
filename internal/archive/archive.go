package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"tin/internal/shared/fileutil"
)

type ZipOptions struct {
	Inputs []string
	Output string
}

type ZipResult struct {
	Output  string
	Files   int
	Bytes   int64
	Skipped int
}

type UnzipOptions struct {
	Input     string
	OutputDir string
}

type UnzipResult struct {
	OutputDir string
	Files     int
	Bytes     int64
}

func Zip(opts ZipOptions) (ZipResult, error) {
	if len(opts.Inputs) == 0 {
		return ZipResult{}, fmt.Errorf("at least one input path is required")
	}

	output := opts.Output
	if output == "" {
		output = DefaultZipOutput(opts.Inputs)
	}

	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return ZipResult{}, fmt.Errorf("create %q: %w", output, err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	result := ZipResult{Output: output}
	seen := make(map[string]struct{})

	for _, input := range opts.Inputs {
		if err := addToZip(zw, input, seen, &result); err != nil {
			_ = zw.Close()
			_ = out.Close()
			_ = os.Remove(output)
			return ZipResult{}, err
		}
	}

	if err := zw.Close(); err != nil {
		_ = os.Remove(output)
		return ZipResult{}, err
	}
	return result, nil
}

func Unzip(opts UnzipOptions) (UnzipResult, error) {
	if opts.Input == "" {
		return UnzipResult{}, fmt.Errorf("archive path is required")
	}
	if strings.ToLower(filepath.Ext(opts.Input)) != ".zip" {
		return UnzipResult{}, fmt.Errorf("unsupported archive type %q; supported type: .zip", filepath.Ext(opts.Input))
	}

	outputDir := opts.OutputDir
	if outputDir == "" {
		outputDir = DefaultUnzipOutput(opts.Input)
	}

	reader, err := zip.OpenReader(opts.Input)
	if err != nil {
		return UnzipResult{}, fmt.Errorf("open %q: %w", opts.Input, err)
	}
	defer reader.Close()

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return UnzipResult{}, fmt.Errorf("create %q: %w", outputDir, err)
	}

	result := UnzipResult{OutputDir: outputDir}
	for _, file := range reader.File {
		written, err := extractZipFile(file, outputDir)
		if err != nil {
			return UnzipResult{}, err
		}
		if !file.FileInfo().IsDir() {
			result.Files++
			result.Bytes += written
		}
	}

	return result, nil
}

func DefaultZipOutput(inputs []string) string {
	if len(inputs) == 1 {
		clean := filepath.Clean(inputs[0])
		stem := fileutil.Stem(clean)
		if stem != "." && stem != string(filepath.Separator) && stem != "" {
			return filepath.Join(fileutil.Parent(clean), stem+".zip")
		}
	}
	return "archive.zip"
}

func DefaultUnzipOutput(input string) string {
	return filepath.Join(fileutil.Parent(filepath.Clean(input)), UnzipDirName(input))
}

// UnzipDirName is the directory name an archive extracts into: its base name
// without the .zip extension.
func UnzipDirName(input string) string {
	base := fileutil.Base(filepath.Clean(input))
	ext := fileutil.Ext(base)
	if strings.EqualFold(ext, ".zip") {
		base = strings.TrimSuffix(base, ext)
	}
	if base == "" || base == "." {
		return "archive"
	}
	return base
}

func addToZip(zw *zip.Writer, input string, seen map[string]struct{}, result *ZipResult) error {
	clean := filepath.Clean(input)
	info, err := os.Lstat(clean)
	if err != nil {
		return fmt.Errorf("stat %q: %w", input, err)
	}

	base := filepath.Base(clean)
	if info.IsDir() {
		return filepath.WalkDir(clean, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			rel, err := filepath.Rel(filepath.Dir(clean), path)
			if err != nil {
				return err
			}
			return addZipEntry(zw, path, rel, seen, result)
		})
	}

	return addZipEntry(zw, clean, base, seen, result)
}

func addZipEntry(zw *zip.Writer, path string, archiveName string, seen map[string]struct{}, result *ZipResult) error {
	archiveName = filepath.ToSlash(filepath.Clean(archiveName))
	if archiveName == "." || archiveName == "" {
		return nil
	}
	if _, ok := seen[archiveName]; ok {
		return fmt.Errorf("duplicate archive path %q", archiveName)
	}
	seen[archiveName] = struct{}{}

	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		result.Skipped++
		return nil
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = archiveName
	if info.IsDir() {
		header.Name += "/"
		_, err := zw.CreateHeader(header)
		return err
	}
	if !info.Mode().IsRegular() {
		result.Skipped++
		return nil
	}
	header.Method = zip.Deflate

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	n, err := io.Copy(writer, file)
	if err != nil {
		return err
	}
	result.Files++
	result.Bytes += n
	return nil
}

func extractZipFile(file *zip.File, outputDir string) (int64, error) {
	name := decodeEntryName(file.Name)
	target, err := safeExtractPath(outputDir, name)
	if err != nil {
		return 0, err
	}

	info := file.FileInfo()
	if info.IsDir() {
		// Directories are created writable rather than with their stored mode.
		// Zips written on Windows record the DOS read-only attribute as mode
		// 0555, and honouring that makes every entry inside the directory fail
		// to extract with "permission denied".
		return 0, os.MkdirAll(target, 0700)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return 0, fmt.Errorf("cannot unzip symlink %q", name)
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("cannot unzip unsupported file type %q", name)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
		return 0, err
	}

	src, err := file.Open()
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, info.Mode().Perm())
	if err != nil {
		return 0, fmt.Errorf("create %q: %w", target, err)
	}
	n, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	if copyErr != nil {
		return 0, copyErr
	}
	return n, closeErr
}

func safeExtractPath(outputDir string, name string) (string, error) {
	cleanName := filepath.Clean(filepath.FromSlash(name))
	if filepath.IsAbs(cleanName) || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}

	target := filepath.Join(outputDir, cleanName)
	rel, err := filepath.Rel(outputDir, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path %q", name)
	}
	return target, nil
}
