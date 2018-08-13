#include "resource.h"
#include "base.h"

typedef struct _ResourceMgr
{
        CircleQueue * pQueue_;
        pthread_t mgrThreadId_;
        int nQuit_;
        int nIsStarted_;
}ResourceMgr;

static ResourceMgr manager;

static void * recycle(void *_pOpaque)
{
        while(!manager.nQuit_) {
                AsyncInterface *pAsync = NULL;
                int ret = manager.pQueue_->Pop(manager.pQueue_, (char *)(&pAsync), sizeof(AsyncInterface *));
                if (ret == sizeof(TsUploader *)) {
                        fprintf(stderr, "pop from mgr:%p\n", pAsync);
                        AsynFunction func = pAsync->function;
                        func(pAsync);
                }
        }
        return NULL;
}

int PushFunction(void *_pAsyncInterface)
{
        if (!manager.nIsStarted_) {
                return -1;
        }
        return manager.pQueue_->Push(manager.pQueue_, (char *)(&_pAsyncInterface), sizeof(AsyncInterface *));
}

int StartMgr()
{
        if (manager.nIsStarted_) {
                return 0;
        }

        int ret = NewCircleQueue(&manager.pQueue_, TSQ_FIX_LENGTH, sizeof(void *), 100);
        if (ret != 0){
                return ret;
        }

        ret = pthread_create(&manager.mgrThreadId_, NULL, recycle, NULL);
        if (ret != 0) {
                manager.nIsStarted_ = 0;
                return TK_THREAD_ERROR;
        }
        manager.nIsStarted_ = 1;
        
        return 0;
}

void StopMgr()
{
        manager.nQuit_ = 1;
        if (manager.nIsStarted_) {
                pthread_join(manager.mgrThreadId_, NULL);
                manager.nIsStarted_ = 0;
                if (manager.pQueue_) {
                        DestroyQueue(&manager.pQueue_);
                }
        }
        return;
}
