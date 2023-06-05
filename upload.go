package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cheggaaa/pb"
	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

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
		bar := pb.New64(ps.Size).Prefix(inFileName).SetUnits(pb.U_BYTES)
		bar.Output = os.Stderr
		bar.Start()
		barUpdate := true

		go func() {
			for barUpdate {
				bar.Set64(ps.Progress)
				<-time.After(100 * time.Millisecond)
			}
		}()

		defer func() {
			barUpdate = false
			bar.Finish()
		}()
	}

	if _, err := svc.PutObject(&s3.PutObjectInput{
		Body:            ps,
		Bucket:          aws.String(cfg.Bucket),
		ContentEncoding: contentEncoding,
		ContentType:     aws.String(mimeType),
		Key:             aws.String(upFile),
	}); err != nil {
		return "", err
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

func hashFile(inFileHandle io.ReadSeeker) (string, error) {
	if _, err := inFileHandle.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(inFileHandle)
	if err != nil {
		return "", err
	}
	sum1 := sha1.Sum(data)
	return fmt.Sprintf("%x", sum1), nil
}

func executeTemplate(tplStr string, vars map[string]interface{}) (string, error) {
	tpl, err := template.New("filename").Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, vars)
	return buf.String(), err
}
