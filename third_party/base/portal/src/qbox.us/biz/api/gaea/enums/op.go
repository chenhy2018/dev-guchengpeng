package enums

type Op string // 中文对照表: src/website/assets/js/setting/oplog.js [cnmap.op]

var (
	OpLogin     Op = "login"
	OpLogout    Op = "logout"
	OpSSOLogin  Op = "sso login"
	OpSSOLogout Op = "sso logout"

	OpRegister Op = "register"
	OpActivate Op = "acitvate"

	OpForgetPwd Op = "forget password"
	OpResetPwd  Op = "reset password"
	OpChangePwd Op = "modify password"

	OpChangeInfo   Op = "modify account info"
	OpBindMobile   Op = "bind mobile"
	OpChangeMobile Op = "change mobile"
	OpChangeEmail  Op = "change email"

	OpKeyNew     Op = "create secret key"
	OpKeyDelete  Op = "delete secret key"
	OpKeyDisable Op = "disable secret key"
	OpKeyEnable  Op = "enable secret key"

	OpBucketCreate       Op = "create bucket"
	OpBucketDelete       Op = "delete bucket"
	OpBucketImage        Op = "image site"
	OpBucketUnimage      Op = "unimage site"
	OpBucket404          Op = "setup 404 page"
	OpBucketCancel404    Op = "cancel 404 page"
	OpBucketPublic       Op = "change to public"
	OpBucketPrivate      Op = "change to private"
	OpBucketProtectedOn  Op = "bucket protect"
	OpBucketProtectedOff Op = "cancel bucket protect"
	OpBucketIndexPageOn  Op = "bucket indexpage on"
	OpBucketIndexPageOff Op = "bucket indexpage off"
	OpBucketMaxAge       Op = "bucket maxage"

	OpDomainApply        Op = "apply domain"
	OpDomainDelete       Op = "delete domain"
	OpCustomDomainApply  Op = "apply custom domain"
	OpCustomDomainDelete Op = "delete custorm domain"

	OpBwListApply  Op = "apply black/white list"
	OpBwListDelete Op = "delete black/white list"

	OpLogFileEnable  Op = "enable log"
	OpLogFileDisable Op = "disable log"

	OpFopNewStyle        Op = "create data process style"
	OpFopDeleteStyle     Op = "delete data process style"
	OpFopModifyStyle     Op = "modify data process style"
	OpFopSeperator       Op = "modify seperator"
	OpFopNewPipeline     Op = "create new pipeline"
	OpFopNewJob          Op = "create new job"
	OpFopDeletePipeline  Op = "delete pipeline"
	OpFopNewUfop         Op = "create new ufop"
	OpFopResizeInstance  Op = "ufop resize instance"
	OpFopSwitchVersion   Op = "ufop switch version"
	OpFopInstanceUpgrade Op = "ugrade instance"
	OpFopInstanceStart   Op = "instance start"
	OpFopInstanceStop    Op = "instance stop"
	OpFopDeleteUfop      Op = "delete ufop"
	OpFopSwitchFlavor    Op = "switch flavor"

	OpRechargeAlipayPaying Op = "alipay recharge paying"
	OpRechargeAlipayPayed  Op = "alipay recharge payed"
	OpRechargePrepaidCard  Op = "prepaid card recharge"
	OpRechargeVoucher      Op = "voucher recharge"

	OpThirdSignup Op = "third signup"
	OpThirdBind   Op = "third bind"
	OpThirdUnbind Op = "third unbind"
	OpThirdLogin  Op = "third login"

	OpThirdServiceNrop    Op = "third service nrop"
	OpThirdServiceEnable  Op = "third service enable"
	OpThirdServiceDisable Op = "third service disable"

	OpPiliCreateHub           Op = "pili create hub"
	OpPiliChangeHubAuth       Op = "pili change hub auth"
	OpPiliChangeCallback      Op = "pili change callback"
	OpPiliChangePersistentCfg Op = "pili change persistent cfg"
	OpPiliBindDomain          Op = "pili bind domain"
	OpPiliUnbindDomain        Op = "pili unbind domain"
	OpPiliSetDefaultDomain    Op = "pili set default domain"
	OpPiliDisableStream       Op = "pili disable stream"
	OpPiliDeleteHub           Op = "pili delete hub"
)
