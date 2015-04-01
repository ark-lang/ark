CC = clang
LLVM_CONFIG = llvm-config
CFLAGS = -g -Iincludes/ -Wall `llvm-config --cflags` -I`${LLVM_CONFIG} --includedir`
LD=clang++
LDFLAGS=`llvm-config --cxxflags --ldflags --libs core executionengine jit interpreter analysis native bitwriter --system-libs`
SOURCES = src/*.c

all: ${SOURCES}
	llvm-config --version
	@mkdir -p bin/
	$(CC) $(CFLAGS) ${SOURCES} -c ${SOURCES}
	$(LD) *.o $(LDFLAGS) -o bin/alloyc
	@rm *.o
	
clean:
	@rm *.o
	@rm -rf bin

.PHONY: clean
