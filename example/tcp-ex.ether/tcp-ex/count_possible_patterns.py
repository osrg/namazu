#!/usr/bin/python
N = 10
VALIDATE = False  # validation is slow

from math import *
from itertools import *


def f(n):
    return factorial(2 * n) / (2 ** n)


def g(n):
    def generate_chain(m):
        return chain.from_iterable((('REQ', i), ('RES', i)) for i in range(m))

    def valid(before, after):
        return not(before[0] == 'RES' and
                   after[0] == 'REQ' and
                   before[1] == after[1])

    def generate_permutations(chain):
        for p in permutations(chain):
            if all(valid(c[0], c[1]) for c in combinations(p, 2)):
                yield p

    count = 0
    for p in generate_permutations(generate_chain(n)):
        count += 1
    return count

for i in range(1, N + 1):
    if VALIDATE:
        assert f(i) == g(i)
    print(i, f(i))

