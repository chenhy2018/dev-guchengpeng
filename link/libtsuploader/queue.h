#ifndef __CIRCLE_QUEUE_H__
#define __CIRCLE_QUEUE_H__

#include <pthread.h>

enum CircleQueuePolicy{
        TSQ_FIX_LENGTH,
        TSQ_VAR_LENGTH
};

typedef struct _CircleQueue CircleQueue;


typedef int(*CircleQueuePush)(CircleQueue *pQueue, char * pData, int nDataLen);
typedef int(*CircleQueuePop)(CircleQueue *pQueue, char * pBuf, int nBufLen);
typedef int(*CircleQueuePopWithTimeout)(CircleQueue *pQueue, char * pBuf, int nBufLen, int64_t nTimeoutAfterUsec);
typedef void(*CircleQueueStopPush)(CircleQueue *pQueue);

typedef struct _CircleQueue{
        CircleQueuePush Push;
        CircleQueuePop Pop;
        CircleQueuePopWithTimeout PopWithTimeout;
        CircleQueueStopPush StopPush;
}CircleQueue;

int NewCircleQueue(CircleQueue **pQueue, enum CircleQueuePolicy policy, int nMaxItemLen, int nInitItemCount);
void DestroyQueue(CircleQueue **_pQueue);

#endif
