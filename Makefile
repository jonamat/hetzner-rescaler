VERSION = ${shell git describe --tag}
ifeq (VERSION, "")
  VERSION = "v0-alpha"
endif

COMMAND := $(filter-out $(firstword $(MAKECMDGOALS)), $(MAKECMDGOALS))
FLAGS := $(filter-out -Werror, $(CFLAGS))

watch:
	gow run ./main.go $(COMMAND)

run:
	go run ./main.go $(COMMAND)

serve:
	./bin/hetzner-rescaler $(COMMAND)

build:
	go build -v -x -o ./bin/hetzner-rescaler ./main.go

# Build including dynamic libraries
build-static:
	CGO_ENABLED=0 && GOOS=linux && GOARCH=amd64 && go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ./bin/hetzner-rescaler_static ./main.go

# Multi-platform build requires buildx installed
build-multiarch-docker: 
	go mod vendor && \
	docker buildx build --platform linux/arm/v7,linux/amd64,linux/arm64 -t jonamat/hetzner-rescaler:latest -t jonamat/hetzner-rescaler:${VERSION} . --push

build-docker: 
	docker build -t jonamat/hetzner-rescaler:latest -t jonamat/hetzner-rescaler:${VERSION} .

push-docker:
	docker push jonamat/hetzner-rescaler:latest jonamat/hetzner-rescaler:${VERSION}

create-release:
	./scripts/release.sh