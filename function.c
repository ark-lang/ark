#include "function.h"
#include "stdio.h"
int factorial(const int a) {
if (a==0) {
return 1;
}
else {
int x = 0;
int __bzj9Y0kEpk = factorial(a-1);
x = __bzj9Y0kEpk;
return x*a;
}
}
int main() {
int fac = 0;
int __xOPzPXyKqL = factorial(6);
fac = __xOPzPXyKqL;
printf("factorial of 6 is %d\n", fac);
return 0;
}
