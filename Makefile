SWAG ?= go run github.com/swaggo/swag/cmd/swag@v1.16.4

.PHONY: swagger clean-swagger test

swagger:
	$(SWAG) init -g cmd/botmsg/main.go -o docs --parseInternal --parseDependency --parseDepth 2

clean-swagger:
	rm -rf docs

test:
	go test ./...
