.PHONY: build clean format external external_clean test lint

build:
	(stat build && cd build && ninja && (sudo ninja install || ninja install)) || $(MAKE) _build

_build: clean
	mkdir -p build && \
	cd build && \
	cmake -G Ninja \
		.. && \
	ninja && \
	(sudo ninja install || ninja install)

clean:
	rm -rf build out bin

lint:
	find . -regex '.*\.\(cpp\|hpp\|cc\|cxx\|c\|h\)' -not -path "./build/*" -not -path "./third-party/*" -exec clang-format -style=file --Werror --dry-run {} \;
	find . -regex '.*\(CMakeLists.txt\|\.cmake\)' -not -path "./build/*" -not -path "./third-party/*" -exec cmake-format --check {} \;

format:
	find . -regex '.*\.\(cpp\|hpp\|cc\|cxx\|c\|h\)' -not -path "./build/*" -not -path "./third-party/*" -exec clang-format -style=file -i {} \;
	find . -regex '.*\(CMakeLists.txt\|\.cmake\)' -not -path "./build/*" -not -path "./third-party/*" -exec cmake-format -i {} \;

external:
	$(MAKE) -C third-party all

external_clean:
	$(MAKE) -C third-party clean

dev_deps:
	pip3 install -r requirements.txt

test:
	echo "Testing isnt done yet ;)"
