#include "uploader.h"
#include "base.h"
#include <string.h>
#include <stdlib.h>
#include <assert.h>
#include <sys/time.h>

size_t getDataCallback(void* buffer, size_t size, size_t n, void* rptr);

#define TS_DIVIDE_LEN 4096

static char gUid[64];
static char gDeviceId[64];

typedef struct _KodoUploader{
        TsUploader uploader;
#ifdef TK_STREAM_UPLOAD
        CircleQueue * pQueue_;
#else
        char *pTsData;
        int nTsDataCap;
        int nTsDataLen;
#endif
        pthread_t workerId_;
        int isThreadStarted_;
        char token_[256];
        char ak_[64];
        char sk_[64];
        char bucketName_[256];
        int deleteAfterDays_;
        char callback_[512];
        int64_t nSegmentId;
        int64_t nFirstFrameTimestamp;
        int64_t nLastFrameTimestamp;
}KodoUploader;

static void setSegmentId(TsUploader* _pUploader, int64_t _nId)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        pKodoUploader->nSegmentId = _nId;
}

static void setAccessKey(TsUploader* _pUploader, char *_pAk, int _nAkLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        assert(sizeof(pKodoUploader->ak_) - 1 > _nAkLen);
        memcpy(pKodoUploader->ak_, _pAk, _nAkLen);
}

static void setSecretKey(TsUploader* _pUploader, char *_pSk, int _nSkLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        assert(sizeof(pKodoUploader->sk_) - 1 > _nSkLen);
        memcpy(pKodoUploader->sk_, _pSk, _nSkLen);
}

static void setBucket(TsUploader* _pUploader, char *_pBucketName, int _nBucketNameLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        assert(sizeof(pKodoUploader->bucketName_) - 1 > _nBucketNameLen);
        memcpy(pKodoUploader->bucketName_, _pBucketName, _nBucketNameLen);
}

static void setCallbackUrl(TsUploader* _pUploader, char *_pCallbackUrl, int _nCallbackUrlLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        assert(sizeof(pKodoUploader->callback_) - 1 > _nCallbackUrlLen);
        memcpy(pKodoUploader->callback_, _pCallbackUrl, _nCallbackUrlLen);
}

static void setDeleteAfterDays(TsUploader* _pUploader, int nDays)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        pKodoUploader->deleteAfterDays_ = nDays;
}

static void setToken(TsUploader* _pUploader, char *pToken)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        assert(strlen(pToken) < sizeof(pKodoUploader->token_));
        strcpy(pKodoUploader->token_, pToken);
}

static void * streamUpload(void *_pOpaque)
{
        KodoUploader * pUploader = (KodoUploader *)_pOpaque;
        
        char *uptoken = NULL;
        Qiniu_Client client;
        if (pUploader->token_[0] == 0) {
                Qiniu_Mac mac;
                mac.accessKey = pUploader->ak_;
                mac.secretKey = pUploader->sk_;
                
                Qiniu_RS_PutPolicy putPolicy;
                Qiniu_Zero(putPolicy);
                putPolicy.scope = pUploader->bucketName_;
                putPolicy.deleteAfterDays = pUploader->deleteAfterDays_;
                uptoken = Qiniu_RS_PutPolicy_Token(&putPolicy, &mac);

                //init
                Qiniu_Client_InitMacAuth(&client, 1024, &mac);
        } else {
                logdebug("client upload");
                uptoken = pUploader->token_;
                Qiniu_Client_InitNoAuth(&client, 1024);
        }
        
        Qiniu_Io_PutRet putRet;
        Qiniu_Io_PutExtra putExtra;
        Qiniu_Zero(putExtra);
        //设置机房域名
        //Qiniu_Use_Zone_Beimei(Qiniu_False);
        //Qiniu_Use_Zone_Huabei(Qiniu_True);
        //Qiniu_Use_Zone_Huadong(Qiniu_True);
        Qiniu_Use_Zone_Huadong(Qiniu_False);
        //Qiniu_Use_Zone_Huanan(Qiniu_True);
        
        //put extra
        //putExtra.upHost="http://nbxs-gate-up.qiniu.com";
        
        char key[128] = {0};
        client.lowSpeedLimit = 30;
        client.lowSpeedTime = 3;
#ifdef TK_STREAM_UPLOAD
        sprintf(key, "%s_%s_%lld_%ld.ts", gUid, gDeviceId, pUploader->nSegmentId, time(NULL));
        Qiniu_Error error = Qiniu_Io_PutStream(&client, &putRet, uptoken, key, pUploader, -1, getDataCallback, &putExtra);
#else
        sprintf(key, "%s_%s_%lld_%lld.ts", gUid, gDeviceId, pUploader->nSegmentId, pUploader->nFirstFrameTimestamp,
                pUploader->nLastFrameTimestamp);
        Qiniu_Error error = Qiniu_Io_PutBuffer(&client, &putRet, uptoken, key, (const char*)pUploader->pTsData,
                                               pUploader->nTsDataLen, &putExtra);
#endif
        if (error.code != 200) {
                logerror("upload file %s:%s error:%s", pUploader->bucketName_, key, Qiniu_Buffer_CStr(&client.b));
                //debug_log(&client, error);
        } else {
                logdebug("upload file %s: key:%s success", pUploader->bucketName_, key);
        }
        
        if (pUploader->token_[0] == 0) {
                Qiniu_Free(uptoken);
        }
        Qiniu_Client_Cleanup(&client);

        return 0;
}

