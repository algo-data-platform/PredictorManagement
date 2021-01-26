FBTHRIFT = thrift1
EXEC = main/client
THRIFT_FILES = if/echo.thrift
THRIFT_OUT = if
RM = rm -rf

$(EXEC): thrift
	cd main && go build

thrift: $(THRIFT_FILES)
	$(FBTHRIFT) --gen go -out $(THRIFT_OUT) $<
	find $(THRIFT_OUT) -type f | xargs sed -i 's/github.com\/facebook\/fbthrift-go/github.com\/algo-data-platform\/predictor\/golibs\/adgo\/thirdparty\/thrift/g'
clean:
	$(RM) $(EXEC) $(THRIFT_OUT)/echo
