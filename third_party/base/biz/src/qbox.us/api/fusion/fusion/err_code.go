package fusion

import (
	"qbox.us/api/fusion/netpkg/restutil"
)

const (
	ErrOK                  = 200
	ErrInvalidInput        = 400000
	ErrInvalidName         = 400001
	ErrInvalidSourceType   = 400002
	ErrInvalidPlatform     = 400003
	ErrInvalidGeoCover     = 400004
	ErrInvalidProtocol     = 400005
	ErrInvalidLineId       = 400006
	ErrDomainHasExist      = 400007
	ErrNoSuchBucket        = 400008
	ErrNoSuchEntry         = 400009
	ErrBucketExists        = 400010
	ErrTooManyBuckets      = 400011
	ErrInvalidPubDomain    = 400012
	ErrRepeatOperation     = 400013
	ErrInvalidUid          = 400014
	ErrInvalidResult       = 400015
	ErrInvalidTaskId       = 400016
	ErrInvalidResultDesc   = 400017
	ErrInvalidCname        = 400018
	ErrInvalidModifySource = 400019
	ErrDomainNotICP        = 400020
	ErrInvalidMarker       = 400021
	ErrInvalidTestURLPath  = 400022
	ErrInvalidIP           = 400023
	ErrVerifySource        = 400024
	ErrDomainConflict      = 400025
	ErrExistentDomain      = 400026
	ErrExistentTestURLPath = 400027
	ErrModifyV1Domain      = 400028
	ErrConflictBucket      = 400029
	ErrDomainProcessing    = 400030

	ErrInvalidUrl       = 400031
	ErrInvalidHost      = 400032
	ErrPrefetchUrlLimit = 400033
	ErrRefreshUrlLimit  = 400034
	ErrRefreshDirLimit  = 400035
	ErrInvalidRequestId = 400036
	ErrExistentUrl      = 400037
	ErrNoDirAuthority   = 400038
	ErrCommitTooMuch    = 400039

	ErrInvalidSourceDomain   = 400040
	ErrInvalidSourceIp       = 400041
	ErrInvalidCdn            = 400042
	ErrInvalidAdvancedSource = 400043
	ErrInvalidSourceHost     = 400044
	ErrHttpsVerifyFail       = 400045
	ErrInvalidURLRewrite     = 400046

	ErrNoSuchDomain = 404001

	ErrInvalidReferertype  = 400060
	ErrInvalidRefererValue = 400061
	ErrDuplicateDomain     = 400062
	ErrDuplicateCreate     = 400063
	ErrNotOwnDomain        = 400064
	ErrNoSwitchPlatform    = 400065
	ErrDomainFailed        = 400066
	ErrUserConflictDomain  = 400067
	ErrConflictPlatform    = 400068
	ErrInvalidCrtType      = 400069

	ErrInvalidMidSourceAddrs    = 400070
	ErrInvalidMidSourceAddrType = 400071
	ErrInvalidMidSourceConfig   = 400072

	ErrInvalidStartTime = 400080
	ErrInvalidEndTime   = 400081
	ErrInvalidTimeRange = 400082

	ErrInvalidNotetype     = 400090
	ErrNotStandCache       = 400091
	ErrInvalidAbilityType  = 400091
	ErrInvalidAbilityValue = 400092
	ErrInvalidType         = 400093
	ErrInvalidRegion       = 400094
	ErrInvalidPareDomain   = 400095
	ErrOperatingProcessing = 400999
	ErrInvalidBucket       = 400200
	ErrShotLivedCrtFile    = 400201

	ErrInvalidFeature        = 400202
	ErrUserDisabled          = 400203
	ErrNotEnoughTimeACLKeys  = 400204
	ErrInvalidDomainTaskType = 400301

	ErrInternal         = 500000
	ErrSourceRequestDNS = 500001
	ErrInsertData       = 500002
	ErrDeleteData       = 500003
	ErrUpdateData       = 500004
	ErrQueryData        = 500005
	ErrRequestCreateDNS = 500006
	ErrRequestDeleteDNS = 500007
	ErrRequestCreateCDN = 500008
	ErrRequestDeleteCDN = 500009
	ErrGetFusionDomains = 500010

	ErrPubDomain              = 500011
	ErrListBuckets            = 500012
	ErrGetBucketDomains       = 500013
	ErrGetCnameRecord         = 500014
	ErrGetDomainOwner         = 500015
	ErrCdnRefreshUrlLimit     = 500016
	ErrCdnRefreshDirLimit     = 500017
	ErrCdnPrefetchUrlLimit    = 500018
	ErrNonstandardSourceCname = 500019
	ErrInvalidMidSourceState  = 500020
	ErrInvalidSourceHostInCDN = 500921
	ErrInvalidSourceHostInDB  = 500922

	ErrOperatingFrozen   = 500923
	ErrOperatingUnFrozen = 500924

	ErrDNSPositonNotFound = 500925
	ErrNoSuchRegionLine   = 500926
	ErrDNSSync            = 500927
	ErrDomainFreeze       = 500928
	ErrDomainUnFreeze     = 500929

	ErrOperatingDeleted = 500930

	ErrGetTimeAclKeys   = 500931
	ErrInvalidKeyAction = 500932
)

