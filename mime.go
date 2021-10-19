package main

import (
	"mime"
	"strings"

	"github.com/pkg/errors"
)

type mimeDB map[string]string

// mimeResolver contains some well-known mime-types and falls back
// to mime package to resolve the extension if no internal override
// is known for the given mime-type
var mimeResolver = mimeDB{
	"text/plain": ".txt", // Detected as .asc when using mime package
}

func (m mimeDB) ExtensionsByType(t string) (string, error) {
	if v, ok := m[t]; ok {
		return v, nil
	}

	exts, err := mime.ExtensionsByType(t)
	if err != nil {
		return "", err
	}

	for _, ext := range exts {
		if !strings.HasPrefix(ext, ".") {
			continue
		}
		return ext, nil
	}

	return "", errors.New("no extension found")
}
