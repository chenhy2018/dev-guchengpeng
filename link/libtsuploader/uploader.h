#ifndef __TS_UPLOADER_H__
#define __TS_UPLOADER_H__

#include <qiniu/io.h>
#include <qiniu/rs.h>
#include <pthread.h>
#include <errno.h>
#include "queue.h"


typedef struct _TsUploader TsUploader;
typedef void (*AccessKeySetter)(TsUploader*, char *, int);
typedef void (*SecretKeySetter)(TsUploader*, char *, int);
typedef void (*BucketSetter)(TsUploader*, char *, int);
typedef void (*CallbackUrlSetter)(TsUploader*, char *, int);
typedef void (*DeleteAfterDaysSetter)(TsUploader*, int);
typedef int (*StreamUploadStart)(TsUploader*);
typedef void (*StreamUploadStop)(TsUploader*);

typedef struct _TsUploader{
        AccessKeySetter SetAccessKey;
        SecretKeySetter SetSecretKey;
        BucketSetter SetBucket;
        CallbackUrlSetter SetCallbackUrl;
        DeleteAfterDaysSetter SetDeleteAfterDays;
        StreamUploadStart UploadStart;
        StreamUploadStop UploadStop;
        int(*Push)(TsUploader *pTsUploader, char * pData, int nDataLen);
        void (*GetStatInfo)(TsUploader *pTsUploader, UploaderStatInfo *pStatInfo);
}TsUploader;

int NewUploader(TsUploader ** pUploader, enum CircleQueuePolicy policy, int nMaxItemLen, int nInitItemCount);
void DestroyUploader(TsUploader ** _pUploader);

#endif
