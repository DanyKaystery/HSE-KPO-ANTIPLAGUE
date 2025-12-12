package plagiarism

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShingleDetector_Compare(t *testing.T) {
	detector := NewShingleDetector()
	detector.ShingleLen = 2

	tests := []struct {
		name     string
		text1    string
		text2    string
		expected float64
	}{
		{
			name:     "Identical texts",
			text1:    "Hello world this is a test",
			text2:    "Hello world this is a test",
			expected: 1.0,
		},
		{
			name:     "Completely different",
			text1:    "Hello world this is a test",
			text2:    "Completely unique content here",
			expected: 0.0,
		},
		{
			name:     "Partial match",
			text1:    "The quick brown fox jumps over the lazy dog",
			text2:    "The quick brown fox jumps over the active cat",
			expected: 0.5,
		},
		{
			name:     "Empty text",
			text1:    "",
			text2:    "Some text",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := detector.Compare(tt.text1, tt.text2)
			assert.NoError(t, err)

			if tt.expected == 1.0 || tt.expected == 0.0 {
				assert.Equal(t, tt.expected, score)
			} else {
				assert.Greater(t, score, 0.0)
				assert.Less(t, score, 1.0)
			}
		})
	}
}
