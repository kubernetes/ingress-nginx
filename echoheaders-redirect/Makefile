all: push

# TAG 0.0 shouldn't clobber any release builds
TAG = 1.3
PREFIX = gcr.io/google_containers/echoserver-redirect

container:
	docker build -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)
