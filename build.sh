go build -o wasabi_api
env GOOS=linux GOARCH=arm go build -o wasabi_api_arm
env GOOS=linux GOARCH=amd64 go build -o wasabi_api_x64
