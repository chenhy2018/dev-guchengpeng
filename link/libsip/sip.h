#ifndef __SIP_H__
#define __SIP_H__

#ifndef IN
#define IN
#endif

#ifndef OUT
#define OUT
#endif

#ifndef INOUT
#define INOUT
#endif

/**
 * This enumeration lists standard SIP status codes according to RFC 3261.
 * for more info see https://en.wikipedia.org/wiki/List_of_SIP_response_codes
 */
typedef enum SipAnswerCode
{
        SIP_SUCCESS = 0,

        /* Normal Sip answer code */
        TRYING = 100,
        RINGING = 180,
        CALL_BEING_FORWARDED = 181,
        QUEUED = 182,
        PROGRESS = 183,

        OK = 200,
        ACCEPTED = 202,

        MULTIPLE_CHOICES = 300,
        MOVED_PERMANENTLY = 301,
        MOVED_TEMPORARILY = 302,
        USE_PROXY = 305,
        ALTERNATIVE_SERVICE = 380,

        BAD_REQUEST = 400,
        UNAUTHORIZED = 401,
        PAYMENT_REQUIRED = 402,
        FORBIDDEN = 403,
        NOT_FOUND = 404,
        METHOD_NOT_ALLOWED = 405,
        NOT_ACCEPTABLE = 406,
        PROXY_AUTHENTICATION_REQUIRED = 407,
        REQUEST_TIMEOUT = 408,
        GONE = 410,
        REQUEST_ENTITY_TOO_LARGE = 413,
        REQUEST_URI_TOO_LONG = 414,
        UNSUPPORTED_MEDIA_TYPE = 415,
        UNSUPPORTED_URI_SCHEME = 416,
        BAD_EXTENSION = 420,
        EXTENSION_REQUIRED = 421,
        SESSION_TIMER_TOO_SMALL = 422,
        INTERVAL_TOO_BRIEF = 423,
        TEMPORARILY_UNAVAILABLE = 480,
        CALL_TSX_DOES_NOT_EXIST = 481,
        LOOP_DETECTED = 482,
        TOO_MANY_HOPS = 483,
        ADDRESS_INCOMPLETE = 484,
        AMBIGUOUS = 485,
        BUSY_HERE = 486,
        REQUEST_TERMINATED = 487,
        NOT_ACCEPTABLE_HERE = 488,
        BAD_EVENT = 489,
        REQUEST_UPDATED = 490,
        REQUEST_PENDING = 491,
        UNDECIPHERABLE = 493,

        INTERNAL_SERVER_ERROR = 500,
        NOT_IMPLEMENTED = 501,
        BAD_GATEWAY = 502,
        SERVICE_UNAVAILABLE = 503,
        SERVER_TIMEOUT = 504,
        VERSION_NOT_SUPPORTED = 505,
        MESSAGE_TOO_LARGE = 513,
        PRECONDITION_FAILURE = 580,

        BUSY_EVERYWHERE = 600,
        DECLINE = 603,
        DOES_NOT_EXIST_ANYWHERE = 604,
        NOT_ACCEPTABLE_ANYWHERE = 606,

        /* for user function(SipRegAccount/SipMakeNewCall etc), after user call it
           we just send it to our messageque below answer code is error when we handle it */
        SIP_INVALID_ARG = 701,
        SIP_MALLOC_FAILED,

        /* SipCreateInstance Error Code */
        SIP_PJ_INIT_FAILED,
        SIP_CREATE_ENDPOINT_FALIED,
        SIP_START_TP_FAILED,
        SIP_INIT_TRANS_FAILED,
        SIP_UA_LAYER_INIT_FAILED,
        SIP_INIT_INV_SESS_FALIED,
        SIP_INIT_100_REL_FALIED,
        SIP_INIT_SESS_TIMER_FAILED,
        SIP_REG_INCOMING_FAILED,
        SIP_REG_LOG_FAILED,
        SIP_CREATE_MQ_THREAD_FAILED,

        /* SipAddNewAccount Error Code */
        SIP_TOO_MANY_ACCOUNT,
        SIP_USR_ALREADY_EXIST,

        /* SipRegAccount Error Code */
        SIP_CREATE_REG_FAILED,
        SIP_USR_NOT_REGISTERED,
        SIP_SEND_REG_FAILED,


        /* SipMakeNewCall Error Code */
        SIP_CREATE_DLG_FAILED,
        SIP_CREATE_INV_SESS_FAILED,
        SIP_SESS_TIMER_INIT_FALIED,
        SIP_CREATE_INV_REQ_FAILED,
        SIP_SNED_INV_REQ_FAILED,
        SIP_TOO_MANY_CALLS_FOR_INSTANCE,
        SIP_TOO_MANY_CALLS_FOR_ACCOUNT,

        /* SipAnswerCall Error Code */
        SIP_CREATE_RES_FAILED,
        SIP_SEND_RES_FAILED,
} SipAnswerCode;


