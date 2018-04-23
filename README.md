[![Go Report Card](https://goreportcard.com/badge/github.com/Luzifer/share)](https://goreportcard.com/report/github.com/Luzifer/share)
![](https://badges.fyi/github/license/Luzifer/share)
![](https://badges.fyi/github/downloads/Luzifer/share)
![](https://badges.fyi/github/latest-release/Luzifer/share)

# Luzifer / share

`share` is a small replacement I wrote for sharing my files through external services like CloudApp using Amazon S3. Files are uploaded using this utility into S3 and previewed (if supported) using the included frontend.

## Browser Support

The web frontend uses ECMAScript 6 (ES6) so a relatively modern browser is required:

- Apple iOS >= 10.0
- Apple Safari >= 10
- Google Chrome >= 58
- Microsoft Edge >= 14
- Mozilla Firefox >= 52

Internet Explorer and older versions of the above listed browsers might work (except IE) but are not supported and will never be.

## Setup / usage

- Create a S3 bucket and CloudFront distribution  
  (See [docs/cloudformation.yml](docs/cloudformation.yml) for an example stack)
- Run bootstrap to initialize frontend files:  
  `./share --bucket=<bucket from step 1> --bootstrap`
- Upload files to your sharing bucket:  
  `./share --bucket=<bucket from step 1> --base-url='https://your.site.com/#' <yourfile>`
- Share the URL you received from last step

After you've updated the binary you need to run the `--bootstrap` command once more to have the latest interface changes uploaded to your bucket.
