#include "tsuploaderapi.h"
#include "tsmuxuploader.h"
#include <assert.h>
#include "log.h"
#include <pthread.h>
#include <curl/curl.h>
#include "servertime.h"

static char gAk[65] = {0};
static char gSk[65] = {0};
static char gBucket[128] = {0};
static char gCallbackUrl[256] = {0};
static int nProcStatus = 0;
static TsMuxUploader *gpTsMuxUploader = NULL;
static int nDeleteAfterDays = -1;
static AvArg gAvArg;

#define CALLBACK_URL "http://39.107.247.14:8088/qiniu/upload/callback"

typedef struct _Token {
        int nQuit;
        char * pPrevToken_;
        int nPrevTokenLen_;
        char * pToken_;
        int nTokenLen_;
        pthread_mutex_t tokenMutex_;
}Token;
static Token gToken;

int UpdateToken(char * pToken)
{
        if (gToken.nQuit) {
                return 0;
        }
        pthread_mutex_lock(&gToken.tokenMutex_);
        int nTokenLen = strlen(pToken);
        if (gToken.pToken_ == NULL) {
                gToken.pToken_ = malloc(nTokenLen + 1);
        }else {
                if (gToken.pPrevToken_ != NULL) {
                        free(gToken.pPrevToken_);
                }
                gToken.pPrevToken_ = gToken.pToken_;
                gToken.nPrevTokenLen_ = gToken.nTokenLen_;
                
                gToken.pToken_ = malloc(nTokenLen + 1);
        }
        memcpy(gToken.pToken_, pToken, nTokenLen);
        gToken.nTokenLen_ = nTokenLen;
        gToken.pToken_[nTokenLen] = 0;
        if (gpTsMuxUploader != NULL) {
                gpTsMuxUploader->SetToken(gpTsMuxUploader, gToken.pToken_);
        }
        pthread_mutex_unlock(&gToken.tokenMutex_);
        return 0;
}

void SetBucketName(char *_pName)
{
        int nLen = strlen(_pName);
        assert(nLen < sizeof(gBucket));
        strcpy(gBucket, _pName);
        gBucket[nLen] = 0;
        
        return;
}

int InitUploader(char * _pUid, char *_pDeviceId, char * _pToken, AvArg *_pAvArg)
{
        if (nProcStatus) {
                return 0;
        }

        Qiniu_Global_Init(-1);

        int ret = 0;
        ret = InitTime();
        if (ret != 0) {
                return ret;
        }
        ret = pthread_mutex_init(&gToken.tokenMutex_, NULL);
        if (ret != 0) {
                return ret;
        }

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

        gpTsMuxUploader->SetToken(gpTsMuxUploader, gToken.pToken_);
        gpTsMuxUploader->SetCallbackUrl(gpTsMuxUploader, CALLBACK_URL, strlen(CALLBACK_URL));
        ret = TsMuxUploaderStart(gpTsMuxUploader);
        if (ret != 0){
                StopMgr();
                DestroyTsMuxUploader(&gpTsMuxUploader);
                logerror("UploadStart fail:%d\n", ret);
                return ret;
        }

        nProcStatus = 1;
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
        if (nProcStatus != 1)
                return;
        nProcStatus = 2;
        DestroyTsMuxUploader(&gpTsMuxUploader);
        StopMgr();
        Qiniu_Global_Cleanup();
        
        gToken.nQuit = 1;
        pthread_mutex_lock(&gToken.tokenMutex_);
        if (gToken.pToken_) {
                free(gToken.pToken_);
                gToken.pToken_ = NULL;
        }
        if (gToken.pPrevToken_) {
                free(gToken.pPrevToken_);
        }
        pthread_mutex_unlock(&gToken.tokenMutex_);

        pthread_mutex_destroy(&gToken.tokenMutex_);
}

void SetAk(char *_pAk)
{
        int nLen = strlen(_pAk);
        assert(nLen < sizeof(gAk));
        strcpy(gAk, _pAk);
        gAk[nLen] = 0;
        
        return;
}

void SetSk(char *_pSk)
{
        int nLen = strlen(_pSk);
        assert(nLen < sizeof(gSk));
        strcpy(gSk, _pSk);
        gSk[nLen] = 0;
        
        return;
}

void SetCallbackUrl(char *pUrl)
{
        int nLen = strlen(pUrl);
        assert(nLen < sizeof(gCallbackUrl));
        strcpy(gCallbackUrl, pUrl);
        gCallbackUrl[nLen] = 0;
        return;
}

void SetDeleteAfterDays(int nDays)
{
        nDeleteAfterDays = nDays;
}

struct CurlToken {
        char * pData;
        int nDataLen;
        int nCurlRet;
};

size_t writeData(void *pTokenStr, size_t size,  size_t nmemb,  void *pUserData) {
        struct CurlToken *pToken = (struct CurlToken *)pUserData;
        if (pToken->nDataLen < size * nmemb) {
                pToken->nCurlRet = -11;
                return 0;
        }
        char *pTokenStart = strstr(pTokenStr, "\"token\"");
        if (pTokenStart == NULL) {
                pToken->nCurlRet = -11;
                return 0;
        }
        pTokenStart += strlen("\"token\"");
        while(*pTokenStart++ != '\"') {
        }
        
        char *pTokenEnd = strchr(pTokenStart, '\"');
        if (pTokenEnd == NULL) {
                pToken->nCurlRet = -11;
                return 0;
        }
        memcpy(pToken->pData, pTokenStart, pTokenEnd - pTokenStart);
        return size * nmemb;
}

int GetUploadToken(char *pBuf, int nBufLen)
{
#ifdef DISABLE_OPENSSL
        memset(pBuf, 0, nBufLen);
        CURL *curl;
        curl_global_init(CURL_GLOBAL_ALL);
        curl = curl_easy_init();
        curl_easy_setopt(curl, CURLOPT_URL, "http://39.107.247.14:8086/qiniu/upload/token");
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeData);
        
        struct CurlToken token;
        token.pData = pBuf;
        token.nDataLen = nBufLen;
        token.nCurlRet = 0;
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &token);
        int ret =curl_easy_perform(curl);
        if (ret != 0) {
                curl_easy_cleanup(curl);
                return ret;
        }
        curl_easy_cleanup(curl);
        return token.nCurlRet;
#else
        if (gAk[0] == 0 || gSk[0] == 0 || gBucket[0] == 0)
                return -11;
        memset(pBuf, 0, nBufLen);
        Qiniu_Mac mac;
        mac.accessKey = gAk;
        mac.secretKey = gSk;
        
        Qiniu_RS_PutPolicy putPolicy;
        Qiniu_Zero(putPolicy);
        putPolicy.scope = gBucket;
        putPolicy.expires = 40;
        putPolicy.deleteAfterDays = 7;
        putPolicy.callbackBody = "{\"key\":\"$(key)\",\"hash\":\"$(etag)\",\"fsize\":$(fsize),\"bucket\":\"$(bucket)\",\"name\":\"$(x:name)\",\"duration\":\"$(avinfo.format.duration)\"}";
        putPolicy.callbackUrl = CALLBACK_URL;
        putPolicy.callbackBodyType = "application/json";
        
        char *uptoken;
        uptoken = Qiniu_RS_PutPolicy_Token(&putPolicy, &mac);
        assert(nBufLen > strlen(uptoken));
        strcpy(pBuf, uptoken);
        Qiniu_Free(uptoken);
        return 0;
#endif
}