# The CLI language

## The name

**CLI** - **C** **LI**sp. The goal was to create a C-like language
with lisp-like syntax (I don't know lisp though, rofl).

CLI is currently a [stack
machine](https://en.wikipedia.org/wiki/Stack_machine).

The project is in a very early stage of development. Anything can
change at any time. Anything can break. Do not expect much.

The final goal is to create an interesting language with unique
features and self-host the compiler.

## Usage

```cmd
go run main.go --help
```

## Examples

See [the examples directory](/examples).

## Linking with C functions:

```lisp
(exfun print_s64 (n:s64))
```

```c
#include <stdio.h>
#include <stdint.h>

void print_s64(int64_t n)
{
    printf("%ld\n", n);
}
```

Command line flags are kind of scuffed at the moment, but you can
still do something like this:

```cmd
go run main.go -bf "-o out extern.c" <file>
```
