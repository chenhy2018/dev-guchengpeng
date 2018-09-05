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
#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif
#include "queue.h"
#include "sip.h"
#include "../../third_party/pjproject-2.7.2/pjmedia/include/pjmedia/sdp.h"
#include "qrtc.h"

#define MESSAGE_QUEUE_MAX (256)
#define CODECS_MAX (128)
#define MEDIA_CONFIG_MAX (64)
#define MAX_FROM_NAME_SIZE (64)
#define MAX_URL_SIZE (128)
#define INVALID_CALL_ID -1
#define MAX_AUDIO_SIZE 160
#define MAX_USER_ID_SIZE 64
#define MAX_CALL_COUNT 32
#define MAX_ACCOUNT 16
#define MAX_ONGOING_CALL_COUNT 32

typedef struct RtpCallback
{
        void (*OnRxRtp)(void *_pUserData, CallbackType _type, void *_pCbData);
}RtpCallback;

typedef void LogFunc(const char *data);

typedef struct {
        int id;
        int nAccountId;
        ReasonCode error;
        char from[MAX_FROM_NAME_SIZE];
        char url[MAX_URL_SIZE];
        struct list_head list;
        MediaInfo mediaInfo;
        SipInviteState callStatus;
        PeerConnection* pPeerConnection;
        pjmedia_sdp_session* pLocal;
        pjmedia_sdp_session* pRemote;
        IceConfig iceConfig;
} Call;

typedef struct {
        MediaConfigSet videoConfigs;
        MediaConfigSet audioConfigs;
        RtpCallback callback;
} UAConfig;

typedef struct {
#ifdef WITH_P2P
        MediaConfigSet *pVideoConfigs;
        MediaConfigSet *pAudioConfigs;
        RtpCallback *pCallback;
        char turnHost[MAX_TURN_HOST_SIZE];
        char turnUsername[MAX_TURN_USR_SIZE];
        char turnPassword[MAX_TURN_PWD_SIZE];
#else
        void *pSdp;
#endif
} CallConfig;

typedef struct {
        AccountID id;
        struct list_head list;
        MessageQueue *pQueue;
        Message *pLastMessage;
        void *pMqttInstance;
        char userId[MAX_USER_ID_SIZE];
#ifndef WITH_P2P
        void *pSdp;
#endif
        pthread_cond_t registerCond;
        SipAnswerCode regStatus;
        CallConfig config;
        Call callList;
} UA;

typedef struct {
        UA UAList;
        bool bInitSdk;
        UAConfig config;
        pthread_mutex_t mutex;
} UAManager;

extern UAManager *pUAManager;;

//call back
SipAnswerCode cbOnIncomingCall(const int _nAccountId, const const char *_pFrom, const void *_pUser,
                               IN const void *_pMedia, OUT int *pCallId);
void cbOnRegStatusChange(IN const int nAccountId, IN const SipAnswerCode RegStatusCode, IN const void *pUser);
void cbOnCallStateChange(IN const int nCallId, IN const int nAccountId, IN const SipInviteState State,
                         IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void *pMedia);

#endif  /*SDK_LOCAL_H*/
