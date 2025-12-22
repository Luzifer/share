package bootstrap

import (
	"bytes"
	"embed"
	"fmt"
	"strings"

	"github.com/Luzifer/share/pkg/uploader"
	"github.com/pkg/errors"
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
			return errors.Wrap(err, "reading baked asset")
		}

		if _, err := uploader.Run(opts.With(
			uploader.WithFile(asset.Name(), bytes.NewReader(content)),
			uploader.WithGzip(),
		)); err != nil {
			return errors.Wrapf(err, "uploading bootstrap asset %q", asset)
		}
	}
	return nil
}
