// Autogenerated by Thrift Compiler (facebook)
// DO NOT EDIT UNLESS YOU ARE SURE THAT YOU KNOW WHAT YOU ARE DOING
// @generated

package feature_master

import (
	"bytes"
	"sync"
	"fmt"
	thrift "github.com/algo-data-platform/predictor/golibs/ads_common_go/thirdparty/thrift"
)

// (needed to ensure safety because of naive import list construction.)
var _ = thrift.ZERO
var _ = fmt.Printf
var _ = sync.Mutex{}
var _ = bytes.Equal

const LABEL_SHOW_SERVER = "show_server"
const LABEL_SHOW_CLIENT = "show_client"
const LABEL_CLICK = "click"
const LABEL_LIKE = "like"
const LABEL_FORWARD = "forward"
const LABEL_COMMENT = "comment"

func init() {
}