#ifdef TK_STREAM_UPLOAD
size_t getDataCallback(void* buffer, size_t size, size_t n, void* rptr)
{
        KodoUploader * pUploader = (KodoUploader *) rptr;
        return pUploader->pQueue_->Pop(pUploader->pQueue_, buffer, size * n);
}

static int streamUploadStart(TsUploader * _pUploader)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        int ret = pthread_create(&pKodoUploader->workerId_, NULL, streamUpload, _pUploader);
        if (ret == 0) {
                pKodoUploader->isThreadStarted_ = 1;
        }
        return ret;
}

static void streamUploadStop(TsUploader * _pUploader)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pUploader;
        if (pKodoUploader->isThreadStarted_) {
                pKodoUploader->pQueue_->StopPush(pKodoUploader->pQueue_);
                pthread_join(pKodoUploader->workerId_, NULL);
                pKodoUploader->isThreadStarted_ = 0;
        }
        return;
}

static int streamPushData(TsUploader *pTsUploader, char * pData, int nDataLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)pTsUploader;
        return pKodoUploader->pQueue_->Push(pKodoUploader->pQueue_, (char *)pData, nDataLen);
}

#else

static int memUploadStart(TsUploader * _pUploader)
{
        return 0;
}

static void memUploadStop(TsUploader * _pUploader)
{
        return;
}

static int memPushData(TsUploader *pTsUploader, char * pData, int nDataLen)
{
        KodoUploader * pKodoUploader = (KodoUploader *)pTsUploader;
        if (pKodoUploader->pTsData == NULL) {
                pKodoUploader->pTsData = malloc(pKodoUploader->nTsDataCap);
                pKodoUploader->nTsDataLen = 0;
        }
        if (pKodoUploader->nTsDataLen + nDataLen > pKodoUploader->nTsDataCap){
                char * tmp = malloc(pKodoUploader->nTsDataCap * 2);
                memcpy(tmp, pKodoUploader->pTsData, pKodoUploader->nTsDataLen);
                free(pKodoUploader->pTsData);
                pKodoUploader->pTsData = tmp;
                pKodoUploader->nTsDataCap *= 2;
                memcpy(tmp + pKodoUploader->nTsDataLen, pData, nDataLen);
                pKodoUploader->nTsDataLen += nDataLen;
                return nDataLen;
        }
        memcpy(pKodoUploader->pTsData + pKodoUploader->nTsDataLen, pData, nDataLen);
        pKodoUploader->nTsDataLen += nDataLen;
        return nDataLen;
}
#endif

