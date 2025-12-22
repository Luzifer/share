package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
	"github.com/Luzifer/share/pkg/bootstrap"
	"github.com/Luzifer/share/pkg/uploader"
)

var (
	cfg = struct {
		BaseURL        string `flag:"base-url" default:"" description:"URL to prepend before filename"`
		Bootstrap      bool   `flag:"bootstrap" default:"false" description:"Upload frontend files into bucket"`
		Bucket         string `flag:"bucket" default:"" description:"S3 bucket to upload files to" validate:"nonzero"`
		ContentType    string `flag:"content-type,c" default:"" description:"Force content-type to be set to this value"`
		Endpoint       string `flag:"endpoint" default:"" description:"Override AWS S3 endpoint (i.e. to use MinIO)"`
		FileTemplate   string `flag:"file-template" vardefault:"file_template" description:"Full name template of the uploaded file"`
		Listen         string `flag:"listen" default:"" description:"Enable HTTP server if set to IP/Port (e.g. ':3000')"`
		LogLevel       string `flag:"log-level" default:"info" description:"Log level (debug, info, warn, error, fatal)"`
		Progress       bool   `flag:"progress" default:"false" description:"Show progress bar while uploading"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	uploaderOpts uploader.Opts

	version = "dev"
)

func initApp() (err error) {
	rconfig.AutoEnv(true)
	rconfig.SetVariableDefaults(map[string]string{
		"file_template": `file/{{ printf "%.6s" .Hash }}/{{ .SafeFileName }}`,
	})

	if err = rconfig.ParseAndValidate(&cfg); err != nil {
		return fmt.Errorf("parsing CLI options: %w", err)
	}

	l, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("parsing log-level: %w", err)
	}
	logrus.SetLevel(l)

	return nil
}

func main() {
	var err error
	if err = initApp(); err != nil {
		logrus.WithError(err).Fatal("initializing app")
	}

	if cfg.VersionAndExit {
		fmt.Printf("share %s\n", version) //nolint:forbidigo // Fine for version info
		os.Exit(0)
	}

	// Base-Options to clone from
	uploaderOpts = uploader.Opts{
		BaseURL:  cfg.BaseURL,
		Bucket:   cfg.Bucket,
		Endpoint: cfg.Endpoint,
	}

	if cfg.Progress {
		uploaderOpts = uploaderOpts.With(uploader.WithProgress())
	}

	switch {
	case cfg.Bootstrap:
		if err := bootstrap.Run(uploaderOpts); err != nil {
			logrus.WithError(err).Fatal("bootstrapping resources")
		}
		logrus.Info("Bucket bootstrap finished: Frontend uploaded")

	case cfg.Listen != "":
		logrus.WithFields(logrus.Fields{
			"addr":    cfg.Listen,
			"version": version,
		}).Info("share HTTP server started")
		if err := doListen(); err != nil {
			logrus.WithError(err).Fatal("running HTTP server")
		}

	default:
		if err := doCLIUpload(); err != nil {
			logrus.WithError(err).Fatal("uploading file")
		}
	}
}

func doCLIUpload() error {
	if len(rconfig.Args()) == 1 {
		return errors.New("missing argument: file to upload")
	}

	if cfg.BaseURL == "" {
		logrus.Warn("No BaseURL configured, output will be no complete URL")
	}

	var inFile io.ReadSeeker

	inFileName := rconfig.Args()[1]

	if inFileName == "-" {
		if cfg.ContentType == "" {
			// If we don't have an explicitly set content-type assume stdin contains text
			inFileName = "stdin"
			cfg.ContentType = "text/plain"
		} else if ext, err := mimeResolver.ExtensionsByType(cfg.ContentType); err == nil {
			inFileName = strings.Join([]string{"stdin", ext}, "")
		}

		// Stdin is not seekable, so we need to buffer it
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, os.Stdin); err != nil {
			logrus.WithError(err).Fatal("reading stdin")
		}

		inFile = bytes.NewReader(buf.Bytes())
	} else {
		inFileHandle, err := os.Open(inFileName) //#nosec:G304 // Inentional read of arbitrary file
		if err != nil {
			return errors.Wrap(err, "opening source file")
		}
		defer inFileHandle.Close() //nolint:errcheck // Irrelevant, file is closed by process exit
		inFile = inFileHandle
	}

	optsSetters := []uploader.OptsSetter{
		uploader.WithFile(inFileName, inFile),
		uploader.WithCalculatedFilename(cfg.FileTemplate),
	}

	if cfg.ContentType != "" {
		optsSetters = append(optsSetters, uploader.WithMimeType(cfg.ContentType))
	}

	url, err := uploader.Run(uploaderOpts.With(optsSetters...))
	if err != nil {
		return errors.Wrap(err, "uploading file")
	}

	fmt.Println(url) //nolint:forbidigo // Intended as programmatic payload
	return nil
}
