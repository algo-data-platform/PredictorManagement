package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/algo-data-platform/predictor/golibs/ads_common_go/common/ip"
	"github.com/algo-data-platform/predictor/golibs/adgo/feature_master/if/feature_master"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client"
	"github.com/algo-data-platform/predictor/golibs/adgo/predictor_client/if/predictor"
)

func main() {
	localIp, _ := ip.GetLocalIPv4Str()
	fmt.Println("local ip is ", localIp)
	// 测试consul地址：127.0.0.125:8500
	consulAddress := "127.0.0.125:8500"
	// 线上service：algo_service_dev
	serviceName := "algo_service_dev"
	predictorClient, err := predictor_client.NewPredictorClient(consulAddress, serviceName, localIp)
	predictorClient.SetTimeout(15 * time.Millisecond)
	predictorClient.SetMaxConnPerServer(64)
	predictorClient.SetWaitConnection(true)
	// 召回load balance为同机房策略，256 * 256 表示IP后两段不同
	predictorClient.SetIpDiffRange(65535)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to init predictor client ")
		return
	}

	testCalculateBatchVector(predictorClient)

	select {}
}

func testCalculateBatchVector(predictorClient *predictor_client.PredictorClient) {
	requests := predictor.NewCalculateBatchVectorRequests()

	// 特征名
	feature_names := []string{
		"label",
		"logid",
		"deliver_id",
		"feed_id",
		"style_code",
		"compaign_id",
		"cust_uid",
		"industry_id",
		"client_id",
		"cust_id_tags",
		"fansnum",
		"picnum",
		"nick_word",
		"desc_word",
		"content_word",
	}
	// 对应特征值，跟feature_names每列一一对应
	feature_values_list := [][]string{
		[]string{
			"1",
			"__6052798666_8689042_4516778192168669_1593572433",
			"8689042",
			"4516778192168669",
			"000002",
			"2742945",
			"6052798666",
			"135",
			"179955",
			"新款:6,现货:6,牛皮:6,最新款:5",
			"68",
			"4",
			"数码港,氧气,家,",
			"深耕,氧气,温度,圈子,有余,家,八,年,一,个,",
			"YQdigital,144Hz,GTX1660Ti,512SSD,I7-9750,17M-4725,原封,外星人,2,牛>皮纸,16G,10,现货,眼球,联保,指路,换机,超薄,国行,＋V,价,独家,款,官,网,心,台,最新,好,年,",
		},
		[]string{
			"1",
			"__5635652794_8158954_4494301339932712_1593572433",
			"8158954",
			"4494301339932712",
			"000001",
			"2058891",
			"5635652794",
			"118",
			"3837",
			"皮肤:41,疤痕:36,知识:23,吸脂:19,雀斑:18,黄褐斑:17,整形:17,双眼皮:16,丰胸:16,宝宝:14,手术:13,皮秒:12,设计:12,变美:12,纹身:12,皮肤问题:11,棉花糖:11,痘痘:10,痘印:10,胸部:10,脂肪:10,痘坑:9,项目:9,肤色:7,假体:7,毛孔:7,毛孔粗大:7,素颜:6,色斑:6,脸型:6,祛斑:6,色素:6,暗黄:5,隆鼻:5,设计方案:5,祛痘:4,瘦身:4,做双眼皮:4,晒斑:4,斑点:4,肉肉:4,美胸:4,粉刺:4,黑头:3,医生:3,小胸:3,妊娠斑:3,下垂:3,治疗:3,红血丝:3,平胸:3,水动力:3,痤疮:3,电话:3,修复:3,丹凤眼:2,老年:2,美瞳:2,想瘦:2,嫩肤:2,祛痘印:2,保湿:2,减肥:2,体重:2,鼻部:2,老年斑:2,松弛:2,皮肤松弛:2,肤色不均:2,果酸:2,肌肤:2,补水:2,胸小:2,体脂:2,胸型:2,美的:2,红包:2,美瞳线:2,小气泡:2,斑斑:2,减肥瘦身:2,复合:2,细纹:2,脱毛:2,荔枝:2,眼型:2,紧致:2,肌肉:1,皮肤干:1,鼻梁:1,桃花眼:1,水光:1,乳腺:1,美容:1,眉毛:1,视频:1,无痕:1,材料:1,干燥:1,沙龙:1,水光针:1,高鼻梁:1,整容:1,皮肤干燥:1,身体:1,血管:1,产后:1,痘痘烦:1,妆容:1,双眼皮贴:1,注射:1,漫画:1,挤痘痘:1,洗纹身:1,双十一:1,玻尿酸:1,挤痘:1,面诊:1,无痕双眼皮:1,黑眼圈:1,老化:1,高仿:1",
			"165327",
			"1",
			"韩啸米,院,北京,",
			"专注,美颜,美学,修饰,和谐,韩,啸,设计,北京,",
			"吸脂,抽脂,私信,脂肪,留言,知识,部位,了解,北京,抽,评论,掉,二,想,",
		},
		[]string{
			"1",
			"__6144272503_8545023_4512514850004958_1593572433",
			"8545023",
			"4512514850004958",
			"110115",
			"2726856",
			"6144272503",
			"129",
			"217304",
			"",
			"44",
			"0",
			"好车,懂,帝,>车,精选,",
			"",
			"盘锦,买车,底价,辽宁,拒绝,成交价,车型,APP,懂,不懂,车主,帝,坑,车,朋友,千万,行业,深,查询,水,注意,天,看到,全国,看,说,没,所有,到,能,上,",
		},
	}

	// construct request
	request := predictor.NewCalculateBatchVectorRequest()
	// 唯一标识(自定义)
	request.ReqID = "100001"
	// 业务标识
	request.Channel = "advec"
	// 模型名
	request.ModelName = "tf_xfea_estimator_v1_sfst_retrieval_cpl_ad"
	// 构建featruesMap map<adid, vector<feature>>
	request.FeaturesMap = getFeaturesMap(feature_values_list, feature_names)
	// 输出变量来自模型的tf_config.json文件的output_tags，可根据需要获取，如果不需要取predict返回值可不用append到request.OutputNames
	request.OutputNames = append(request.OutputNames, "ad_vec")

	requests.Reqs = append(requests.Reqs, request)

	fmt.Println(requests)
	// call predictor
	t1 := time.Now()
	responses, err := predictorClient.CalculateBatchVector(requests)
	t2 := time.Now()
	diff := t2.Sub(t1)
	fmt.Printf("---------------------------- predictorClient.CalculateBatchVector() cost: %v\n", diff)

	if err != nil {
		fmt.Fprintln(os.Stderr, "CalculateBatchVector failed. error: ", err)
	} else {
		fmt.Println("CalculateBatchVector responses: ", responses)
		// code == 0 代表获取成功，其他为失败
		fmt.Println("return code: ", responses.Resps[0].GetReturnCode())
	}
}

func getFeaturesMap(featureValuesList [][]string, feature_names []string) map[int64]*feature_master.Features {
	featureMap := make(map[int64]*feature_master.Features, len(featureValuesList))
	for _, feature_values := range featureValuesList {
		adid, _ := strconv.ParseInt(feature_values[2], 10, 64)

		// 构建request.Features
		features := feature_master.NewFeatures()
		for idx, name := range feature_names {
			features.Features = append(
				features.Features,
				&feature_master.Feature{
					FeatureName:  name,
					FeatureType:  feature_master.FeatureType_STRING_LIST,
					StringValues: []string{feature_values[idx]},
				},
			)
		}
		featureMap[adid] = features
	}
	return featureMap
}
