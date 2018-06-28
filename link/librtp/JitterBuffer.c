#include "JitterBuffer.h"

typedef struct _JitterBufferFrame {
        int nSeq;
        uint32_t nTs;
        int nFrameSize;
        void * pData;
}JitterBufferFrame;

void JitterBufferDestroy(OUT JitterBuffer *_pJbuf)
{
        if (_pJbuf->pJitterPool) {
	        heap_destroy(&_pJbuf->heap);
                pj_pool_release(_pJbuf->pJitterPool);
                _pJbuf->pJitterPool = NULL;
        }
}


pj_status_t JitterBufferInit(OUT JitterBuffer *_pJbuf, IN int _nMaxBufferCount, IN int _nInitCacheCount,
                      IN pj_pool_t *_pJitterPool, IN int _nMaxFrameSize)
{
        pj_assert(_pJbuf);
        pj_bzero(_pJbuf, sizeof(JitterBuffer));
        _pJbuf->nInitCacheCount = _nInitCacheCount;
        _pJbuf->nMaxBufferCount = _nMaxBufferCount;
        _pJbuf->pJitterPool = _pJitterPool;
        _pJbuf->nMaxFrameSize = _nMaxFrameSize;
        _pJbuf->state = JB_STATE_CACHING;
        _pJbuf->nFirstFlag = 1;
        _pJbuf->nLastRecvRtpSeq = -1;

        int nMaxJBFrameSize = _nMaxFrameSize + sizeof(JitterBufferFrame);

        uint8_t *pFrameDataBuf = (uint8_t *)pj_pool_alloc(_pJitterPool, nMaxJBFrameSize * _nMaxBufferCount);
        if (pFrameDataBuf == NULL) {
                return PJ_NO_MEMORY_EXCEPTION;
        }

        heap_create(&_pJbuf->heap, _nMaxBufferCount, NULL);

        for (int i = 0; i < _nMaxBufferCount; i++) {
                _pJbuf->heap.table[i].immutable = pFrameDataBuf + i * nMaxJBFrameSize;
        }

        return PJ_SUCCESS;
}

void JitterBufferPush(IN JitterBuffer *_pJbuf, IN const void *_pFrame, IN int _nFrameSize,
                      IN int _nFrameSeq, IN uint32_t _nTs, OUT int *_pDiscarded)
{
        pj_assert(_pJbuf && _pFrame);

        *_pDiscarded = 0;

        //rtp sequence number restart
        if (_nFrameSeq < _pJbuf->nMaxBufferCount+1 && _pJbuf->nLastRecvRtpSeq > 65000) {
                _nFrameSeq += 65536;
        }

        //drop frame that is arrive too late
        if (_nFrameSeq < _pJbuf->nLastRecvRtpSeq || (_pJbuf->nLastRecvRtpSeq > _nFrameSeq &&
                                                     (_pJbuf->nLastRecvRtpSeq - _nFrameSeq) > 10000 )) {
                *_pDiscarded = 1;
                return;
        }

        JitterBufferFrame * pFrame = NULL;

        if (_pJbuf->nCurrentSize == _pJbuf->nMaxBufferCount) { // jitter buffer is full
                heap_entry *pEntry;
                heap_min(&_pJbuf->heap, &pEntry);
                
                pFrame = (JitterBufferFrame*)(pEntry->immutable);
                if (_nFrameSeq < pFrame->nSeq) {
                        *_pDiscarded = 1;
                        return;
                }
                // drop oldest frame
                heap_delmin(&_pJbuf->heap, &pEntry);
                _pJbuf->nCurrentSize--;
        }

        heap_entry * pEntry = heap_insert(&_pJbuf->heap, (void *)_nFrameSeq, NULL);
        _pJbuf->nCurrentSize++;
        pFrame = (JitterBufferFrame*)(pEntry->immutable);
        pFrame->nSeq = _nFrameSeq & 0x0000FFFF;
        pFrame->nTs = _nTs;
        pFrame->nFrameSize = _nFrameSize;
        pFrame->pData = (uint8_t *)pEntry->immutable + sizeof(JitterBufferFrame);
        pj_memcpy(pFrame->pData, _pFrame, _nFrameSize);

        if (_pJbuf->nCurrentSize == _pJbuf->nInitCacheCount){
                _pJbuf->state = JB_STATE_CACHING_OK;
        }
        return;
}

void JitterBufferPop(IN JitterBuffer *_pJbuf, OUT void *_pFrame, IN OUT int *_pFrameSize,
                     OUT int *_pFrameSeq, OUT uint32_t *_pTs, OUT JBFrameStatus *_pFrameStatus)
{
        if (_pJbuf->state == JB_STATE_CACHING){
                *_pFrameStatus = JBFRAME_STATE_CACHING;
                return;
        }

        if (_pJbuf->nCurrentSize == 0){
                *_pFrameStatus = JBFRAME_STATE_EMPTY;
                return;
        }

        heap_entry *pEntry = NULL;
        heap_min(&_pJbuf->heap, &pEntry);

        int nMinSeq = (int)(pEntry->key);

        JitterBufferFrame *pFrame = (JitterBufferFrame*)pEntry->immutable;
        //what should I do if 65535 is lost. set nMinSeq = nMinSeq & 0x0000FFF
        if (nMinSeq >= 65536) {
                nMinSeq = nMinSeq & 0x0000FFFF;
        }
        pj_assert(nMinSeq == pFrame->nSeq);
        if (_pJbuf->nFirstFlag) {
                _pJbuf->nLastRecvRtpSeq = nMinSeq;
                _pJbuf->nFirstFlag = 0;
                goto GETFRAME;
        } else if (nMinSeq != _pJbuf->nLastRecvRtpSeq + 1) {
                //_pJbuf->nLastRecvRtpSeq++;
                // if JitterBuffer is almost full, and there is no next seq
                // consider it lost by net, so we pick next frame in the jbuf
                if (_pJbuf->nCurrentSize == _pJbuf->nMaxBufferCount - 1) {
                        _pJbuf->nLastRecvRtpSeq = nMinSeq;
                        goto GETFRAME;
                } else {
                        *_pFrameStatus = JBFRAME_STATE_MISSING;
                        return;
                }
        }

        if (pFrame->nSeq == _pJbuf->nLastRecvRtpSeq + 1) {
                _pJbuf->nLastRecvRtpSeq++;
        GETFRAME:
                pj_assert(*_pFrameSize >= pFrame->nFrameSize);

                pj_memcpy(_pFrame, pFrame->pData, pFrame->nFrameSize);
                *_pFrameSize = pFrame->nFrameSize;
                *_pFrameSeq = pFrame->nSeq & 0x0000FFFF;
                *_pTs = pFrame->nTs;
                *_pFrameStatus = JBFRAME_STATE_NORMAL;
        
                _pJbuf->nCurrentSize--;
                heap_delmin(&_pJbuf->heap, &pEntry);
                
                if (_pJbuf->nCurrentSize == 0) {
                        _pJbuf->state = JB_STATE_CACHING;
                }
                if (*_pFrameSeq < _pJbuf->nLastRecvRtpSeq || *_pFrameSeq == 0) {
                        _pJbuf->nLastRecvRtpSeq = *_pFrameSeq;
                        //TOTO rebuild heap
                        heap_rebuild(&_pJbuf->heap);
                }

                return;
        }
}

void JitterBufferPeek(IN JitterBuffer *_pJbuf, OUT const void **_pFrame, OUT int *_pFrameSize,
                      OUT int *_pFrameSeq, OUT uint32_t *_pTs, OUT JBFrameStatus *_pFrameStatus)
{
        
}
