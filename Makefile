.PHONY: build clean format

build:
	$(MAKE) -C cpp fast_build

format:
	yarn prettier --write .
	$(MAKE) -C cpp format
	$(MAKE) -C go lint

test:
	$(MAKE) -C cpp test
