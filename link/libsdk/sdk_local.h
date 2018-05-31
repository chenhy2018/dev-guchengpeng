// Last Update:2018-05-31 15:04:31
/**
 * @file sdk_local.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-31
 */

#ifndef SDK_LOCAL_H
#define SDK_LOCAL_H

#define MESSAGE_QUEUE_MAX 256
#define CODECS_MAX 128

extern UA UaList;

typedef struct {
    Codec codecs[CODECS_MAX];
    int samplerate;
    int channels;
} StreamInfo;

typedef struct {
    struct list_head list;
    MessageQueue *pQueue;
    Message *pLastMessage;
    AccountId id;
    void *pMqttInstance;
    StreamInfo streamInfo;
} UA;

#endif  /*SDK_LOCAL_H*/
