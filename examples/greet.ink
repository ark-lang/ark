// create and greet someone
struct Person {
	string name = "Default Name"
	int age = 0
}

fn greet(Person p): int {
	println("Hello my name is {}, I am {} years old!", p.name, p.age)
}

fn main(): int {
	Person p
	p.name = "Jon"
	p.age = 99
	greet(p)
}
