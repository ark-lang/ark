j4: src/*.c
	clang -o j4 -Iincludes/ src/*.c -Wall

j4gcc: src/*.c
	gcc -o j4 -Iincludes/ src/*.c -Wall
