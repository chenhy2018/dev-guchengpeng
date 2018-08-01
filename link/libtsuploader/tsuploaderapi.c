#include "tsuploaderapi.h"
#include "tsmuxuploader.h"
#include <assert.h>
#include "log.h"

static char gAk[65] = {0};
static char gSk[65] = {0};
static char gBucket[65] = {0};
static char gToken[512] = {0};
static int nIsInited = 0;
static TsMuxUploader *gpTsMuxUploader = NULL;
static AvArg gAvArg;

int UpdateToken(char * pToken)
{
        int ret = 0;
        ret = snprintf(gToken, sizeof(gToken), "%s", pToken);
        assert(ret < sizeof(gToken));
        if (ret == sizeof(gToken)) {
                logerror("token:%s is too long", pToken);
                return TK_ARG_ERROR;
        }
        
        return 0;
}

int SetBucketName(char *_pName)
{
        int ret = 0;
        ret = snprintf(gBucket, sizeof(gBucket), "%s", _pName);
        assert(ret < sizeof(gBucket));
        if (ret == sizeof(gBucket)) {
                logerror("bucketname:%s is too long", _pName);
                return TK_ARG_ERROR;
        }
        
        return 0;
}

int InitUploader(char * _pUid, char *_pDeviceId, char * _pToken, AvArg *_pAvArg)
{
        if (nIsInited) {
                return 0;
        }
        
        Qiniu_Global_Init(-1);

        int ret = 0;
        ret = UpdateToken(_pToken);
        if (ret != 0) {
                return ret;
        }

        ret = SetUid(_pUid);
        if (ret != 0) {
                return ret;
        }
        
        ret = SetDeviceId(_pDeviceId);
        if (ret != 0) {
                return ret;
        }
        
        ret = StartMgr();
        if (ret != 0) {
                logerror("StartMgr fail\n");
                return 0;
        }
        logdebug("main thread id:%ld\n", (long)pthread_self());
        logdebug("main thread id:%ld\n", (long)pthread_self());

        gAvArg = *_pAvArg;
        ret = NewTsMuxUploader(&gpTsMuxUploader, &gAvArg);
        if (ret != 0) {
                StopMgr();
                logerror("NewTsMuxUploader fail\n");
                return ret;
        }

        gpTsMuxUploader->SetToken(gpTsMuxUploader, gToken);
        ret = TsMuxUploaderStart(gpTsMuxUploader);
        if (ret != 0){
                StopMgr();
                DestroyTsMuxUploader(&gpTsMuxUploader);
                logerror("UploadStart fail:%d\n", ret);
                return ret;
        }
        gpTsMuxUploader->SetCallbackUrl(gpTsMuxUploader, "http://39.107.247.14:8088/qiniu/upload/callback",
                                        strlen("http://39.107.247.14:8088/qiniu/upload/callback"));
        nIsInited = 1;
        return 0;
}

int PushVideo(char * _pData, int _nDataLen, int64_t _nTimestamp, int _nIsKeyFrame, int _nIsSegStart)
{
        assert(gpTsMuxUploader != NULL);
        int ret = 0;
        ret = gpTsMuxUploader->PushVideo(gpTsMuxUploader, _pData, _nDataLen, _nTimestamp, _nIsKeyFrame, _nIsSegStart);
        return ret;
}

int PushAudio(char * _pData, int _nDataLen, int64_t _nTimestamp)
{
        assert(gpTsMuxUploader != NULL);
        int ret = 0;
        ret = gpTsMuxUploader->PushAudio(gpTsMuxUploader, _pData, _nDataLen, _nTimestamp);
        return ret;
}

void UninitUploader()
{
        if (!nIsInited)
                return;
        DestroyTsMuxUploader(&gpTsMuxUploader);
        StopMgr();
        Qiniu_Global_Cleanup();
}

int SetAk(char *_pAk)
{
        int ret = 0;
        ret = snprintf(gAk, sizeof(gAk), "%s", _pAk);
        assert(ret > 0);
        if (ret == sizeof(gAk)) {
                logerror("sk:%s is too long", _pAk);
                return TK_ARG_ERROR;
        }
        
        return 0;
}

int SetSk(char *_pSk)
{
        int ret = 0;
        ret = snprintf(gSk, sizeof(gSk), "%s", _pSk);
        assert(ret > 0);
        if (ret == sizeof(gSk)) {
                logerror("sk:%s is too long", _pSk);
                return TK_ARG_ERROR;
        }
        
        return 0;
}

int GetUploadToken(char *pBuf, int nBufLen)
{
        if (gAk[0] == 0 || gSk[0] == 0 || gBucket[0] == 0)
                return -11;
        Qiniu_Mac mac;
        mac.accessKey = gAk;
        mac.secretKey = gSk;
        
        Qiniu_RS_PutPolicy putPolicy;
        Qiniu_Zero(putPolicy);
        putPolicy.scope = gBucket;
        putPolicy.deleteAfterDays = 7;
        putPolicy.callbackBody = "{\"key\":\"$(key)\",\"hash\":\"$(etag)\",\"fsize\":$(fsize),\"bucket\":\"$(bucket)\",\"name\":\"$(x:name)\",\"avinfo\":\"$(avinfo)\"}";
        putPolicy.callbackUrl = "http://39.107.247.14:8088/qiniu/upload/callback";
        putPolicy.callbackBodyType = "application/json";
        
        char *uptoken;
        uptoken = Qiniu_RS_PutPolicy_Token(&putPolicy, &mac);
        assert(nBufLen > strlen(uptoken));
        strcpy(pBuf, uptoken);
        Qiniu_Free(uptoken);
        return 0;
}
