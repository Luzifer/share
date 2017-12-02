[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/share)](https://goreportcard.com/report/github.com/Luzifer/share)
![](https://badges.fyi/github/license/Luzifer/share)
![](https://badges.fyi/github/downloads/Luzifer/share)
![](https://badges.fyi/github/latest-release/Luzifer/share)

# Luzifer / share

`share` is a small replacement I wrote for sharing my files through external services like CloudApp using Amazon S3. Files are uploaded using this utility into S3 and previewed (if supported) using the included frontend.

## Setup / usage

- Create a S3 bucket and CloudFront distribution  
  (See [docs/cloudformation.yml](docs/cloudformation.yml) for an example stack)
- Run bootstrap to initialize frontend files:  
  `./share --bucket=<bucket from step 1> --bootstrap`
- Upload files to your sharing bucket:  
  `./share --bucket=<bucket from step 1> --base-url='https://your.site.com/#' <yourfile>`
- Share the URL you received from last step
