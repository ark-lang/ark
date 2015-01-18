int max = 10
int result = 0

fn do_stuff(int a, int b): int => while true {
	result = (result + (a + b))
	if (result > 20) {
		return result
	}
}

fn main(): void {
	do_stuff(5, 5)
	do_stuff(5, 5)
}