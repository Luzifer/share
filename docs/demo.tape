Output docs/demo.gif

Set FontSize 12
Set Width 811
Set Height 700
Set Padding 10

Hide

# Remove remains of docker container
Type "docker rm -f share-minio"
Enter

# Start MinIO container
Type 'docker run -d --name share-minio -p 9000:9000 -p 9001:9001 --entrypoint sh quay.io/minio/minio -ec "mkdir /data/share && /usr/bin/minio server /data --console-address :9001"'
Enter

# Remove the clutter
Type "clear"
Enter

# Create a big binary to share
Type "make frontend && go build"
Enter 2
Sleep 5s

Type "# Specify S3 credentials through ENV"
Enter
Type "export AWS_ACCESS_KEY_ID=minioadmin AWS_REGION=minio \"
Enter
Type "       AWS_SECRET_ACCESS_KEY=minioadmin ENDPOINT=http://localhost:9000/"
Enter 2
Type "# Set parameters --base-url, -- bucket and --progress through ENV"
Enter
Type "export BUCKET=share BASE_URL=http://localhost:9000/ PROGRESS=true"
Enter 2

Type "# Upload embedded frontend assets to the root of the bucket"
Enter
Show
Type "./share --bootstrap"
Sleep 500ms
Enter

Sleep 2s

Hide
Enter
Type "# Share a text through stdin using a pipe"
Enter
Show
Type "echo 'Hi, I am a shared text!' | share -"
Sleep 500ms
Enter

Sleep 2s

Hide
Enter
Type "# Share a binary by specifying its path"
Enter
Show
Type "./share share"
Sleep 500ms
Enter

Sleep 10s

Hide
Type "docker rm -f share-minio"
Enter
