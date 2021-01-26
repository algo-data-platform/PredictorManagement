LIBS = libcity.a
OBJS = City.o

%.o: %.c
	g++ -o $@ -c $<

$(LIBS): $(OBJS)
	ar -cr $@ $<

clean:
	$(RM) $(LIBS) $(OBJS)
