CC = clang
CFLAGS = -g -Iincludes/ -Wall `llvm-config --cflags`
LD=clang++
LDFLAGS=`llvm-config --cxxflags --ldflags --libs core executionengine jit interpreter analysis native bitwriter --system-libs`
SOURCES = src/*.c

all: ${SOURCES}
	@mkdir -p bin/
	$(CC) $(CFLAGS) ${SOURCES} -c ${SOURCES}
	$(LD) *.o $(LDFLAGS) -o bin/alloyc
	@rm *.o
	
clean:
	@rm *.o
	@rm -rf bin

.PHONY: clean
