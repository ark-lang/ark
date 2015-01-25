// this is a weird inheritance idea that doesnt parse

struct a {
	int x = 10
}

struct b {
	int y = 21
}

// wrapper anonymous type thing
anon name {
	a,
	b
}

// function takes this anonymous thing
fn whatever(anon name) {
	// sets x to 21 if name is the type of a
	name::a.x = 21

	// sets y to 52 if name is the type of b
	name::b.y = 52
}

fn main() {
	// create struct things
	a ^swag = make...
	b ^yolo = make...

	// instance of anon thing
	name thing = a

	// pass to whatever
	whatever(thing)
}
