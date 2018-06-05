// Last Update:2018-06-03 21:28:36
/**
 * @file mock_api.c
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-29
 */

#include "mqtt.h"

int MqttLibInit()
{
}

int MqttLibCleanup()
{
}

void* MqttCreateInstance(IN const struct MqttOptions* pOption)
{
    (void)pOption;

    return NULL;
}
void MqttDestroy(IN const void* pInstance)
{
    (void)pInstance;
}

int MqttPublish(IN const void* _pInstance, IN char* _pTopic, 
                     IN int _nPayloadlen, IN const void* _pPayload)
{
    (void)_pInstance;
    (void)_pTopic;
    (void)_nPayloadlen;
    (void)_pPayload;

    return 0;
}

int MqttSubscribe(IN const void* _pInstance, IN char* _pTopic)
{
    (void)_pInstance;
    (void)_pTopic;

    return 0;
}

int MqttUnsubscribe(IN const void* _pInstance, IN char* pSub)
{
    (void)_pInstance;
    (void)pSub;
    
    return 0;
}

