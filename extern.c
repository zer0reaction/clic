#include <stdio.h>
#include <stdint.h>

void print_bool(bool b)
{
    if (b) printf("true\n");
    else   printf("false\n");
}

void print_s64(int64_t n)
{
    printf("%ld\n", n);
}

void print_u64(uint64_t n)
{
    printf("%lu\n", n);
}

void print_uint(unsigned int n)
{
    printf("%u\n", n);
}
