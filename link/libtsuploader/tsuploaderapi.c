#include "tsuploaderapi.h"
#include "tsmuxuploader.h"
#include <assert.h>
#include "log.h"
#include <pthread.h>
#include <curl/curl.h>
#include "servertime.h"
#ifndef USE_OWN_TSMUX
#include <libavformat/avformat.h>
#endif

static int volatile nProcStatus = 0;

int LinkInitUploader()
{
        if (nProcStatus) {
                return 0;
        }
#ifndef USE_OWN_TSMUX
    #if LIBAVFORMAT_VERSION_MAJOR < 58
        av_register_all();
    #endif
#endif
        setenv("TZ", "GMT-8", 1);

        Qiniu_Global_Init(-1);

        int ret = 0;
        ret = LinkInitTime();
        if (ret != 0) {
                LinkLogError("InitUploader gettime from server fail:%d", ret);
                return LINK_HTTP_TIME;
        }
        
        ret = LinkStartMgr();
        if (ret != 0) {
                LinkLogError("StartMgr fail");
                return ret;
        }
        nProcStatus = 1;
        LinkLogDebug("main thread id:%ld", (long)pthread_self());
        
        return 0;

}

int LinkCreateAndStartAVUploader(LinkTsMuxUploader **_pTsMuxUploader, LinkMediaArg *_pAvArg, LinkUserUploadArg *_pUserUploadArg)
{
        if (_pUserUploadArg->pToken_ == NULL || _pUserUploadArg->nTokenLen_ == 0 ||
            _pUserUploadArg->pDeviceId_ == NULL || _pUserUploadArg->nDeviceIdLen_ == 0 ||
            _pTsMuxUploader == NULL || _pAvArg == NULL || _pUserUploadArg == NULL) {
                LinkLogError("token or deviceid or argument is null");
                return LINK_ARG_ERROR;
        }

        LinkTsMuxUploader *pTsMuxUploader;
        int ret = LinkNewTsMuxUploader(&pTsMuxUploader, _pAvArg, _pUserUploadArg->pDeviceId_,
                                   _pUserUploadArg->nDeviceIdLen_, _pUserUploadArg->pToken_, _pUserUploadArg->nTokenLen_);
        if (ret != 0) {
                LinkLogError("NewTsMuxUploader fail");
                return ret;
        }
        if (_pUserUploadArg->nUploaderBufferSize != 0) {
                pTsMuxUploader->SetUploaderBufferSize(pTsMuxUploader, _pUserUploadArg->nUploaderBufferSize);
        }
        if (_pUserUploadArg->nNewSegmentInterval != 0) {
                pTsMuxUploader->SetNewSegmentInterval(pTsMuxUploader, _pUserUploadArg->nNewSegmentInterval);
        }
        
        ret = LinkTsMuxUploaderStart(pTsMuxUploader);
        if (ret != 0){
                LinkDestroyTsMuxUploader(&pTsMuxUploader);
                LinkLogError("UploadStart fail:%d", ret);
                return ret;
        }
        *_pTsMuxUploader = pTsMuxUploader;

        return 0;
}

int LinkPushVideo(LinkTsMuxUploader *_pTsMuxUploader, char * _pData, int _nDataLen, int64_t _nTimestamp, int _nIsKeyFrame, int _nIsSegStart)
{
        if (_pTsMuxUploader == NULL || _pData == NULL || _nDataLen == 0) {
                return LINK_ARG_ERROR;
        }
        int ret = 0;
        ret = _pTsMuxUploader->PushVideo(_pTsMuxUploader, _pData, _nDataLen, _nTimestamp, _nIsKeyFrame, _nIsSegStart);
        return ret;
}

int LinkPushAudio(LinkTsMuxUploader *_pTsMuxUploader, char * _pData, int _nDataLen, int64_t _nTimestamp)
{
        if (_pTsMuxUploader == NULL || _pData == NULL || _nDataLen == 0) {
                return LINK_ARG_ERROR;
        }
        int ret = 0;
        ret = _pTsMuxUploader->PushAudio(_pTsMuxUploader, _pData, _nDataLen, _nTimestamp);
        return ret;
}

