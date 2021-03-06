[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/share)](https://goreportcard.com/report/github.com/Luzifer/share)
![](https://badges.fyi/github/license/Luzifer/share)
![](https://badges.fyi/github/downloads/Luzifer/share)
![](https://badges.fyi/github/latest-release/Luzifer/share)

# Luzifer / share

`share` is a small replacement I wrote for sharing my files through external services like CloudApp using Amazon S3. Files are uploaded using this utility into S3 and previewed (if supported) using the included frontend.

## Browser Support

The web frontend uses babel to ensure the JavaScript is supported by [browsers with more than 0.25% market share excluding Internet Explorer 11 and Opera Mini](http://browserl.ist/?q=%3E0.25%25%2C+not+ie+11%2C+not+op_mini+all).

## Setup / usage

- Create a S3 bucket and CloudFront distribution  
  (See [docs/cloudformation.yml](docs/cloudformation.yml) for an example stack)
- Run bootstrap to initialize frontend files:  
  `./share --bucket=<bucket from step 1> --bootstrap`
- Upload files to your sharing bucket:  
  `./share --bucket=<bucket from step 1> --base-url='https://your.site.com/#' <yourfile>`
- Share the URL you received from last step

After you've updated the binary you need to run the `--bootstrap` command once more to have the latest interface changes uploaded to your bucket.
