// Last Update:2018-06-03 18:10:48
/**
 * @file sdk_local.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-31
 */

#ifndef SDK_LOCAL_H
#define SDK_LOCAL_H

#include "list.h"
#include "sdk_interface.h"
#include "queue.h"

#define MESSAGE_QUEUE_MAX 256
#define CODECS_MAX 128

typedef struct {
    Codec codecs[CODECS_MAX];
    int samplerate;
    int channels;
} StreamInfo;

typedef struct {
    struct list_head list;
    MessageQueue *pQueue;
    Message *pLastMessage;
    AccountID id;
    void *pMqttInstance;
    StreamInfo streamInfo;
} UA;

extern UA UaList;

#endif  /*SDK_LOCAL_H*/