int LinkUpdateToken(LinkTsMuxUploader *_pTsMuxUploader, char * _pToken, int _nTokenLen)
{
        if (_pTsMuxUploader == NULL || _pToken == NULL || _nTokenLen == 0) {
                return LINK_ARG_ERROR;
        }
        return _pTsMuxUploader->SetToken(_pTsMuxUploader, _pToken, _nTokenLen);
}

void LinkSetUploadBufferSize(LinkTsMuxUploader *_pTsMuxUploader, int _nSize)
{
        if (_pTsMuxUploader == NULL || _nSize < 0) {
                LinkLogError("wrong arg.%p %d", _pTsMuxUploader, _nSize);
                return;
        }
        _pTsMuxUploader->SetUploaderBufferSize(_pTsMuxUploader, _nSize);
}

void LinkSetNewSegmentInterval(LinkTsMuxUploader *_pTsMuxUploader, int _nIntervalSecond)
{
        if (_pTsMuxUploader == NULL || _nIntervalSecond < 0) {
                LinkLogError("wrong arg.%p %d", _pTsMuxUploader, _nIntervalSecond);
                return;
        }
}

void LinkDestroyAVUploader(LinkTsMuxUploader **pTsMuxUploader)
{
        LinkDestroyTsMuxUploader(pTsMuxUploader);
}

int LinkIsProcStatusQuit()
{
        if (nProcStatus == 2) {
                return 1;
        }
        return 0;
}

void LinkUninitUploader()
{
        if (nProcStatus != 1)
                return;
        nProcStatus = 2;
        LinkStopMgr();
        Qiniu_Global_Cleanup();
        
        return;
}

//---------test
#define CALLBACK_URL "http://39.107.247.14:8088/qiniu/upload/callback"
static char gAk[65] = {0};
static char gSk[65] = {0};
static char gBucket[128] = {0};
static char gCallbackUrl[256] = {0};
static int nDeleteAfterDays = -1;
void LinkSetBucketName(char *_pName)
{
        int nLen = strlen(_pName);
        assert(nLen < sizeof(gBucket));
        strcpy(gBucket, _pName);
        gBucket[nLen] = 0;
        
        return;
}

void LinkSetAk(char *_pAk)
{
        int nLen = strlen(_pAk);
        assert(nLen < sizeof(gAk));
        strcpy(gAk, _pAk);
        gAk[nLen] = 0;
        
        return;
}

void LinkSetSk(char *_pSk)
{
        int nLen = strlen(_pSk);
        assert(nLen < sizeof(gSk));
        strcpy(gSk, _pSk);
        gSk[nLen] = 0;
        
        return;
}

void LinkSetCallbackUrl(char *pUrl)
{
        int nLen = strlen(pUrl);
        assert(nLen < sizeof(gCallbackUrl));
        strcpy(gCallbackUrl, pUrl);
        gCallbackUrl[nLen] = 0;
        return;
}

void LinkSetDeleteAfterDays(int nDays)
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
                pToken->nCurlRet = LINK_JSON_FORMAT;
                return 0;
        }
        pTokenStart += strlen("\"token\"");
        while(*pTokenStart++ != '\"') {
        }
        
        char *pTokenEnd = strchr(pTokenStart, '\"');
        if (pTokenEnd == NULL) {
                pToken->nCurlRet = LINK_JSON_FORMAT;
                return 0;
        }
        if (pTokenEnd - pTokenStart >= pToken->nDataLen) {
                pToken->nCurlRet = LINK_BUFFER_IS_SMALL;
        }
        memcpy(pToken->pData, pTokenStart, pTokenEnd - pTokenStart);
        return size * nmemb;
}

int LinkGetUploadToken(char *pBuf, int nBufLen, char *pUrl)
{
#ifdef DISABLE_OPENSSL
        memset(pBuf, 0, nBufLen);
        CURL *curl;
        curl_global_init(CURL_GLOBAL_ALL);
        curl = curl_easy_init();
        if (pUrl != NULL)
                curl_easy_setopt(curl, CURLOPT_URL, pUrl);
        else
                curl_easy_setopt(curl, CURLOPT_URL, "http://47.105.118.51:8086/qiniu/upload/token/testdvice009");
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
