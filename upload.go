package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cheggaaa/pb"
	log "github.com/sirupsen/logrus"
)

func executeUpload(inFileName string, inFileHandle io.ReadSeeker, useCalculatedFilename bool, overrideMimeType string) (string, error) {
	var (
		upFile = inFileName
		err    error
	)

	if useCalculatedFilename {
		if upFile, err = calculateUploadFilename(inFileName, inFileHandle); err != nil {
			return "", fmt.Errorf("Unable to calculate upload filename: %s", err)
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

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)

	ps, err := newProgressSeeker(inFileHandle)
	if err != nil {
		return "", err
	}

	if cfg.Progress {
		bar := pb.New64(ps.Size).Prefix(inFileName).SetUnits(pb.U_BYTES)
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
		Body:        ps,
		Bucket:      aws.String(cfg.Bucket),
		ContentType: aws.String(mimeType),
		Key:         aws.String(upFile),
	}); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", cfg.BaseURL, upFile), nil
}

func calculateUploadFilename(inFile string, inFileHandle io.ReadSeeker) (string, error) {
	upFile := path.Join(
		cfg.BasePath,
		strings.Replace(path.Base(inFile), " ", "_", -1),
	)

	fileHash, err := hashFile(inFileHandle)
	if err != nil {
		return "", err
	}

	return executeTemplate(upFile, map[string]interface{}{
		"Hash": fileHash,
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
	tpl, err := template.New("basepath").Parse(tplStr)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, vars)
	return buf.String(), err
}
