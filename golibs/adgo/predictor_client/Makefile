FBTHRIFT = thrift1
EXEC = demo/client_demo
THRIFT_FILES = if/predictor.thrift
THRIFT_OUT = if
RM = rm -rf

$(EXEC): 
	cd demo && go build

thrift: $(THRIFT_FILES)
	$(FBTHRIFT) --gen go -out $(THRIFT_OUT) -I .. $<
	find $(THRIFT_OUT) -type f | xargs sed -i 's/github.com\/facebook\/fbthrift-go/github.com\/algo-data-platform\/predictor\/golibs\/ads_common_go\/thirdparty\/thrift/g'	

clean:
	$(RM) $(EXEC) $(THRIFT_OUT)/predictor
