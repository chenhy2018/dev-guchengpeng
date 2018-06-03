// Last Update:2018-06-03 21:28:36
/**
 * @file mock_api.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-29
 */

#include "mqtt.h"
#include "sdk_interface.h"

int MosquittoLibInit()
{
}

int MosquittoLibCleanup()
{
}

void* MosquittoCreateInstance(IN const struct MosquittoOptions* pOption)
{
    (void)pOption;

    return NULL;
}
void MosquittoDestroy(IN const void* pInstance)
{
    (void)pInstance;
}

int MosquittoPublish(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic, 
                     IN int _nPayloadlen, IN const void* _pPayload)
{
    (void)_pInstance;
    (void)_pMid;
    (void)_pTopic;
    (void)_nPayloadlen;
    (void)_pPayload;

    return 0;
}

int MosquittoSubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic)
{
    (void)_pInstance;
    (void)_pMid;
    (void)_pTopic;

    return 0;
}

int MosquittoUnsubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* pSub)
{
    (void)_pInstance;
    (void)_pMid;
    (void)pSub;
    
    return 0;
}

