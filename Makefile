.PHONY: prod-build
APP_NAME=ts

prod-build:
	@GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME) main.go
	@scp ./bin/${APP_NAME} linode:/opt/${APP_NAME}/${APP_NAME}
	@scp ./.env linode:/opt/${APP_NAME}/.env
