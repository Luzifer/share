package main

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		BaseURL        string `flag:"base-url" env:"BASE_URL" default:"" description:"URL to prepend before filename"`
		BasePath       string `flag:"base-path" env:"BASE_PATH" default:"file/{{ printf \"%.6s\" .Hash }}" description:"Path to upload the file to"`
		Bootstrap      bool   `flag:"bootstrap" default:"false" description:"Upload frontend files into bucket"`
		Bucket         string `flag:"bucket" env:"BUCKET" default:"" description:"S3 bucket to upload files to" validate:"nonzero"`
		ContentType    string `flag:"content-type,c" default:"" description:"Force content-type to be set to this value"`
		Listen         string `flag:"listen" env:"LISTEN" default:"" description:"Enable HTTP server if set to IP/Port (e.g. ':3000')"`
		Progress       bool   `flag:"progress" default:"false" description:"Show progress bar while uploading"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	//go:embed frontend/*
	frontend embed.FS

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("share %s\n", version)
		os.Exit(0)
	}
}

func main() {
	switch {

	case cfg.Bootstrap:
		if err := doBootstrap(); err != nil {
			log.WithError(err).Fatal("Bootstrap failed")
		}
		log.Info("Bucket bootstrap finished: Frontend uploaded")

	case cfg.Listen != "":
		if err := doListen(); err != nil {
			log.WithError(err).Fatal("HTTP server ended unclean")
		}

	default:
		if err := doCLIUpload(); err != nil {
			log.WithError(err).Fatal("Upload failed")
		}

	}
}

func doCLIUpload() error {
	if len(rconfig.Args()) == 1 {
		return errors.New("missing argument: file to upload")
	}

	if cfg.BaseURL == "" {
		log.Warn("No BaseURL configured, output will be no complete URL")
	}

	var inFile io.ReadSeeker

	inFileName := rconfig.Args()[1]

	if inFileName == "-" {
		inFileName = "stdin"
		if cfg.ContentType == "" {
			// If we don't have an explicitly set content-type assume stdin contains text
			cfg.ContentType = "text/plain"
		}

		// Stdin is not seekable, so we need to buffer it
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, os.Stdin); err != nil {
			log.WithError(err).Fatal("Could not read stdin")
		}

		inFile = bytes.NewReader(buf.Bytes())
	} else {
		inFileHandle, err := os.Open(inFileName)
		if err != nil {
			return errors.Wrap(err, "opening source file")
		}
		defer inFileHandle.Close()
		inFile = inFileHandle
	}

	url, err := executeUpload(inFileName, inFile, true, cfg.ContentType, false)
	if err != nil {
		return errors.Wrap(err, "uploading file")
	}

	fmt.Println(url)
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
