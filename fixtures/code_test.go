package fixtures

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/sebdah/goldie/v2"
)

func verifyMatch(t *testing.T, v interface{}) {
	t.Helper()
	g := goldie.New(t, goldie.WithFixtureDir("./testdata/golden"))

	b := new(bytes.Buffer)

	err := json.NewEncoder(b).Encode(v)
	if err != nil {
		t.Fatal(err)
	}
	g.Assert(t, t.Name(), b.Bytes())
}

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
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := f.Close(); err != nil {
				t.Fatal(err)
			}
		}()

		b := new(bytes.Buffer)
		_, err = io.Copy(b, f)
		if err != nil {
			t.Fatal(err)
		}

		formattedJSON, err := PrettyPrintJSON(b.String())
		if v.hasErr {
			if err == nil {
				t.Fatal("Expected an error but got nil")
			}
			continue
		}

		if err != nil {
			t.Fatal(err)
		}

		verifyMatch(t, formattedJSON)
	}
}
