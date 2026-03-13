package prompt

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ForValue prompts the user for a variable value via the given reader/writer.
// If defaultVal is non-empty, it is shown and used when the user presses Enter
// without typing anything.
// The caller must supply a *bufio.Reader and reuse it across calls so that
// buffered input is not lost between prompts (important when stdin is a pipe).
func ForValue(r *bufio.Reader, w io.Writer, key string, defaultVal string) (string, error) {
	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", key, defaultVal)
	} else {
		fmt.Fprintf(w, "%s: ", key)
	}

	line, err := r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	value := strings.TrimRight(line, "\r\n")
	if value == "" {
		return defaultVal, nil
	}
	return value, nil
}
