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
        char turnHost[MAX_TURN_HOST_SIZE];
        char turnUsername[MAX_TURN_USR_SIZE];
        char turnPassword[MAX_TURN_PWD_SIZE];
        MediaConfig* pVideoConfigs;
        MediaConfig* pAudioConfigs;
        Call callList;
}UA;

typedef struct {
        UA UAList;
        bool bInitSdk;
        MediaConfig videoConfigs;
        MediaConfig audioConfigs;
}UAManager;

extern UAManager *pUAManager;;

#endif  /*SDK_LOCAL_H*/
