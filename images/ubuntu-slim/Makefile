all: push

TAG ?= 0.8
PREFIX ?= gcr.io/google-containers/ubuntu-slim
BUILD_IMAGE ?= ubuntu-build
TAR_FILE ?= rootfs.tar
PUSH_TOOL ?= gcloud

container: clean
	docker build --pull -t $(BUILD_IMAGE) -f Dockerfile.build .
	docker create --name $(BUILD_IMAGE) $(BUILD_IMAGE)
	docker export $(BUILD_IMAGE) > $(TAR_FILE)
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	$(PUSH_TOOL) docker -- push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
	docker rmi -f $(BUILD_IMAGE)   || true
	docker rm  -f $(BUILD_IMAGE)   || true
	rm -f $(TAR_FILE)
