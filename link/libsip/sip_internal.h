#ifndef __SIP_INTERNAL_H__
#define __SIP_INTERNAL_H__
#include <pjsip.h>
#include <pjmedia.h>
#include <pjsip_ua.h>
#include <pjsip_simple.h>
#include <pjlib-util.h>
#include <pjlib.h>
#include "../util/queue.h"
#define SIP_MAX_CRED 1
#define SIP_PORT 5060

#define SIP_REG_INTERVAL 300 // 300 seconds
#define SIP_REG_RETRY_INTERNAL 300 // 300 seconds
#define SIP_REG_DELAY_BEFORE_REFRESH  5 // 5 seconds
#define SIP_UNREG_TIMEOUT 4000 // 4 seconds
#define SIP_KEEP_ALIVE_INTERVAL 15 // 15 seconds
#define SIP_SESSION_EXPIRES 600 // 10 minutes
#define SIP_MIN_SE 90 //90 seconds

#define MESSAGE_QUEUE_MAX 256

#define MUTEX_LOCK(mutex) { PJ_LOG(5, (THIS_FILE, "get lock in %s on line %d", __FUNCTION__, __LINE__));pj_mutex_lock(mutex);}

#define MUTEX_FREE(mutex) { PJ_LOG(5, (THIS_FILE, "free lock in %s on line %d", __FUNCTION__, __LINE__)); pj_mutex_unlock(mutex);}

#define CHECK_RETURN(expr,retval)               \
        do {                                    \
                if (!(expr)) {return retval; }  \
        } while(0)

/**
 * Account info
 */
typedef struct SipAccount {
        pj_pool_t *pPool;

        pj_str_t Id;
        pj_str_t RegUri;
        pj_str_t UserName;
        pj_str_t SipDomain;
        pj_str_t Password;
        pj_str_t Scheme;
        pj_str_t Realm;
        pj_str_t KaData;

        unsigned nRegTimeout;
        unsigned nUnRegTimeout;
        unsigned nRegDelayBeforeRefresh;
        unsigned nRegRetryInterval;
        unsigned nKaInterval;
        int nMaxOngoingCall;

        pjsip_timer_setting TimerSetting;

        int nIndex;
        pj_bool_t bValid;
        pjsip_regc *pRegc;
        pj_str_t Contact;
        int nLastRegCode;
        int nOngoingCall;

        struct {
                pj_bool_t Active;
                pj_timer_entry ReRegTimer;
        } AutoReReg;
        unsigned nCredCnt;
        pjsip_cred_info  Cred[SIP_MAX_CRED];
        pjsip_host_port ViaAddr;
        /* Add Keepalive timer for this account */
        pj_timer_entry   KaTimer;
        pjsip_transport *KaTransport;
        pj_sockaddr     KaTarget;
        unsigned     KaTargetLen;

        void *pUserData;
} SipAccount;

/* Call info */
typedef struct SipCall
{
        unsigned nIndex;
        unsigned nAccountId;
        pj_bool_t bValid;
        pjsip_inv_session *pInviteSession;
        pj_time_val StartTime;
        pj_time_val ResponseTime;
        pj_time_val ConnectTime;
} SipCall;

/**
  * Global pjsua application data.
  */
struct SipData {
        pj_caching_pool Cp;
        pj_pool_t *pPool;
        pj_mutex_t *pMutex;
        pj_str_t LocalIp;
        int LocalPort;
        int nMaxAccount;
        int nMaxCall;

        pjsip_endpoint *pSipEndPoint;
        pj_bool_t bThreadQuit;
        pj_thread_t* pSipThread[2];

        /* Accounts: */
        unsigned nAccountCount;
        SipAccount *Accounts;

        /* Calls: */
        unsigned nCallCount;
        SipCall *Calls;

        /* Mq */
        MessageQueue *pMq;
        void (*OnRegStateChange)(IN const int nAccountId, IN const SipAnswerCode Code, IN const void *pUser);
        void (*OnCallStateChange)(IN const int nCallId,IN const int nAccountId, IN const SipInviteState State, IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void* pMedia);
        SipAnswerCode (*OnIncomingCall)(IN const int nAccountId, IN const int nCallId, IN const char *From, IN const void *pUser, IN const void* pMedia);

};

/**
 * Event IDs.
 */
typedef enum SipEventType {

        SIP_REG_ACCOUNT,
        SIP_MAKE_CALL,
        SIP_HANGUP,
        SIP_HANGUP_BY_ACCOUNT,
        SIP_HANGUP_ALL,
        SIP_ANSWER_CALL,
} SipEventType;

struct SipInstanceConfig;
struct SipAccountConfig;
typedef struct SipEvent
{
        SipEventType Type;

        union
        {
                /* Create Instance */
                SipInstanceConfig *pInstanceConfig;
                /* Add Account */
                SipAccountConfig *pAccConfig;
                int nAccountId;
                /* Reg Account */
                struct {
                        int nAccountId;
                        int Reg;
                } Reg;
                /* Make Call */
                struct {
                        int nAccountId;
                        char *pDestUri;
                        void *pMedia;
                        int nCallId;
                } MakeCall;
                struct {
                        int nCallId;
                        int StatusCode;
                        char *Reason;
                        void *pMedia;
                } AnswerCall;
                /* Hang up */
                struct {
                        int nCallId;
                } HangUp;
                /* Hang Up by Account */
                struct {
                        int nAccount;
                } HangUpByAcc;
                /* Hang Up All */

                /*Destroy */

        } Body;
} SipEvent;

extern pjsip_module SipMod;
/* Callback to be called to handle incoming requests outside dialogs: */
pj_bool_t onRxRequest( IN pjsip_rx_data *pRxData );

/* Callback to be called when SDP negotiation is done in the call: */
void onSipCallOnMediaUpdate(IN pjsip_inv_session *pInviteSession,
                                   IN pj_status_t nStatus);

/* Callback to be called when invite session's state has changed: */
void onSipCallOnStateChanged( IN pjsip_inv_session *pInviteSession,
                                     IN pjsip_event *pEvent);

/* Callback to be called when dialog has forked: */
void onSipCallOnForked(IN pjsip_inv_session *pInviteSession, IN pjsip_event *pEvent);

int SipGetFreeCallId();
void RegisterToPjLib();
SIP_ERROR_CODE OnSipAddNewAccount(IN const SipAccountConfig *_pConfig, OUT int *_pAccountId);
SIP_ERROR_CODE OnSipDeleteAccount(IN const int _nAccountId);
SIP_ERROR_CODE OnSipRegAccount(IN const SipEvent *pEvent);
SIP_ERROR_CODE OnSipMakeNewCall(IN const SipEvent *pEvent);
SIP_ERROR_CODE OnSipAnswerCall(IN const SipEvent *pEvent);
SIP_ERROR_CODE OnSipHangUp(IN const SipEvent *pEvent);
SIP_ERROR_CODE OnSipHangUpAll(IN const SipEvent *pEvent);
SIP_ERROR_CODE OnSipHangUpByAccountId(IN const SipEvent *pEvent);

#endif
