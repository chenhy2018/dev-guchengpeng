#ifndef __JITTERBUFFER_H__
#define __JITTERBUFFER_H__

#include "heap.h"
#include "qrtc.h"
#include <pjlib.h>

typedef enum _JBStatus {
        JB_STATE_CACHING,
        JB_STATE_CACHING_OK
}JBStatus;

typedef enum _JBFrameStatus {
        JBFRAME_STATE_EMPTY,
        JBFRAME_STATE_NORMAL,
        JBFRAME_STATE_MISSING,
        JBFRAME_STATE_CACHING
}JBFrameStatus;

typedef struct _JitterBuffer {
        heap heap;
        int nFirstFlag;
        int nMaxBufferCount;
        int nInitCacheCount; //can pop from JitterBuffer, decide min seq.
                             //every time JB is empty, will cache too
                             //but full to empty will not cache
        int nMaxFrameSize;

        JBStatus state;

        int nCurrentSize;
        int nLastRecvRtpSeq;
        char getBuf[1500];
        pj_pool_t *pJitterPool;
}JitterBuffer;

pj_status_t JitterBufferInit(OUT JitterBuffer *pJbuf, IN int nMaxBufferCount, IN int nInitCacheCount,
                      IN pj_pool_t *pJitterPool, IN int nMaxFrameSize);
void JitterBufferPush(IN JitterBuffer *pJbuf, IN const void *pFrame, IN int nFrameSize,
                        IN int nFrameSeq, IN uint32_t nTs, OUT int *pDiscarded);
void JitterBufferPop(IN JitterBuffer *pJbuf, OUT void *pFrame, IN OUT int *pFrameSize,
                     OUT int *pFrameSeq, OUT uint32_t *pTs, OUT JBFrameStatus *pFrameStatus);
void JitterBufferPeek(IN JitterBuffer *pJbuf, OUT const void **pFrame, OUT int *pFrameSize,
                     OUT int *pFrameSeq, OUT uint32_t *pTs, OUT JBFrameStatus *pFrameStatus);
void JitterBufferDestroy(IN JitterBuffer *pJbuf);
#endif
