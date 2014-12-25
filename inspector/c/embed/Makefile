
SRC = eq_embed.cc i2g_message.pb.cc
HDR = i2g_message.pb.h

.PHONY: eq_embed.so

eq_embed.so: $(SRC) $(HDR)
	g++ -static -fPIC -std=c++11 -fno-exceptions -o eq_embed.o -c eq_embed.cc
	g++ -static -lprotobuf -fPIC -std=c++11 -fno-exceptions -o i2g_message.o -c i2g_message.pb.cc
	gcc -shared -o eq_embed.so eq_embed.o i2g_message.o -lprotobuf -lstdc++ -lpthread

install: eq_embed.so
	cp eq_embed.so /usr/local/lib
