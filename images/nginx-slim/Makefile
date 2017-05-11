all: push

# 0.0.0 shouldn't clobber any released builds
TAG = 0.15
PREFIX = docker.cloudin.com/google_containers/nginx-slim

container: clean
	docker build --no-cache -t $(PREFIX):$(TAG) .

push: container
	docker push $(PREFIX):$(TAG)

clean:
	docker rmi -f $(PREFIX):$(TAG) || true