var ErrorText = map[int]string{
	ErrOK:                  "success",
	ErrInvalidInput:        "invalid input",
	ErrInvalidName:         "invalid domain name",
	ErrInvalidSourceType:   "invalid source type",
	ErrInvalidPlatform:     "invalid platform",
	ErrInvalidGeoCover:     "invalid GeoCover",
	ErrInvalidProtocol:     "invalid Protocol",
	ErrInvalidLineId:       "invalid LineId",
	ErrDomainHasExist:      "domain has existed",
	ErrNoSuchBucket:        "no such bucket",
	ErrNoSuchEntry:         "no such entry",
	ErrBucketExists:        "bucket has existed",
	ErrTooManyBuckets:      "too many buckets",
	ErrInvalidPubDomain:    "invalid pub domain",
	ErrRepeatOperation:     "repeat operation",
	ErrInvalidUid:          "uid not match",
	ErrInvalidResult:       "invalid result",
	ErrInvalidTaskId:       "invalid taskId",
	ErrInvalidResultDesc:   "invalid resultDesc",
	ErrInvalidCname:        "invalid cname",
	ErrInvalidModifySource: "current domain can't be modified",
	ErrInvalidMarker:       "invalid marker",
	ErrDomainNotICP:        "domain have no icp",
	ErrInvalidTestURLPath:  "invalid test url path",
	ErrInvalidIP:           "invalid ip",
	ErrVerifySource:        "failed to verify source",
	ErrDomainConflict:      "domain conflict in CDN",
	ErrExistentDomain:      "existent domain",
	ErrExistentTestURLPath: "existent testURLPath",
	ErrModifyV1Domain:      "try to modify v1 domain",
	ErrNoSuchDomain:        "no such domain",
	ErrConflictBucket:      "the same domain, exist conflict bucket",
	ErrDomainProcessing:    "try to operate processing domain",
	ErrInvalidUrl:          "invalid url",
	ErrInvalidHost:         "invalid host",
	ErrPrefetchUrlLimit:    "prefetch url limit error",
	ErrRefreshUrlLimit:     "refresh url limit error",
	ErrRefreshDirLimit:     "refresh dir limit error",
	ErrInvalidRequestId:    "invalid request id",
	ErrExistentUrl:         "url has existed",
	ErrNoDirAuthority:      "refresh dir authority error",
	ErrCommitTooMuch:       "commit too much",

	ErrInvalidSourceDomain: "invalid source domain",
	ErrInvalidSourceIp:     "invalid source ip",
	ErrInvalidCdn:          "invalid cdn",

	ErrInvalidReferertype:  "invalid referer type",
	ErrInvalidRefererValue: "invalid referer value",
	ErrDuplicateDomain:     "duplicate domain",
	ErrDuplicateCreate:     "duplicate create",
	ErrNotOwnDomain:        "not own this domain",
	ErrNoSwitchPlatform:    "cannot switch platform",
	ErrDomainFailed:        "failed domain",
	ErrUserConflictDomain:  "user got conflict domain",
	ErrConflictPlatform:    "conflict platform",

	ErrInvalidStartTime: "invalid start time",
	ErrInvalidEndTime:   "invalid end time",
	ErrInvalidTimeRange: "invalid time range",

	ErrInvalidAdvancedSource: "invalid advanced source",
	ErrInvalidSourceHost:     "invalid source host",
	ErrInvalidURLRewrite:     "invalid urlrewrite",

	ErrInternal:               "internal error",
	ErrSourceRequestDNS:       "request dns for source error",
	ErrInsertData:             "insert data error",
	ErrUpdateData:             "update data error",
	ErrQueryData:              "query data error",
	ErrRequestCreateDNS:       "request create dns record error",
	ErrRequestDeleteDNS:       "request delete dns record error",
	ErrRequestCreateCDN:       "request create cdn record error",
	ErrRequestDeleteCDN:       "request delete cdn record error",
	ErrGetFusionDomains:       "get fusion domains error",
	ErrPubDomain:              "pub domain error",
	ErrListBuckets:            "list buckets error",
	ErrGetBucketDomains:       "get bucket domains  error",
	ErrGetCnameRecord:         "get domain cname error",
	ErrGetDomainOwner:         "get domain owner error",
	ErrCdnRefreshUrlLimit:     "cdn refresh url limit error",
	ErrCdnRefreshDirLimit:     "cdn refresh dir limit error",
	ErrCdnPrefetchUrlLimit:    "cdn prefetch url limit error",
	ErrInvalidMidSourceState:  "invalid midsource state",
	ErrNonstandardSourceCname: "nonstandard edge source domain",
	ErrInvalidSourceHostInCDN: "invalid source host in cdn",
	ErrInvalidSourceHostInDB:  "invalid source host in db",
	ErrHttpsVerifyFail:        "verify https crt fail",
	ErrNotStandCache:          "not  stand cache format",
	ErrOperatingFrozen:        "try to operate frozen domain",
	ErrOperatingUnFrozen:      "try to opearte unfrozen domain",
	ErrDNSPositonNotFound:     "dns position not found",
	ErrNoSuchRegionLine:       "no such regionline",
	ErrDNSSync:                "DNS sync err",
	ErrDomainFreeze:           "fail to freeze domain",
	ErrDomainUnFreeze:         "fail to unfreeze domain",
	ErrOperatingDeleted:       "operate deleted domain",
	ErrGetTimeAclKeys:         "set domain TimeACLTime failed",
	ErrInvalidKeyAction:       "invalid set timeaclkey action",
	ErrUserDisabled:           "User disabled",
	ErrNotEnoughTimeACLKeys:   "not enough timeaclkeys",
}

