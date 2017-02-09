all: push

# 0.0 shouldn't clobber any released builds
TAG = 0.9.1
PREFIX = gcr.io/google_containers/glbc

server:
	CGO_ENABLED=0 GOOS=linux godep go build -a -installsuffix cgo -ldflags '-w' -o  glbc *.go

container: server
	docker build --pull -t $(PREFIX):$(TAG) .

push: container
	gcloud docker push $(PREFIX):$(TAG)

clean:
	rm -f glbc
