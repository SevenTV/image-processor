.PHONY: build clean format lint

build:
	$(MAKE) -C cpp fast_build

lint:
	yarn prettier --check .
	$(MAKE) -C cpp lint
	$(MAKE) -C go lint

format:
	yarn prettier --write .
	$(MAKE) -C cpp format
	$(MAKE) -C go format

test:
	$(MAKE) -C cpp test
	$(MAKE) -C go test
