package main

//go:generate make pack

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/Luzifer/rconfig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		BaseURL        string `flag:"base-url" default:"" description:"URL to prepend before filename"`
		BasePath       string `flag:"base-path" default:"file/{{ printf \"%.2s\" .Hash }}/{{.Hash}}" description:"Path to upload the file to"`
		Bootstrap      bool   `flag:"bootstrap" default:"false" description:"Upload frontend files into bucket"`
		Bucket         string `flag:"bucket" default:"" description:"S3 bucket to upload files to" validate:"nonzero"`
		VersionAndExit bool   `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	version = "dev"
)

func init() {
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("share %s\n", version)
		os.Exit(0)
	}
}

func main() {
	if cfg.Bootstrap {
		for _, asset := range []string{"index.html", "app.js"} {
			if _, err := executeUpload(asset, asset, bytes.NewReader(MustAsset("frontend/"+asset))); err != nil {
				log.WithError(err).Fatalf("Unable to upload bootstrap asset %q", asset)
			}
		}
		log.Info("Bucket bootstrap finished: Frontend uploaded.")
		return
	}

	if len(rconfig.Args()) == 1 {
		log.Fatalf("Usage: share <file to upload>")
	}

	if cfg.BaseURL == "" {
		log.Error("No BaseURL configured, output will be no complete URL")
	}

	inFile := rconfig.Args()[1]
	inFileHandle, err := os.Open(inFile)
	if err != nil {
		log.WithError(err).Fatal("Unable to open source file")
	}
	upFile, err := calculateUploadFilename(inFile)
	if err != nil {
		log.WithError(err).Fatal("Unable to calculate upload filename")
	}

	url, err := executeUpload(inFile, upFile, inFileHandle)
	if err != nil {
		log.WithError(err).Fatal("Failed to upload file")
	}
	fmt.Println(url)
}

func calculateUploadFilename(inFile string) (string, error) {
	inFile = strings.Replace(inFile, " ", "_", -1)
	upFile := path.Join(cfg.BasePath, path.Base(inFile))

	fileHash, err := hashFile(inFile)
	if err != nil {
		return "", err
	}

	return executeTemplate(upFile, map[string]interface{}{
		"Hash": fileHash,
	})
}

func executeUpload(inFile, upFile string, inFileHandle io.ReadSeeker) (string, error) {
	mimeType := mime.TypeByExtension(path.Ext(inFile))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	log.Debugf("Uploading %q to %q with type %q", inFile, upFile, mimeType)

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	if _, err := svc.PutObject(&s3.PutObjectInput{
		Body:        inFileHandle,
		Bucket:      aws.String(cfg.Bucket),
		ContentType: aws.String(mimeType),
		Key:         aws.String(upFile),
	}); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", cfg.BaseURL, upFile), nil
}

func hashFile(inFile string) (string, error) {
	data, err := ioutil.ReadFile(inFile)
	if err != nil {
		return "", err
	}
	sum1 := sha1.Sum(data)
	return fmt.Sprintf("%x", sum1), nil
}

func executeTemplate(tplStr string, vars map[string]interface{}) (string, error) {
	tpl, err := template.New("basepath").Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, vars)
	return buf.String(), err
}
