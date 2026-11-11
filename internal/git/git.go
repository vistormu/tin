package git

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"tin/internal/installhint"
)

type Runner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, dir string, name string, args ...string) (string, error)
}

type execRunner struct{}

func (execRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (execRunner) Run(ctx context.Context, dir string, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

type Client struct {
	workDir string
	runner  Runner
}

func NewClient(workDir string) *Client {
	workDir = filepath.Clean(workDir)
	if resolved, err := filepath.EvalSymlinks(workDir); err == nil {
		workDir = resolved
	}
	return &Client{
		workDir: workDir,
		runner:  execRunner{},
	}
}

type Commit struct {
	Hash      string
	ShortHash string
	Subject   string
}

type UndoPlan struct {
	Root           string
	Commit         Commit
	TrackedChanges int
	UntrackedFiles int
}

func (c *Client) BuildUndoPlan(ctx context.Context) (UndoPlan, error) {
	root, err := c.repositoryRoot(ctx)
	if err != nil {
		return UndoPlan{}, err
	}

	if _, err := c.runGit(ctx, root, "rev-parse", "--verify", "HEAD"); err != nil {
		return UndoPlan{}, fmt.Errorf("repository has no commits to undo")
	}
	if _, err := c.runGit(ctx, root, "rev-parse", "--verify", "HEAD^"); err != nil {
		return UndoPlan{}, fmt.Errorf("cannot undo the root commit because it has no parent")
	}

	out, err := c.runGit(ctx, root, "show", "-s", "--format=%H%x00%h%x00%s", "HEAD")
	if err != nil {
		return UndoPlan{}, err
	}
	parts := strings.SplitN(out, "\x00", 3)
	if len(parts) != 3 {
		return UndoPlan{}, fmt.Errorf("unexpected git output while reading HEAD")
	}

	status, err := c.runGit(ctx, root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return UndoPlan{}, err
	}
	tracked, untracked := statusCounts(status)

	return UndoPlan{
		Root: root,
		Commit: Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Subject:   parts[2],
		},
		TrackedChanges: tracked,
		UntrackedFiles: untracked,
	}, nil
}

func (c *Client) Undo(ctx context.Context, plan UndoPlan, hard bool) error {
	head, err := c.runGit(ctx, plan.Root, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	if head != plan.Commit.Hash {
		return fmt.Errorf("HEAD changed after the undo plan was created; run the command again")
	}

	mode := "--soft"
	if hard {
		mode = "--hard"
	}
	_, err = c.runGit(ctx, plan.Root, "reset", mode, plan.Commit.Hash+"^")
	return err
}

type ObliteratePlan struct {
	Root            string
	Path            string
	AffectedCommits int
	Origin          string
}

func (c *Client) BuildObliteratePlan(ctx context.Context, input string) (ObliteratePlan, error) {
	root, err := c.repositoryRoot(ctx)
	if err != nil {
		return ObliteratePlan{}, err
	}
	if err := c.checkFilterRepo(); err != nil {
		return ObliteratePlan{}, err
	}

	status, err := c.runGit(ctx, root, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return ObliteratePlan{}, err
	}
	if strings.TrimSpace(status) != "" {
		return ObliteratePlan{}, fmt.Errorf("working tree must be clean before rewriting history")
	}

	path, err := c.repositoryPath(root, input)
	if err != nil {
		return ObliteratePlan{}, err
	}

	history, err := c.runGit(ctx, root, "log", "--all", "--format=%H", "--", path)
	if err != nil {
		return ObliteratePlan{}, err
	}
	commits := nonEmptyLines(history)
	if len(commits) == 0 {
		return ObliteratePlan{}, fmt.Errorf("%q does not appear in repository history", path)
	}

	origin := ""
	if out, runErr := c.runner.Run(ctx, root, "git", "remote", "get-url", "origin"); runErr == nil {
		origin = strings.TrimSpace(out)
	}

	return ObliteratePlan{
		Root:            root,
		Path:            path,
		AffectedCommits: len(commits),
		Origin:          origin,
	}, nil
}

func (c *Client) Obliterate(ctx context.Context, plan ObliteratePlan, force bool) (string, error) {
	args := []string{"filter-repo"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--path", plan.Path, "--invert-paths")

	out, err := c.runner.Run(ctx, plan.Root, "git", args...)
	if err != nil {
		return "", commandError("git filter-repo", out, err)
	}
	return strings.TrimSpace(out), nil
}

func (c *Client) repositoryRoot(ctx context.Context) (string, error) {
	if _, err := c.runner.LookPath("git"); err != nil {
		return "", installhint.MissingToolError("git", installhint.PackageNames{
			"default": "git",
		})
	}

	out, err := c.runGit(ctx, c.workDir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a git repository", c.workDir)
	}
	root, err := filepath.Abs(strings.TrimSpace(out))
	if err != nil {
		return "", err
	}
	return filepath.Clean(root), nil
}

func (c *Client) repositoryPath(root string, input string) (string, error) {
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	path := input
	if !filepath.IsAbs(path) {
		path = filepath.Join(c.workDir, path)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	relative, err := filepath.Rel(root, filepath.Clean(absolute))
	if err != nil {
		return "", err
	}
	if relative == "." {
		return "", fmt.Errorf("refusing to obliterate the entire repository")
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%q is outside repository %s", input, root)
	}
	if relative == ".git" || strings.HasPrefix(relative, ".git"+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to target git metadata")
	}
	return filepath.ToSlash(relative), nil
}

func (c *Client) checkFilterRepo() error {
	if _, err := c.runner.LookPath("git-filter-repo"); err != nil {
		return installhint.MissingToolError("git-filter-repo", installhint.PackageNames{
			"default": "git-filter-repo",
		})
	}
	return nil
}

func (c *Client) runGit(ctx context.Context, dir string, args ...string) (string, error) {
	out, err := c.runner.Run(ctx, dir, "git", args...)
	if err != nil {
		label := "git"
		if len(args) > 0 {
			label += " " + args[0]
		}
		return "", commandError(label, out, err)
	}
	return strings.TrimSpace(out), nil
}

func commandError(label string, output string, err error) error {
	output = strings.TrimSpace(output)
	if output != "" {
		return fmt.Errorf("%s failed: %s", label, output)
	}
	return fmt.Errorf("%s failed: %w", label, err)
}

func statusCounts(status string) (tracked int, untracked int) {
	for _, line := range nonEmptyLines(status) {
		if strings.HasPrefix(line, "?? ") {
			untracked++
		} else {
			tracked++
		}
	}
	return tracked, untracked
}

func nonEmptyLines(value string) []string {
	var lines []string
	for _, line := range strings.Split(value, "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
