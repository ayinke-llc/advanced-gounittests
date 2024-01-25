package fixtures

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrettyPrintJSON(t *testing.T) {
	tt := []struct {
		name     string
		filePath string
		hasErr   bool
	}{
		{
			name:     "Invalid json",
			filePath: "testdata/invalid.json",
			hasErr:   true,
		},
		{
			name:     "valid json",
			filePath: "testdata/valid.json",
			hasErr:   false,
		},
	}

	for _, v := range tt {
		f, err := os.Open(v.filePath)
		require.NoError(t, err)

		defer func() {
			require.NoError(t, f.Close())
		}()

		b := new(bytes.Buffer)
		_, err = io.Copy(b, f)
		require.NoError(t, err)

		formattedJSON, err := PrettyPrintJSON(b.String())
		if v.hasErr {
			require.Error(t, err)
			continue
		}

		require.NoError(t, err)
		// Make sure the formatted item differs from what was passed to i
		require.NotEqual(t, b.String(), formattedJSON)
	}
}
