.PHONY: build clean format

build:
	$(MAKE) -C cpp fast_build

format:
	$(MAKE) -C cpp format

test:
	$(MAKE) -C cpp test
