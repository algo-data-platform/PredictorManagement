GOEXEC = go/test
CPPEXEC = cpp/test

EXECS = $(GOEXEC) \
		$(CPPEXEC)

OBJS = lib/City.o \
	   cpp/test.o

LIBS = lib/libcity.a
CXXFLAGS = -std=c++11 

$(EXECS): $(LIBS)
	go build -o $@ github.com/algo-data-platform/predictor/golibs/adgo/test/hash/go/
	g++ cpp/test.o -L./lib -lcity -o $(CPPEXEC)

$(LIBS): $(OBJS)
	ar -cr $@ $<

all: $(EXECS)
test: $(EXECS)
	@echo "Starting run test..."
	@./$(CPPEXEC)
	@./$(GOEXEC)

clean:
	$(RM) $(OBJS) $(EXECS) $(LIBS)
