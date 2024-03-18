package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
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

	//go:embed frontend/*
	frontend embed.FS

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

	switch {
	case cfg.Bootstrap:
		if err := doBootstrap(); err != nil {
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

	url, err := executeUpload(inFileName, inFile, true, cfg.ContentType, false)
	if err != nil {
		return errors.Wrap(err, "uploading file")
	}

	fmt.Println(url) //nolint:forbidigo // Intended as programmatic payload
	return nil
}

func doBootstrap() error {
	for _, asset := range []string{"index.html", "app.js", "bundle.css", "bundle.js"} {
		content, err := frontend.ReadFile(strings.Join([]string{"frontend", asset}, "/"))
		if err != nil {
			return errors.Wrap(err, "reading baked asset")
		}

		if _, err := executeUpload(asset, bytes.NewReader(content), false, "", true); err != nil {
			return errors.Wrapf(err, "uploading bootstrap asset %q", asset)
		}
	}
	return nil
}
