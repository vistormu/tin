package dedup

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

type Group struct {
	Keep  string
	Trash []string
	Size  int64
}

type Plan struct {
	Groups []Group
}

func (p Plan) Removable() int {
	n := 0
	for _, g := range p.Groups {
		n += len(g.Trash)
	}
	return n
}

func (p Plan) RecoverableBytes() int64 {
	var b int64
	for _, g := range p.Groups {
		b += g.Size * int64(len(g.Trash))
	}
	return b
}

type Result struct {
	Trashed int
}

type Mover interface {
	Move(path string) error
}

func BuildPlan(root string) (Plan, error) {
	info, err := os.Stat(root)
	if err != nil {
		return Plan{}, err
	}
	if !info.IsDir() {
		return Plan{}, fmt.Errorf("%s is not a directory", root)
	}

	bySize := map[int64][]string{}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		entryInfo, err := d.Info()
		if err != nil {
			return err
		}
		size := entryInfo.Size()
		bySize[size] = append(bySize[size], path)
		return nil
	})
	if err != nil {
		return Plan{}, err
	}

	var groups []Group
	for size, files := range bySize {
		if len(files) < 2 {
			continue
		}
		byHash := map[string][]string{}
		for _, f := range files {
			h, err := hashFile(f)
			if err != nil {
				return Plan{}, err
			}
			byHash[h] = append(byHash[h], f)
		}
		for _, dups := range byHash {
			if len(dups) < 2 {
				continue
			}
			sort.Strings(dups)
			groups = append(groups, Group{
				Keep:  dups[0],
				Trash: dups[1:],
				Size:  size,
			})
		}
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Keep < groups[j].Keep
	})

	return Plan{Groups: groups}, nil
}

func Apply(plan Plan, mover Mover) (Result, error) {
	var result Result
	for _, g := range plan.Groups {
		for _, path := range g.Trash {
			if err := mover.Move(path); err != nil {
				return result, fmt.Errorf("trash %q: %w", path, err)
			}
			result.Trashed++
		}
	}
	return result, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
