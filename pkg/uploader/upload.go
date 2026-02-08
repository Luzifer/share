package uploader

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"fmt"
	"html/template"
	"io"
	"mime"
	"path"
	"strings"
	"time"

	"github.com/Luzifer/share/pkg/progress"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/cheggaaa/pb/v3"
	"github.com/gofrs/uuid"
	"github.com/gosimple/slug"
	"github.com/sirupsen/logrus"
)

const barUpdateInterval = 100 * time.Millisecond

// Run executes the upload process
func Run(opts Opts) (string, error) {
	var (
		upFile = opts.InfileName
		err    error
	)

	if opts.UseCalculatedFilename {
		if upFile, err = calculateUploadFilename(opts.FileTemplate, opts.InfileName, opts.InfileHandle); err != nil {
			return "", fmt.Errorf("calculating upload filename: %w", err)
		}
	}

	mimeType := mime.TypeByExtension(path.Ext(upFile))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	if opts.OverrideMimeType != "" {
		mimeType = opts.OverrideMimeType
	}

	logrus.Debugf("Uploading file to %q with type %q", upFile, mimeType)

	var contentEncoding *string
	if opts.ForceGzip {
		buf := new(bytes.Buffer)
		gw := gzip.NewWriter(buf)

		if _, err := io.Copy(gw, opts.InfileHandle); err != nil {
			return "", fmt.Errorf("compressing file: %w", err)
		}

		if err := gw.Close(); err != nil {
			return "", fmt.Errorf("closing gzip writer: %w", err)
		}

		opts.InfileHandle = bytes.NewReader(buf.Bytes())
		contentEncoding = aws.String("gzip")
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return "", fmt.Errorf("loading AWS uploader config: %w", err)
	}

	var cfgOpts []func(*s3.Options)
	if opts.Endpoint != "" {
		cfgOpts = append(cfgOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(opts.Endpoint)
			o.UsePathStyle = true
		})
	}
	s3Client := s3.NewFromConfig(awsCfg, cfgOpts...)

	ps, err := progress.New(opts.InfileHandle)
	if err != nil {
		return "", fmt.Errorf("creating progress seeker: %w", err)
	}

	if opts.ProgressBar != nil {
		opts.ProgressBar.SetTotal(ps.Size)
		opts.ProgressBar.Set(pb.Bytes, true)
		opts.ProgressBar.Set("prefix", path.Base(opts.InfileName))
		opts.ProgressBar.Start()
		barUpdate := true

		go func() {
			for barUpdate {
				opts.ProgressBar.SetCurrent(ps.Progress)
				time.Sleep(barUpdateInterval)
			}
		}()

		defer func() {
			barUpdate = false
			opts.ProgressBar.SetCurrent(ps.Progress)
			opts.ProgressBar.Finish()
		}()
	}

	if _, err = transfermanager.New(s3Client).
		UploadObject(context.TODO(), &transfermanager.UploadObjectInput{
			Body:            ps,
			Bucket:          aws.String(opts.Bucket),
			ContentEncoding: contentEncoding,
			ContentType:     aws.String(mimeType),
			Key:             aws.String(upFile),
		}); err != nil {
		return "", fmt.Errorf("putting object to S3: %w", err)
	}

	return fmt.Sprintf("%s%s", opts.BaseURL, upFile), nil
}

func calculateUploadFilename(fileTemplate, inFile string, inFileHandle io.ReadSeeker) (string, error) {
	fileHash, err := hashFile(inFileHandle)
	if err != nil {
		return "", err
	}

	safeFileName := strings.Join([]string{
		slug.Make(strings.TrimSuffix(path.Base(inFile), path.Ext(inFile))),
		path.Ext(inFile),
	}, "")

	return executeTemplate(fileTemplate, map[string]interface{}{
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
