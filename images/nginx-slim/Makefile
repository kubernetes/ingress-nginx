all: push

# 0.0.0 shouldn't clobber any released builds
TAG = 0.3
PREFIX = gcr.io/google_containers/nginx-slim

container:
	docker build -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