static void getStatInfo(TsUploader *pTsUploader, UploaderStatInfo *_pStatInfo)
{
        KodoUploader * pKodoUploader = (KodoUploader *)pTsUploader;
#ifdef TK_STREAM_UPLOAD
        pKodoUploader->pQueue_->GetStatInfo(pKodoUploader->pQueue_, _pStatInfo);
#else
        _pStatInfo->nLen_ = 0;
        _pStatInfo->nPopDataBytes_ = pKodoUploader->nTsDataLen;
        _pStatInfo->nPopDataBytes_ = pKodoUploader->nTsDataLen;
#endif
        return;
}

void recordTimestamp(TsUploader *_pTsUploader, int64_t _nTimestamp)
{
        KodoUploader * pKodoUploader = (KodoUploader *)_pTsUploader;
        if (pKodoUploader->nFirstFrameTimestamp == -1) {
                pKodoUploader->nFirstFrameTimestamp = _nTimestamp;
                pKodoUploader->nLastFrameTimestamp = _nTimestamp;
        }
        pKodoUploader->nLastFrameTimestamp = _nTimestamp;
        return;
}

int NewUploader(TsUploader ** _pUploader, enum CircleQueuePolicy _policy, int _nMaxItemLen, int _nInitItemCount)
{
        KodoUploader * pKodoUploader = (KodoUploader *) malloc(sizeof(KodoUploader));
        if (pKodoUploader == NULL) {
                return TK_NO_MEMORY;
        }
        memset(pKodoUploader, 0, sizeof(KodoUploader));
#ifdef TK_STREAM_UPLOAD
        int ret = NewCircleQueue(&pKodoUploader->pQueue_, _policy, _nMaxItemLen, _nInitItemCount);
        if (ret != 0) {
                free(pKodoUploader);
                return ret;
        }
#else
        pKodoUploader->nTsDataCap = 1024 * 1024;
#endif
        pKodoUploader->nFirstFrameTimestamp = -1;
        pKodoUploader->nLastFrameTimestamp = -1;
        pKodoUploader->uploader.SetToken = setToken;
        pKodoUploader->uploader.SetAccessKey = setAccessKey;
        pKodoUploader->uploader.SetSecretKey = setSecretKey;
        pKodoUploader->uploader.SetBucket = setBucket;
        pKodoUploader->uploader.SetCallbackUrl = setCallbackUrl;
        pKodoUploader->uploader.SetDeleteAfterDays = setDeleteAfterDays;
#ifdef TK_STREAM_UPLOAD
        pKodoUploader->uploader.UploadStart = streamUploadStart;
        pKodoUploader->uploader.UploadStop = streamUploadStop;
        pKodoUploader->uploader.Push = streamPushData;
#else
        pKodoUploader->uploader.UploadStart = memUploadStart;
        pKodoUploader->uploader.UploadStop = memUploadStop;
        pKodoUploader->uploader.Push = memPushData;
#endif
        pKodoUploader->uploader.GetStatInfo = getStatInfo;
        pKodoUploader->uploader.SetSegmentId = setSegmentId;
        
        *_pUploader = (TsUploader*)pKodoUploader;
        return 0;
}

void DestroyUploader(TsUploader ** _pUploader)
{
        KodoUploader * pKodoUploader = (KodoUploader *)(*_pUploader);
#ifdef TK_STREAM_UPLOAD
        if (pKodoUploader->isThreadStarted_) {
                pthread_join(pKodoUploader->workerId_, NULL);
        }
        DestroyQueue(&pKodoUploader->pQueue_);
#else
        free(pKodoUploader->pTsData);
        streamUpload(* _pUploader);
#endif
        
        free(pKodoUploader);
        * _pUploader = NULL;
        return;
}

int SetUid(char *_pUid)
{
        int ret = 0;
        ret = snprintf(gUid, sizeof(gUid), "%s", _pUid);
        assert(ret > 0);
        if (ret == sizeof(gUid)) {
                logerror("uid:%s is too long", _pUid);
                return TK_ARG_ERROR;
        }
        
        return 0;
}

int SetDeviceId(char *_pDeviceId)
{
        int ret = 0;
        ret = snprintf(gDeviceId, sizeof(gDeviceId), "%s", _pDeviceId);
        assert(ret > 0);
        if (ret == sizeof(gDeviceId)) {
                logerror("deviceid:%s is too long", _pDeviceId);
                return TK_ARG_ERROR;
        }
        
        return 0;
}
