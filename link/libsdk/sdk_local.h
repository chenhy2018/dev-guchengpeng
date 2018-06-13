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
#include "../../third_party/pjproject-2.7.2/pjmedia/include/pjmedia/sdp.h"
#include "qrtc.h"

#define MESSAGE_QUEUE_MAX (256)
#define CODECS_MAX (128)
#define MEDIA_CONFIG_MAX (64)

typedef struct RtpCallback
{
        void (*OnRxRtp)(void *_pUserData, CallbackType _type, void *_pCbData);
}RtpCallback;

typedef struct {
        int id;
        int nAccountId;
        struct list_head list;
        SipInviteState callStatus;
        PeerConnection* pPeerConnection;
        pjmedia_sdp_session* pOffer;
        pjmedia_sdp_session* pAnswer;
        IceConfig iceConfig;
} Call;

typedef struct {
        MediaConfigSet videoConfigs;
        MediaConfigSet audioConfigs;
        RtpCallback callback;
} UAConfig;

typedef struct {
        MediaConfigSet *pVideoConfigs;
        MediaConfigSet *pAudioConfigs;
        RtpCallback *pCallback;
        char turnHost[MAX_TURN_HOST_SIZE];
        char turnUsername[MAX_TURN_USR_SIZE];
        char turnPassword[MAX_TURN_PWD_SIZE];
} CallConfig;

typedef struct {
        AccountID id;
        struct list_head list;
        MessageQueue *pQueue;
        Message *pLastMessage;
        void *pMqttInstance;
        pthread_cond_t registerCond;
        SipAnswerCode regStatus;
        CallConfig config;
        Call callList;
}UA;

typedef struct {
        UA UAList;
        bool bInitSdk;
        UAConfig config;
        pthread_mutex_t mutex;
}UAManager;

extern UAManager *pUAManager;;

#endif  /*SDK_LOCAL_H*/
