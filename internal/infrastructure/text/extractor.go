package text

import (
	"bytes"
	"io"

	"github.com/DanyKaystery/HSE-KPO-ANTIPLAGUE/internal/domain/shared"
)

type SimpleExtractor struct{}

func NewSimpleExtractor() *SimpleExtractor {
	return &SimpleExtractor{}
}

func (e *SimpleExtractor) ExtractText(content io.Reader, mimeType string) (string, error) {
	if mimeType == "text/plain" {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(content)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}

	return "", shared.ErrUnsupportedFormat
}
