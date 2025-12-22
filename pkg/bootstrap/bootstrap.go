package bootstrap

import (
	"bytes"
	"embed"
	"fmt"
	"strings"

	"github.com/Luzifer/share/pkg/uploader"
)

//go:embed frontend/*
var frontend embed.FS

// Run executes the bootstrap by uploading the frontend assets into
// the bucket
func Run(opts uploader.Opts) error {
	files, err := frontend.ReadDir("frontend")
	if err != nil {
		return fmt.Errorf("listing embedded files: %w", err)
	}

	for _, asset := range files {
		content, err := frontend.ReadFile(strings.Join([]string{"frontend", asset.Name()}, "/"))
		if err != nil {
			return fmt.Errorf("reading baked asset: %w", err)
		}

		if _, err := uploader.Run(opts.With(
			uploader.WithFile(asset.Name(), bytes.NewReader(content)),
			uploader.WithGzip(),
		)); err != nil {
			return fmt.Errorf("uploading bootstrap asset %q: %w", asset, err)
		}
	}
	return nil
}
