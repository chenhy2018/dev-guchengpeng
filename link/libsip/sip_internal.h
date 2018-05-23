#ifndef __SIP_INTERNAL_H__
#define __SIP_INTERNAL_H__

#define SIP_MAX_CALLS 10
#define SIP_MAX_ACC 10
#define SIP_MAX_CRED 1
#define SIP_PORT 5060

#define SIP_REG_INTERVAL 300 // 300 seconds
#define SIP_REG_RETRY_INTERNAL 300 // 300 seconds
#define SIP_REG_DELAY_BEFORE_REFRESH  5 // 5 seconds
#define SIP_UNREG_TIMEOUT 4000 // 4 seconds
#define SIP_KEEP_ALIVE_INTERVAL 15 // 15 seconds

/**
 * This structure describes SIP account connfigure to be specified when
 * adding a new acount with #AddNewAccount().
 */
typedef struct SipAccountConfig {
        pj_str_t Id;
        pj_str_t RegUri;
        pj_str_t UserName;
        pj_str_t SipDomain;
        pj_str_t Password;
        pj_str_t Scheme;
        pj_str_t Realm;
        pj_str_t KaData;

        unsigned    nRegTimeout;
        unsigned    nUnRegTimeout;
        unsigned    nRegDelayBeforeRefresh;
        unsigned    nRegRetryInterval;
        unsigned    nKaInterval;
} SipAccountConfig;

/**
 * Account info
 */
typedef struct SipAccount {
        pj_pool_t *pPool;
        SipAccountConfig Config;
        int nIndex;
        pj_bool_t bValid;
        pjsip_regc *pRegc;
        pj_str_t Contact;
        int nLastRegCode;

        unsigned nCredCnt;
        pjsip_cred_info  Cred[SIP_MAX_CRED];
        pjsip_host_port ViaAddr;
        /* Add Keepalive timer for this account */
        pj_timer_entry   KaTimer;
        pjsip_transport *KaTransport;
        pj_sockaddr     KaTarget;
        unsigned     KaTargetLen;
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

        pjsip_endpoint*pSipEndPoint;
        pj_bool_t bThreadQuit;
        pj_thread_t* pSipThread[1];

        /* Accounts: */
        unsigned nAccountCount;
        SipAccount Accounts[SIP_MAX_ACC];

        /* Calls: */
        unsigned nCallCount;
        SipCall Calls[SIP_MAX_CALLS];


        void (*OnRegStateChange)(int nAccountId, SipAnswerCode Code);
        void (*OnCallStateChange)(IN const int nCallId, IN const SipInviteState State, IN const SipAnswerCode StatusCode);
        SipAnswerCode (*OnIncomingCall)(IN const int nAccountId, IN const int nCallId, IN const char *From);

};
extern struct SipData sipData;
#endif
