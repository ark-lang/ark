; ModuleID = 'tuple_test'

@str = private unnamed_addr constant [7 x i8] c"%d %f\0A\00"

declare i32 @printf(i8*, ...)

define fastcc { i32, float } @_F_9something() {
entry:
  %_V_1x = alloca { i32, double }
  store { i32, double } { i32 4, double 2.300000e+00 }, { i32, double }* %_V_1x
  %_V_1y = alloca { i32, float }
  store { i32, float } { i32 0, float 0x4003333340000000 }, { i32, float }* %_V_1y
  %0 = getelementptr { i32, float }* %_V_1y, i32 0
  %1 = load { i32, float }* %0
  ret { i32, float } %1
}

define fastcc i32 @main() {
entry:
  %_V_1x = alloca { i32, float }
  %0 = call fastcc { i32, float } @_F_9something()
  store { i32, float } %0, { i32, float }* %_V_1x
  %_V_1y = alloca i32
  %1 = getelementptr { i32, float }* %_V_1x, i32 0
  %2 = getelementptr { i32, float }* %1, i32 0, i32 0
  %3 = load i32* %2
  store i32 %3, i32* %_V_1y
  %_V_1z = alloca float
  %4 = getelementptr { i32, float }* %_V_1x, i32 0
  %5 = getelementptr { i32, float }* %4, i32 0, i32 1
  %6 = load float* %5
  store float %6, float* %_V_1z
  %7 = getelementptr { i32, float }* %_V_1x, i32 0
  %8 = getelementptr { i32, float }* %7, i32 0, i32 0
  %9 = load i32* %8
  %10 = getelementptr { i32, float }* %_V_1x, i32 0
  %11 = getelementptr { i32, float }* %10, i32 0, i32 1
  %12 = load float* %11
  %13 = fpext float %12 to double
  %14 = call i32 (i8*, ...)* @printf(i8* getelementptr inbounds ([7 x i8]* @str, i32 0, i32 0), i32 %9, double %13)
  %15 = getelementptr i32* %_V_1y, i32 0
  %16 = load i32* %15
  ret i32 %16
}
