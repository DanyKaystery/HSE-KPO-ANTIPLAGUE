package plagiarism

import (
	"strings"
)

type ShingleDetector struct {
	ShingleLen int
}

func NewShingleDetector() *ShingleDetector {
	return &ShingleDetector{ShingleLen: 3}
}

func (d *ShingleDetector) Compare(text1, text2 string) (float64, error) {
	if text1 == "" || text2 == "" {
		return 0.0, nil
	}

	set1 := d.getShingles(text1)
	set2 := d.getShingles(text2)

	intersection := 0
	for shingle := range set1 {
		if _, exists := set2[shingle]; exists {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0, nil
	}

	return float64(intersection) / float64(union), nil
}

func (d *ShingleDetector) getShingles(text string) map[string]struct{} {
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, ".", "")
	text = strings.ReplaceAll(text, ",", "")
	text = strings.ReplaceAll(text, "\n", " ")

	words := strings.Fields(text)
	shingles := make(map[string]struct{})

	if len(words) < d.ShingleLen {
		return shingles
	}

	for i := 0; i <= len(words)-d.ShingleLen; i++ {
		shingle := strings.Join(words[i:i+d.ShingleLen], " ")
		shingles[shingle] = struct{}{}
	}
	return shingles
}
