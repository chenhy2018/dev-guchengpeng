#ifndef __SIP_INTERNAL_H__
#define __SIP_INTERNAL_H__

#define SIP_MAX_CRED 1
#define SIP_PORT 5060

#define SIP_REG_INTERVAL 300 // 300 seconds
#define SIP_REG_RETRY_INTERNAL 300 // 300 seconds
#define SIP_REG_DELAY_BEFORE_REFRESH  5 // 5 seconds
#define SIP_UNREG_TIMEOUT 4000 // 4 seconds
#define SIP_KEEP_ALIVE_INTERVAL 15 // 15 seconds
#define SIP_SESSION_EXPIRES 600 // 10 minutes
#define SIP_MIN_SE 90 //90 seconds

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
        pj_thread_t* pSipThread[1];

        /* Accounts: */
        unsigned nAccountCount;
        SipAccount *Accounts;

        /* Calls: */
        unsigned nCallCount;
        SipCall *Calls;


        void (*OnRegStateChange)(IN const int nAccountId, IN const SipAnswerCode Code, IN const void *pUser);
        void (*OnCallStateChange)(IN const int nCallId,IN const int nAccountId, IN const SipInviteState State, IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void* pMedia);
        SipAnswerCode (*OnIncomingCall)(IN const int nAccountId, IN const int nCallId, IN const char *From, IN const void *pUser, IN const void* pMedia);

};
extern struct SipData sipData;
#endif
