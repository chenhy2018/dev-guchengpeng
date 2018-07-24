#ifndef __TS_UPLOADER_API__
#define __TS_UPLOADER_API__

#include "tsmuxuploader.h"
#include "log.h"
#include "base.h"

int InitUploader(char * pUid, char *pDeviceId, char *pBucketName, char * pToken, AvArg *pAvArg);
int UpdateToken(char * pToken);
int PushVideo(char * pData, int nDataLen, int64_t nTimestamp, int nIsKeyFrame, int nIsSegStart);
int PushAudio(char * pData, int nDataLen, int64_t nTimestamp);
void UninitUploader();


//for test
int GetUploadToken(char *pBuf, int nBufLen);
int SetAk(char *pAk);
int SetSk(char *pSk);
int SetBucketName(char *_pName);


#endif
