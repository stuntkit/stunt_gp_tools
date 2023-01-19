package dir

import (
	"testing"
)

func TestHashingFunction(t *testing.T) {
	// names should be provided without trailing \0, but the tests also check for that
	testCases := []struct {
		path         string
		expectedHash uint32
	}{
		{
			path:         "",
			expectedHash: 0,
		},
		{
			path:         "a",
			expectedHash: 97,
		},
		{
			path:         "test",
			expectedHash: 655,
		},
		{
			path:         "test\x00",
			expectedHash: 287,
		},
		{
			path:         "graphics24\\cars\\car15\\b15grid5.pc",
			expectedHash: 341,
		},
		{
			path:         "replays\\track19.rpl",
			expectedHash: 738,
		},
	}

	for _, testCase := range testCases {
		hash := GetHash(testCase.path)
		if hash != testCase.expectedHash {
			t.Fatalf("HashingFunction: path: %s, expected hash: %d, got %d", testCase.path, testCase.expectedHash, hash)
		}
	}
}
