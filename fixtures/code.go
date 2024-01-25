package fixtures

import (
	"bytes"
	"encoding/json"
)

func PrettyPrintJSON(str string) (string, error) {
	var b bytes.Buffer
	if err := json.Indent(&b, []byte(str), "", "    "); err != nil {
		return "", err
	}
	return b.String(), nil
}

// func HighlightCode(w io.Writer, code string) error {
// 	return quick.Highlight(w, code, "json", "terminal256", "monokai")
// }
