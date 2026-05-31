watch:
	@wgo \
    -exit \
    -file=.go \
    ENVIRONMENT=development node esbuild.config.js \
    :: ENVIRONMENT=development go run . \
    :: wgo -dir=./web node esbuild.config.js

run: build
	@ENVIRONMENT=production ./bin/main

build:
	@ENVIRONMENT=production node esbuild.config.js
	@go build -o ./bin/main .
