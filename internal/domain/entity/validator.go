package entity

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aimv/media-gallery/internal/pkg/apperror"
)

// DetectContentType читает первые 512 байт потока, определяет MIME-тип
// и возвращает его как MediaType, а также восстановленный поток,
// содержащий прочитанный заголовок и оставшиеся данные.
// Если тип не поддерживается, возвращает apperror.ErrInvalidInput.
func DetectContentType(src io.Reader) (MediaType, io.Reader, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", nil, fmt.Errorf("read header: %w", err)
	}
	head = head[:n]

	contentType := http.DetectContentType(head)
	mediaType := MediaType(contentType)
	if !mediaType.IsValid() {
		return "", nil, apperror.ErrInvalidInput
	}

	combined := io.MultiReader(bytes.NewReader(head), src)
	return mediaType, combined, nil
}
