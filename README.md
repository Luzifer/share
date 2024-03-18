[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/share)](https://goreportcard.com/report/github.com/Luzifer/share)
![](https://badges.fyi/github/license/Luzifer/share)
![](https://badges.fyi/github/downloads/Luzifer/share)
![](https://badges.fyi/github/latest-release/Luzifer/share)

# Luzifer / share

`share` is a small replacement I wrote for sharing my files through external services like CloudApp using Amazon S3. Files are uploaded using this utility into S3 and previewed (if supported) using the included frontend.

## Browser Support

The frontend can be used in all modern browsers. Internet Explorer is not supported.

## Setup / usage

- Create a S3 bucket and CloudFront distribution  
  (See [docs/cloudformation.yml](docs/cloudformation.yml) for an example stack)
- Run bootstrap to initialize frontend files:  
  `./share --bucket=<bucket from step 1> --bootstrap`
- Upload files to your sharing bucket:  
  `./share --bucket=<bucket from step 1> --base-url='https://your.site.com/#' <yourfile>`
- Share the URL you received from last step

After you've updated the binary you need to run the `--bootstrap` command once more to have the latest interface changes uploaded to your bucket.

### Templating in `file-template`

You can specify where in the bucket the file should be stored and how it should be named by passing the `--file-template` parameter. It takes a Go template with these placeholders:

- `{{ .Ext }}` - The extension of the file (including the leading dot, i.e. `.txt`)
- `{{ .FileName }}` - The original filename without changes (i.e. `my video.mp4`)
- `{{ .Hash }}` - The SHA256 hash of the file content
- `{{ .SafeFileName }}` - URL-safe version of the filename (i.e. `my-video.mp4`)
- `{{ .UUID }}` - Random UUIDv4 to be used within the URL to make it hard to guess

Examples:

- `--file-template="file/{{ printf \"%.8s\" .Hash}}/{{ .SafeFileName }}"`
- `--file-template="file/{{ printf \"%.8s\" .Hash}}/{{ .UUID }}{{ .Ext }}"`