/**
 * This enumeration describes invite session state.
 */
typedef enum SipInviteState
{
        INV_STATE_NULL,    /**< Before INVITE is sent or received  */
        INV_STATE_CALLING,    /**< After INVITE is sent    */
        INV_STATE_INCOMING,    /**< After INVITE is received.    */
        INV_STATE_EARLY,    /**< After response with To tag.    */
        INV_STATE_CONNECTING,    /**< After 2xx is sent/received.    */
        INV_STATE_CONFIRMED,    /**< After ACK is sent/received.    */
        INV_STATE_DISCONNECTED,   /**< Session is terminated.    */
} SipInviteState;

/**
 * callback function for various pjlib event notification from pjsip stack
 *all of these callback are OPTIONAL
 *
 */
typedef struct SipCallBack
{
        /**
         * Notify application on incoming call
         *
         * @param nAccountid, the account id for incoming call
         * @param nCallId, call id for incoming call
         * @param pFrom
         * @param pUser(user data add in add account)
         * @param pMedia, sdp info
         * @param sdk call Id 
         * @return answercode, refer SipAnswercode
         */
        SipAnswerCode (*OnIncomingCall)(IN const int nAccountId, IN const char *pFrom, IN const void *pUser, IN const void *pMedia, OUT int *pCallId);

        /**
         * Notify when registration status has changed
         *
         * @param nAccountId
         * @param reg_st_code, registration status code
         *
         */
        void (*OnRegStatusChange)(IN const int nAccountId, IN const SipAnswerCode RegStatusCode, IN const void *pUser);

        /**
         * Notify application when call state has changed
         * @param CallId
         * @param nAccountId, Call associated accountId
         * @param state state of this call see inv_state
         *
         */
        void (*OnCallStateChange)(IN const int nCallId, IN const int nAccountId, IN const SipInviteState State, IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void *pMedia);

} SipCallBack;

typedef struct SipInstanceConfig
{
        unsigned nMaxCall;
        unsigned nMaxAccount;
        SipCallBack Cb;
}SipInstanceConfig;

typedef struct SipAccountConfig
{
        char *pUserName;
        char *pPassWord;
        char *pDomain;
        void *pUserData;

        int nMaxOngoingCall;
} SipAccountConfig;

/**
 * Initialize sip instance
 *
 * @param pConfig, see SipInstanceconfig
 * @return see #SipAnswerCode
 */
SipAnswerCode SipCreateInstance(IN const SipInstanceConfig *pConfig);

/**
 * Registar the user
 *
 * @param nAccountId, account id returned by add_account
 * @param DeReg, reg or de-reg
 *
 * @return see #SipAnswerCode
 */

SipAnswerCode SipRegAccount(IN const SipAccountConfig *pConfig, IN const int nAccountId);

/**
 * UnReg Account
 *
 * @param account id
 *
 */
SipAnswerCode SipUnRegAccount(IN const int nAccountId);


/**
 * Make a new call
 * @param nFromAccountid, The account id for caller
 * @param pDestUri, callee sip uri, like sip:1003@host, for tcp case should be sip:1003@host;transport=tcp
 * @param pCallId
 * @return see #SipAnswerCode
 **/
SipAnswerCode SipMakeNewCall(IN const int nFromAccountId, IN const char *pDestUri, IN const void *pMedia, IN const int npCallId);

/**
 * Hangup call
 * @param nCallId, hangup call id
 *
 */
SipAnswerCode SipHangUp(IN const int nCallId);

/**
 * Hangup all calls
 *
 **/
SipAnswerCode SipHangUpAll();

SipAnswerCode SipHangUpByAccountId(int nAccountId);

/**
 * Destroy sip instance
 */
SipAnswerCode SipDestroyInstance();

/**
 * Answer the call
 * @param nCallId
 * @param StatusCode, see SipAnswerCode
 * @param AnswerReason
 * @return see #SipAnswerCode
 **/
SipAnswerCode SipAnswerCall(IN const int nCallId, IN const SipAnswerCode StatusCode, IN const char *pReason, IN const void* pMedia);


/**
 * set log level the maximum level of verbosity of the logging messages (6=very detailed..1=error only, 0=disabled)
 */

void SipSetLogLevel(IN const int nLevel);
/**
 * !!!! for test offer/answer sdp
 * create tmp sdp
 */
int CreateTmpSDP(OUT void **pSdp);
void PrintSdp(IN const void *pSdp);
#endif
