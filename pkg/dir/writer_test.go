package dir

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestToDir(t *testing.T) {
	var testCases = []struct {
		name                   string
		files                  Dir
		expectedOutputFilename string
	}{
		{
			name: "packing 2 files",
			files: Dir{"one.txt": []byte("first"),
				"two.txt":          []byte("second"),
				"three\\three.txt": []byte("third")},
			expectedOutputFilename: testDir + "data.wad",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			expectedOutput, err := os.ReadFile(test.expectedOutputFilename)
			if err != nil {
				t.Errorf("Couldn't load test datat: %s", err)
			}
			var buf bytes.Buffer
			test.files.ToWriter(&buf)
			if err != nil {
				t.Errorf("ToWriter failed: %s", err)
			}
			if !reflect.DeepEqual(buf.Bytes(), expectedOutput) {
				t.Errorf("Saved file differs from expected")
			}
		})
	}
}
