flags := --unused
entry_src := tests/test.ark
stdlib_path := .

test_srcs := $(wildcard tests/*.ark)
tests := $(patsubst %.ark,%.test,$(wildcard tests/*.ark))

.PHONY: test clean

%.test: %.ark
	ark build -I$(stdlib_path) -o $@ $< $(flags) $(debug)
	./$@

test: clean $(tests)

clean:
	@-rm $(tests)
