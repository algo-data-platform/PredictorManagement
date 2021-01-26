FBTHRIFT = thrift1
THRIFT_FILES = if/feature_master.thrift
THRIFT_OUT = if
RM = rm -rf

thrift: $(THRIFT_FILES)
	$(FBTHRIFT) --gen go -out $(THRIFT_OUT) $<
	find $(THRIFT_OUT) -type f | xargs sed -i 's/github.com\/facebook\/fbthrift-go/github.com\/algo-data-platform\/predictor\/golibs\/ads_common_go\/thirdparty\/thrift/g'	

clean:
	$(RM) $(THRIFT_OUT)/feature_master