var (
	QueryDataErr           = restutil.NewError(ErrQueryData, ErrorText[ErrQueryData])
	DeleteDataErr          = restutil.NewError(ErrDeleteData, ErrorText[ErrDeleteData])
	InvalidNameErr         = restutil.NewError(ErrInvalidName, ErrorText[ErrInvalidName])
	NoSuchDomainErr        = restutil.NewError(ErrNoSuchDomain, ErrorText[ErrNoSuchDomain])
	ExistentDomainErr      = restutil.NewError(ErrExistentDomain, ErrorText[ErrExistentDomain])
	InsertDataErr          = restutil.NewError(ErrInsertData, ErrorText[ErrInsertData])
	InvalidCrtTypeErr      = restutil.NewError(ErrInvalidCrtType, "invalid https crt type")
	RepeatOperationErr     = restutil.NewError(ErrRepeatOperation, ErrorText[ErrRepeatOperation])
	DomainProcessingErr    = restutil.NewError(ErrDomainProcessing, ErrorText[ErrDomainProcessing])
	UpdateDataErr          = restutil.NewError(ErrUpdateData, ErrorText[ErrUpdateData])
	NotOwnDomainErr        = restutil.NewError(ErrNotOwnDomain, ErrorText[ErrNotOwnDomain])
	InternalErr            = restutil.NewError(ErrInternal, ErrorText[ErrInternal])
	InvalidNotetypeErr     = restutil.NewError(ErrInvalidNotetype, "invalid note type")
	InvalidSourceDomainErr = restutil.NewError(ErrInvalidSourceDomain, ErrorText[ErrInvalidSourceDomain])
	InvalidLineIdErr       = restutil.NewError(ErrInvalidLineId, ErrorText[ErrInvalidLineId])
	InvalidAbilityTypeErr  = restutil.NewError(ErrInvalidAbilityType, "invalid ability type")
	InvalidAbilityValueErr = restutil.NewError(ErrInvalidAbilityValue, "invalid ability value")
	NoSuchBucketErr        = restutil.NewError(ErrNoSuchBucket, ErrorText[ErrNoSuchBucket])
	InvalidTypeErr         = restutil.NewError(ErrInvalidType, "invalid type")
	InvalidRegionErr       = restutil.NewError(ErrInvalidRegion, "invalid region")
	InvalidPareDomainErr   = restutil.NewError(ErrInvalidPareDomain, "invalid parent domain")

	InvalidMidSourceAddrsErr    = restutil.NewError(ErrInvalidMidSourceAddrs, "invalid midsource addrs")
	InvalidMidSourceAddrTypeErr = restutil.NewError(ErrInvalidMidSourceAddrType, "invalid midsource addrType")
	NonstandardSourceCnameErr   = restutil.NewError(ErrNonstandardSourceCname, "nonstandard edge source domain")
	InvalidSourceHostInCDNErr   = restutil.NewError(ErrInvalidSourceHostInCDN, ErrorText[ErrInvalidSourceHostInCDN])
	InvalidSourceHostInDBErr    = restutil.NewError(ErrInvalidSourceHostInDB, ErrorText[ErrInvalidSourceHostInDB])
	OperatingProcessingErr      = restutil.NewError(ErrOperatingProcessing, "operating processing")
	InvalidURLRewriteErr        = restutil.NewError(ErrInvalidURLRewrite, ErrorText[ErrInvalidURLRewrite])
	InvalidBucketErr            = restutil.NewError(ErrInvalidBucket, "invalid bucket")
	HttpsVerifyFailErr          = restutil.NewError(ErrHttpsVerifyFail, ErrorText[ErrHttpsVerifyFail])
	ShotLivedCrtFileErr         = restutil.NewError(ErrShotLivedCrtFile, "shot lived crt file")
	InvalidDomainTaskTypeErr    = restutil.NewError(ErrInvalidDomainTaskType, "invlaid domain task type")
	InvalidCnameErr             = restutil.NewError(ErrInvalidCname, ErrorText[ErrInvalidCname])

	InvalidUidErr = restutil.NewError(ErrInvalidUid, ErrorText[ErrInvalidUid])

	PubDomainErr = restutil.NewError(ErrPubDomain, ErrorText[ErrPubDomain])

	DNSPositonNotFoundErr = restutil.NewError(ErrDNSPositonNotFound, ErrorText[ErrDNSPositonNotFound])
	NoSuchRegionLineErr   = restutil.NewError(ErrNoSuchRegionLine, ErrorText[ErrNoSuchRegionLine])
	DNSSyncErr            = restutil.NewError(ErrDNSSync, ErrorText[ErrDNSSync])
	DomainFreezeErr       = restutil.NewError(ErrDomainFreeze, ErrorText[ErrDomainFreeze])
	DomainUnFreezeErr     = restutil.NewError(ErrDomainUnFreeze, ErrorText[ErrDomainUnFreeze])
	RequestDeleteDNSErr   = restutil.NewError(ErrRequestDeleteDNS, ErrorText[ErrRequestDeleteDNS])
	RequestCreateDNSErr   = restutil.NewError(ErrRequestCreateDNS, ErrorText[ErrRequestCreateDNS])
	OperatingDeletedErr   = restutil.NewError(ErrOperatingDeleted, ErrorText[ErrOperatingDeleted])
	SetTimeAclKeysErr     = restutil.NewError(ErrGetTimeAclKeys, ErrorText[ErrGetTimeAclKeys])
	InvalidKeyActionErr   = restutil.NewError(ErrInvalidKeyAction, ErrorText[ErrInvalidKeyAction])
	InvalidHostErr        = restutil.NewError(ErrInvalidHost, ErrorText[ErrInvalidHost])

	InvalidFeatureErr = restutil.NewError(ErrInvalidFeature, ErrorText[ErrInvalidFeature])
	UserDisabledErr   = restutil.NewError(ErrUserDisabled, ErrorText[ErrUserDisabled])

	NotEnoughTimeACLKeysErr = restutil.NewError(ErrNotEnoughTimeACLKeys, ErrorText[ErrNotEnoughTimeACLKeys])
	DomainFailedErr         = restutil.NewError(ErrDomainFailed, ErrorText[ErrDomainFailed])

	InvalidMarkerErr = restutil.NewError(ErrInvalidMarker, ErrorText[ErrInvalidMarker])
)
