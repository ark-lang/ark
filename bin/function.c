#include "function.h"
#include "stdio.h"
int factorial(const int a) {
if (a==0) {
return 1;
}
else {
int x = 0;
int __Xhxun_Ziibs = factorial(a-1);
x = __Xhxun_Ziibs;
return x*a;
}
}
int main() {
int fac = 0;
int __BSQ_aVqzzks = factorial(6);
fac = __BSQ_aVqzzks;
printf("factorial of 6 is %d\n", fac);
return 0;
}
