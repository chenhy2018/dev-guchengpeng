package pay

const (
	PRODUCT_MARKET Product = "market" // market

	GROUP_MARKET_TUPU Group = "tupu" // 图普鉴黄

	MARKET_TUPU_NROP_CERTAIN Item = "market:tupu:nrop:certain" // 鉴黄:确定结果
	MARKET_TUPU_NROP_DEPEND  Item = "market:tupu:nrop:depend"  // 鉴黄:不确定结果

	GROUP_MARKET_TUPU_ADV Group = "tupu_adv" // 图普广告

	MARKET_TUPU_ADV_CERTAIN      Item = "market:tupu:adv:certain"      // 广告:确定结果
	MARKET_TUPU_ADV_DEPEND       Item = "market:tupu:adv:depend"       // 广告:不确定结果
	MARKET_TUPU_ADV_PLUS_CERTAIN Item = "market:tupu:adv_plus:certain" // 广告增强版:确定结果
	MARKET_TUPU_ADV_PLUS_DEPEND  Item = "market:tupu:adv_plus:depend"  // 广告增强版:不确定结果

	GROUP_MARKET_TUPU_VIDEO Group = "tupu_video" // 图普视频鉴黄 (Issue #21159)

	MARKET_TUPU_VIDEO_CERTAIN Item = "market:tupu:video:certain" // 视频鉴黄:确定结果
	MARKET_TUPU_VIDEO_DEPEND  Item = "market:tupu:video:depend"  // 视频鉴黄:不确定结果
	MARKET_TUPU_VIDEO_UNIFIED Item = "market:tupu:video:unified" // 视频鉴黄:统一计费模式 (Issue #24003)

	GROUP_MARKET_TUPU_TERROR Group = "tupu_terror" // 图普鉴暴恐 (Jira BOP-304)

	MARKET_TUPU_TERROR_CERTAIN Item = "market:tupu:terror:certain" // 鉴暴恐:确定结果
	MARKET_TUPU_TERROR_DEPEND  Item = "market:tupu:terror:depend"  // 鉴暴恐:不确定结果

	GROUP_MARKET_YIFANG_CONVERT Group = "yifang_convert" // 亿方云文档转换

	MARKET_YIFANG_CONVERT_WORD  Item = "market:yifang:convert:word"  // 亿方云word转换
	MARKET_YIFANG_CONVERT_EXCEL Item = "market:yifang:convert:excel" // 亿方云excel转换
	MARKET_YIFANG_CONVERT_PPT   Item = "market:yifang:convert:ppt"   // 亿方云ppt转换

	GROUP_MARKET_FACEPP Group = "facepp" // face++

	MARKET_FACEPP_FACECROP2 Item = "market:facepp:facecrop2" // face++ 人脸裁剪 (Issue #20945)

	GROUP_MARKET_SEQUICKIMAGE Group = "sequickimage" // sequickimage

	MARKET_SEQUICKIMAGE_CONVERT Item = "market:sequickimage:convert" // sequickimage convert (Issue #24145)

	GROUP_MARKET_TUSDK Group = "tusdk" // TuSDK

	MARKET_TUSDK_FACE_DETECTION Item = "market:tusdk:face:detection" // TuSDK 人脸检测 (Jira PAY-67)
	MARKET_TUSDK_FACE_LANDMARK  Item = "market:tusdk:face:landmark"  // TuSDK 人脸特征点标识 (Jira PAY-67)

	GROUP_MARKET_DG Group = "dg" // 达观数据

	MARKET_DG_CONTENT_AUDIT_V5 Item = "market:dg:content:audit:v5" // 达观文本鉴黄鉴政 (Jira PAY-67)

	GROUP_MARKET_YUEMIAN Group = "yuemian"

	MARKET_YUEMIAN_FACE_VERIFICATION Item = "market:yuemian:face:verification" // 阅面人脸识别比对 (Jira BOP-387)
	MARKET_YUEMIAN_FACE_LANDMARKS    Item = "market:yuemian:face:landmarks"    // 阅面人脸关键点定位 (Jira BOP-387)
	MARKET_YUEMIAN_FACE_ATTRIBUTES   Item = "market:yuemian:face:attributes"   // 阅面人脸属性识别 (Jira BOP-387)

	GROUP_MARKET_NETEASE Group = "netease"

	MARKET_NETEASE_YDTEXT Item = "market:netease:ydtext" // 网易易盾文本反垃 (Jira BOP-491)
)
