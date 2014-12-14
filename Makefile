COMPILER = clang
C_FLAGS = -Wall -Iincludes/
SOURCES = src/*.c

all: ${SOURCES}
	${COMPILER} -o j4 ${SOURCES} ${C_FLAGS}
