build-AnonymousChatFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap .
