.PHONY: echoserv
echoserv: echoserv.o helper.o
	cp ../../embed/eq_embed.so .
	gcc -o echoserv echoserv.o helper.o -leq_embed -Wall

echoserv.o: echoserv.c helper.h
	gcc -o echoserv.o echoserv.c -c -ansi -pedantic -Wall

helper.o: helper.c helper.h
	gcc -o helper.o helper.c -c -ansi -pedantic -Wall

.PHONY: inspect
inspect:
	../../llvm/ninja_build/bin/eq_c_inspector -inspection-list-path=`pwd`/echo_inspection.json echoserv.c --
