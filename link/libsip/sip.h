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
         * @return answercode, refer SipAnswercode
         */
        SipAnswerCode (*OnIncomingCall)(IN const int nAccountId, IN const int nCallId, IN const char *pFrom, IN const void *pUser);

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
        void (*OnCallStateChange)(IN const int nCallId, IN const int nAccountId, IN const SipInviteState State, IN const SipAnswerCode StatusCode, IN const void *pUser);

} SipCallBack;

/**
 * Initialize sip instance
 *
 * @param pConfig, configuration for sip instance
 *
 * @return 1 on success, 0 for failed
 */
int SipCreateInstance(IN const SipCallBack *pCallBack);

/**
 * Add new account
 *
 * @param pUserName, set the username for authentication. same as username in From header
 * @param pPassoWord, set the plain-text passowrd for authentication
 * @param PDomain, registar server
 * @param RegWhenAdd, whether register when user added or not
 *
 * @return account id, -1 for error
 *
 */
int SipAddNewAccount(IN const char *pUserName, IN const char *pPassWord, IN const char *pDomain, IN void *pUserData);

/**
 * Add new account
 *
 * @param account id
 *
 */
void SipDeleteAccount(IN const int nAccountId);

/**
 * Registar the user
 *
 * @param nAccountId, account id returned by add_account
 * @param DeReg, reg or de-reg
 *
 * @return 1 on success, -1 for failed
 */
int SipRegAccount(IN const int nAccountId, IN const int bDeReg);

/**
 * Make a new call
 * @param nFromAccountid, The account id for caller
 * @param pDestUri, callee sip uri, like sip:1003@host, for tcp case should be sip:1003@host;transport=tcp
 *
 * @return call id, -1 for failed
 **/
int SipMakeNewCall(IN const int nFromAccountId, IN const char *pDestUri);

/**
 * Hangup call
 * @param nCallId, hangup call id
 *
 */
void  SipHangUp(IN const int nCallId);

/**
 * Hangup all calls
 *
 **/
void SipHangUpAll();

void SipHangUpByAccountId(int nAccountId);

/**
 * Destroy sip instance
 */
void SipDestroyInstance();

/**
 * Answer the call
 * @param nCallId
 * @param StatusCode, see SipAnswerCode
 * @param AnswerReason
 **/
int SipAnswerCall(IN const int nCallId, IN const SipAnswerCode StatusCode, IN const char *pReason);
#endif
