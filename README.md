# The CLI language

## The name

**CLI** - **C** **LI**sp (pronounced as *silly*). The goal was to
create a C-like language with lisp-like syntax (I don't know lisp
though, rofl).

CLI is currently a [stack
machine](https://en.wikipedia.org/wiki/Stack_machine).

The project is in a very early stage of development. Anything can
change at any time. Do not expect much.

The final goal is to create an interesting language with unique
features and self-host the compiler.

## Features

### Linking with C funcitons:

```lisp
(exfun print_s64)

(let s64 foo)
(:= foo 1337)
(print_s64 foo)
```

```c
#include <stdio.h>
#include <stdint.h>

void print_s64(int64_t n)
{
	printf("%ld\n", n);
}
```

### Strong static type system:

The following code will not compile. By default all integers are of
type s64.

```lisp
(let u64 foo)
(:= foo 1337)
```
