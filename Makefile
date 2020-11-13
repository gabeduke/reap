VERSION = $(shell git describe --tags --always --dirty)

DOCKER_REGISTRY = dukeman
DOCKER_IMAGE = reap
DOCKER_TAG = $(VERSION)
DOCKER_CONTEXT = .

build:
	docker buildx build \
	--platform linux/arm64,linux/amd64 \
	--push -t $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_CONTEXT)