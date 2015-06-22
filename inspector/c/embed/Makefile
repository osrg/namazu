
SRC = eq_embed.cc inspector_message.pb.cc mkdep.cc utils.cc
HDR = inspector_message.pb.h utils.h

.PHONY: eq_embed.so

eq_embed.so: $(SRC) $(HDR)
	g++ -Wall -static -fPIC -std=c++11 -fno-exceptions -o eq_embed.o -c eq_embed.cc
	g++ -Wall -static -fPIC -std=c++11 -fno-exceptions -o mkdep.o -c mkdep.cc
	g++ -Wall -static -fPIC -std=c++11 -fno-exceptions -o utils.o -c utils.cc
	g++ -Wall -static -lprotobuf -fPIC -std=c++11 -fno-exceptions -o inspector_message.o -c inspector_message.pb.cc
	gcc -Wall -shared -o eq_embed.so eq_embed.o inspector_message.o mkdep.o utils.o -lprotobuf -lstdc++ -lpthread

install: eq_embed.so
	cp eq_embed.so /usr/local/lib/libeq_embed.so
