.PHONY: fast_build all build clean format external external_clean test 

fast_build:
	(stat build && cd build && ninja) || $(MAKE) build

all: build

build: clean
	mkdir -p build && \
	cd build && \
	cmake -G Ninja \
		.. && \
	ninja

clean:
	rm -rf build out bin

format:
	find . -regex '.*\.\(cpp\|hpp\|cc\|cxx\|c\|h\)' -not -path "./build/*" -not -path "./third-party/*" -exec clang-format -style=file -i {} \;
	find . -regex '.*\(CMakeLists.txt\|\.cmake\)' -not -path "./build/*" -not -path "./third-party/*" -exec cmake-format -i {} \;

external:
	cd third-party && $(MAKE)

external_clean:
	cd third-party && $(MAKE) clean

test:
	echo "Testing isnt done yet ;)"
