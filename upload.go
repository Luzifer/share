package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cheggaaa/pb/v3"
	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const barUpdateInterval = 100 * time.Millisecond

//revive:disable-next-line:flag-parameter // Fine in this case
func executeUpload(inFileName string, inFileHandle io.ReadSeeker, useCalculatedFilename bool, overrideMimeType string, forceGzip bool) (string, error) {
	var (
		upFile = inFileName
		err    error
	)

	if useCalculatedFilename {
		if upFile, err = calculateUploadFilename(inFileName, inFileHandle); err != nil {
			return "", errors.Wrap(err, "calculating upload filename")
		}
	}

	mimeType := mime.TypeByExtension(path.Ext(upFile))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if overrideMimeType != "" {
		mimeType = overrideMimeType
	}

	log.Debugf("Uploading file to %q with type %q", upFile, mimeType)

	var contentEncoding *string
	if forceGzip {
		buf := new(bytes.Buffer)
		gw := gzip.NewWriter(buf)

		if _, err := io.Copy(gw, inFileHandle); err != nil {
			return "", errors.Wrap(err, "compressing file")
		}

		if err := gw.Close(); err != nil {
			return "", errors.Wrap(err, "closing gzip writer")
		}

		inFileHandle = bytes.NewReader(buf.Bytes())
		contentEncoding = aws.String("gzip")
	}

	var awsCfgs []*aws.Config
	if cfg.Endpoint != "" {
		awsCfgs = append(awsCfgs, &aws.Config{Endpoint: &cfg.Endpoint, S3ForcePathStyle: aws.Bool(true)})
	}

	sess := session.Must(session.NewSession(awsCfgs...))
	svc := s3.New(sess)

	ps, err := newProgressSeeker(inFileHandle)
	if err != nil {
		return "", err
	}

	if cfg.Progress {
		bar := pb.New64(ps.Size)
		bar.Set(pb.Bytes, true)
		bar.Set("prefix", inFileName)
		bar.Start()
		barUpdate := true

		go func() {
			for barUpdate {
				bar.SetCurrent(ps.Progress)
				time.Sleep(barUpdateInterval)
			}
		}()

		defer func() {
			barUpdate = false
			bar.Finish()
		}()
	}

	if _, err = svc.PutObject(&s3.PutObjectInput{
		Body:            ps,
		Bucket:          aws.String(cfg.Bucket),
		ContentEncoding: contentEncoding,
		ContentType:     aws.String(mimeType),
		Key:             aws.String(upFile),
	}); err != nil {
		return "", fmt.Errorf("putting object to S3: %w", err)
	}

	return fmt.Sprintf("%s%s", cfg.BaseURL, upFile), nil
}

func calculateUploadFilename(inFile string, inFileHandle io.ReadSeeker) (string, error) {
	fileHash, err := hashFile(inFileHandle)
	if err != nil {
		return "", err
	}

	safeFileName := strings.Join([]string{
		slug.Make(strings.TrimSuffix(path.Base(inFile), path.Ext(inFile))),
		path.Ext(inFile),
	}, "")

	return executeTemplate(cfg.FileTemplate, map[string]interface{}{
		"Ext":          path.Ext(inFile),
		"FileName":     path.Base(inFile),
		"Hash":         fileHash,
		"SafeFileName": safeFileName,
		"UUID":         uuid.Must(uuid.NewV4()).String(),
	})
}

func hashFile(inFileHandle io.ReadSeeker) (hexHash string, err error) {
	if _, err = inFileHandle.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("resetting reader: %w", err)
	}

	shaHash := sha256.New()
	if _, err = io.Copy(shaHash, inFileHandle); err != nil {
		return "", fmt.Errorf("reading data into hash: %w", err)
	}

	return fmt.Sprintf("%x", shaHash.Sum(nil)), nil
}

func executeTemplate(tplStr string, vars map[string]interface{}) (string, error) {
	tpl, err := template.New("filename").Parse(tplStr)
	if err != nil {
		return "", fmt.Errorf("parsing filename template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err = tpl.Execute(buf, vars); err != nil {
		return "", fmt.Errorf("executing filename template: %w", err)
	}

	return buf.String(), nil
}
