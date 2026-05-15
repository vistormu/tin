package trash

import (
	"fmt"
	"runtime"
)

type Mover interface {
	Move(path string) error
}

func NewMover() (Mover, error) {
	return NewMoverForOS(runtime.GOOS)
}

func NewMoverForOS(goos string) (Mover, error) {
	switch goos {
	case "darwin":
		return newDarwinMover()
	case "linux":
		return newLinuxMover()
	default:
		return nil, fmt.Errorf("tin trash supports macOS and Linux only; current OS is %q", goos)
	}
}
