#include "function.h"
#include "stdio.h"
int add(int a, int b) {
int x = 5;
if (x==5) {
printf("x is five!\n");
}
else {
printf("whatever\n");
}
return a+b;
}
int main() {
add(5, 5);
int x = 5;
printf("%d\n", x);
return 0;
}
