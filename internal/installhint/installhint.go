package installhint

import (
	"fmt"
	"os/exec"
)

type PackageNames map[string]string

var lookPath = exec.LookPath

type manager struct {
	executable string
	command    func(pkg string) string
}

var managers = []manager{
	{executable: "brew", command: func(pkg string) string { return "brew install " + pkg }},
	{executable: "apt", command: func(pkg string) string { return "sudo apt install " + pkg }},
	{executable: "dnf", command: func(pkg string) string { return "sudo dnf install " + pkg }},
	{executable: "pacman", command: func(pkg string) string { return "sudo pacman -S " + pkg }},
	{executable: "apk", command: func(pkg string) string { return "sudo apk add " + pkg }},
	{executable: "zypper", command: func(pkg string) string { return "sudo zypper install " + pkg }},
}

func MissingToolError(binary string, packages PackageNames) error {
	return fmt.Errorf("%s is not installed. %s", binary, InstallHint(packages))
}

func InstallHint(packages PackageNames) string {
	if packages == nil {
		packages = PackageNames{}
	}

	for _, manager := range managers {
		if _, err := lookPath(manager.executable); err == nil {
			return "Install it with: " + manager.command(packageName(manager.executable, packages))
		}
	}

	return "Install it with your package manager."
}

func packageName(manager string, packages PackageNames) string {
	if pkg := packages[manager]; pkg != "" {
		return pkg
	}
	if pkg := packages["default"]; pkg != "" {
		return pkg
	}
	return "the required package"
}
