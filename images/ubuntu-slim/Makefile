all: push

TAG ?= 0.6
PREFIX ?= gcr.io/google_containers/ubuntu-slim
BUILD_IMAGE ?= ubuntu-build
TAR_FILE ?= rootfs.tar

container: clean
	docker build --pull -t $(BUILD_IMAGE) -f Dockerfile.build .
	docker create --name $(BUILD_IMAGE) $(BUILD_IMAGE)
	docker export $(BUILD_IMAGE) > $(TAR_FILE)
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
	docker rmi -f $(BUILD_IMAGE)   || true
	docker rm  -f $(BUILD_IMAGE)   || true
	rm -f $(TAR_FILE)
