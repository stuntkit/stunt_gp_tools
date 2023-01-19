package dir

import (
	"reflect"
	"strings"
	"testing"
)

func TestFromDir(t *testing.T) {
	var testCases = []struct {
		name          string
		filename      string
		expectedFiles Dir
		shouldFail    bool
	}{
		{
			name:          "Empty dir file",
			filename:      testDir + "empty.wad",
			expectedFiles: Dir{},
			shouldFail:    false,
		},
		{
			name:     "Dir file with content",
			filename: testDir + "data.wad",
			expectedFiles: Dir{
				"one.txt":          []byte("first"),
				"two.txt":          []byte("second"),
				"three\\three.txt": []byte("third"),
			},
			shouldFail: false,
		},

		{
			name:          "Broken dir file",
			filename:      testDir + "broken.wad",
			expectedFiles: Dir{},
			shouldFail:    true,
		},
		{
			name:          "Not a dir file",
			filename:      testDir + "dummy.txt",
			expectedFiles: Dir{},
			shouldFail:    true,
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			result, err := FromDir(test.filename)
			if test.shouldFail {
				if err == nil {
					t.Error("FromDir didn't fail when it should")
				}
			} else {
				if err != nil {
					t.Errorf("FromDir failed when it shouldn't: %s", err)

				}
				if !reflect.DeepEqual(test.expectedFiles, result) {
					want := mapKeysToString(test.expectedFiles)
					got := mapKeysToString(result)
					t.Errorf("List of files is incorrect:\nwant: %s\ngot: %s", want, got)
				}
			}
		})
	}
}

func mapKeysToString(dir Dir) string {
	elements := make([]string, 0)
	for k, data := range dir {
		element := k + ": " + string(data)
		elements = append(elements, element)
	}
	return "[\n\t" + strings.Join(elements, ",\n\t") + "\n]"
}
