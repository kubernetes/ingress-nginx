all: push

# 0.0.0 shouldn't clobber any released builds
TAG = 0.13
PREFIX = gcr.io/google_containers/nginx-slim

container:
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
