// Last Update:2018-06-04 14:44:07
/**
 * @file sdk_local.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-31
 */

#ifndef SDK_LOCAL_H
#define SDK_LOCAL_H

#include <stdbool.h>
#include <pthread.h>
#include "list.h"
#include "sdk_interface.h"
#include "queue.h"
#include "sip.h"
#include "qrtc.h"

#define MESSAGE_QUEUE_MAX (256)
#define CODECS_MAX (128)
#define MEDIA_CONFIG_MAX (64)

typedef struct {
        int id;
        struct list_head list;
        SipInviteState callStatus;
        PeerConnection* pPeerConnection;
        pjmedia_sdp_session* pOffer;
        pjmedia_sdp_session* pAnswer;
        IceConfig iceConfig;
        MediaConfig audioConfig;
        MediaConfig videoConfig;
        void* pMedia;
}Call;

typedef struct {
        AccountID id;
        struct list_head list;
        MessageQueue *pQueue;
        Message *pLastMessage;
        void *pMqttInstance;
        pthread_cond_t registerCond;
        SipAnswerCode regStatus;
        Call callList;
}UA;

typedef struct {
        UA UAList;
        bool bInitSdk;
        Media mediaConfigs[MEDIA_CONFIG_MAX];
}UAManager;

extern UAManager *pUAManager;;

#endif  /*SDK_LOCAL_H*/
