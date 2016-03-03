#include <stdio.h>

extern void eq_event_func_call(const char *);

/* below eq_dep and __eq_nop() are stuff just for making dependency*/
extern int eq_dep;
static __attribute__((unused)) void __eq_nop(void)
{
	eq_dep++;
}

int main(void)
{
	printf("f1\n");
	eq_event_func_call("f1");
	printf("f2\n");
	eq_event_func_call("f2");
	printf("f3\n");
	eq_event_func_call("f3");

	return 0;
}
