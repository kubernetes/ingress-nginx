all:

change-version:
	bin/change-version.sh $(version)

package:
	bin/package.sh $(version)

deploy:
	bin/deploy.sh $(version)

.PHONY: build
build:
	bin/package.sh
	bin/docker-build.sh $(name)
	bin/clean.sh

chart:
	bin/chart.sh $(version)
