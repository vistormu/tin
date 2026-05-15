package confirm

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

func Ask(yes bool, prompt string, in io.Reader, out io.Writer) (bool, error) {
	if yes {
		return true, nil
	}

	fmt.Fprintf(out, "%s [y/N]: ", prompt)

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return false, err
		}
		return false, nil
	}

	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes", nil
}
