package path

import (
	"fmt"
	"net/url"
	"strings"
)

func decodeFileURI(uri string) (string, error) {
	if !strings.HasPrefix(uri, "file://") {
		return "", fmt.Errorf("unsupported URI scheme: %s", uri)
	}
	rawPath := strings.TrimPrefix(uri, "file://")

	decoded, err := url.PathUnescape(rawPath)
	if err != nil {
		return "", fmt.Errorf("decoding URI path %q: %w", uri, err)
	}
	return decoded, nil
}
