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
func ForValue(r io.Reader, w io.Writer, key string, defaultVal string) (string, error) {
	reader := bufio.NewReader(r)

	if defaultVal != "" {
		fmt.Fprintf(w, "%s [%s]: ", key, defaultVal)
	} else {
		fmt.Fprintf(w, "%s: ", key)
	}

	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}

	value := strings.TrimRight(line, "\r\n")
	if value == "" {
		return defaultVal, nil
	}
	return value, nil
}
