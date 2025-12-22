package uploader

import (
	"io"

	"github.com/cheggaaa/pb/v3"
)

type (
	// Opts define how the upload works
	Opts struct {
		/// Upload-Related

		// Original name of the file to uploaded
		InfileName string
		// ReadSeeker for the content of the upload
		InfileHandle io.ReadSeeker

		// Enable gzip compression on the upload
		ForceGzip bool
		// Set specific mime-type instead of detected from extension
		OverrideMimeType string
		// Use filename from template
		UseCalculatedFilename bool

		/// Display-Related

		// BaseURL to reach the uploaded content
		BaseURL string
		// Template to calculate the uploaded filename from
		FileTemplate string
		// If set this bar will be provided with the progress
		ProgressBar *pb.ProgressBar

		/// S3-Related

		// Bucket to upload into
		Bucket string
		// Endpoint if using MinIO or any other S3-compatible storage
		Endpoint string
	}

	OptsSetter func(*Opts)
)

// WithBucket sets the bucket for the upload
func WithBucket(b string) OptsSetter { return func(o *Opts) { o.Bucket = b } }

// WithCalculatedFilename enables filename calculation
func WithCalculatedFilename(template string) OptsSetter {
	return func(o *Opts) {
		o.FileTemplate = template
		o.UseCalculatedFilename = true
	}
}

// WithEndpoint sets a custom endpoint for S3 connection
func WithEndpoint(e string) OptsSetter { return func(o *Opts) { o.Endpoint = e } }

// WithFile sets the name and handle for the file to upload
func WithFile(name string, handle io.ReadSeeker) OptsSetter {
	return func(o *Opts) {
		o.InfileName = name
		o.InfileHandle = handle
	}
}

// WithGzip enables Gzip compression for the uploaded file
func WithGzip() OptsSetter { return func(o *Opts) { o.ForceGzip = true } }

// WithMimeType sets the mime-type for the upload
func WithMimeType(t string) OptsSetter { return func(o *Opts) { o.OverrideMimeType = t } }

// With returns a copy of the original Opts with all setters applied
func (o Opts) With(setters ...OptsSetter) Opts {
	for _, s := range setters {
		s(&o)
	}

	return o
}
