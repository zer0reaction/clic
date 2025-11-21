install:
	go build -o ~/.local/bin/clic cmd/main.go

uninstall:
	rm -f ~/.local/bin/clic

test: test.cli
	@mkdir -p .build
	clic -o .build/test.s test.cli
	gcc -o .build/test .build/test.s extern.c
	.build/test

clean:
	rm -rf .build
