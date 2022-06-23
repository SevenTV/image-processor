.PHONY: build clean format external external_clean test lint

build:
	mkdir -p build && \
	cd build && \
	cmake -G Ninja \
		-DCMAKE_INSTALL_PREFIX=$$(pwd)/out \
		.. && \
	ninja && \
	ninja install

clean:
	rm -rf build

lint:
	find . -regex '.*\.\(cpp\|hpp\|cc\|cxx\|c\|h\)' -not -path "./build/*" -not -path "./third-party/*" -not -path "./out/*" -exec clang-format -style=file --Werror --dry-run {} \;
	find . -regex '.*\(CMakeLists.txt\|\.cmake\)' -not -path "./build/*" -not -path "./third-party/*" -not -path "./out/*" -exec cmake-format --check {} \;

format:
	find . -regex '.*\.\(cpp\|hpp\|cc\|cxx\|c\|h\)' -not -path "./build/*" -not -path "./third-party/*" -not -path "./out/*" -exec clang-format -style=file -i {} \;
	find . -regex '.*\(CMakeLists.txt\|\.cmake\)' -not -path "./build/*" -not -path "./third-party/*" -not -path "./out/*" -exec cmake-format -i {} \;

external:
	$(MAKE) -C third-party all

dev_deps:
	pip3 install -r requirements.txt

test:
	echo "Testing isnt done yet ;)"
