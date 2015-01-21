// create and greet someone
struct Person {
	name = string
	age = int
}

fn greet(Person p): int {
	println("Hello my name is " +  p.name + " and i am " + p.age + " years old!")
}

fn main(): int {
	p = Person
	p.name = "Jon"
	p.age = 99
	greet(p)
}
