#include "function.h"
#include "stdio.h"
#include "stdlib.h"
int whatever(int x) {
return x+1;
}
int factorial(float a) {
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
int test(int *d) {
*d =10;
return *d;
}
int main() {
Cat *cat = 0;
int *swag = malloc(sizeof(*swag));
*swag = 23;
printf("%d\n", *swag);
test(swag);
printf("%d\n", *swag);
free(swag);
int z = 0;
int __MtCiijGGwF = whatever(5);
z = __MtCiijGGwF;
printf("%d\n", z);
int fac = 0;
int __2TjTMZloZo = factorial(6);
fac = __2TjTMZloZo;
printf("%d\n", fac);
return 0;
}
