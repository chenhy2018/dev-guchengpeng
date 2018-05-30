#include <pjsip.h>
#include <pjmedia.h>
#include <pjsip_ua.h>
#include <pjsip_simple.h>
#include <pjlib-util.h>
#include <pjlib.h>

#include <stdlib.h>
#include <stdio.h>
#include "sip.h"
#include "sip_internal.h"


#define THIS_FILE "sip.c"

void PrintErrorMsg(IN const pj_status_t _Status, IN const char *_pMsg)
{
        char errmsg[PJ_ERR_MSG_SIZE];
        pj_strerror(_Status, errmsg, sizeof(errmsg));
        PJ_LOG(1, (THIS_FILE, "%s: %s [Status=%d]", _pMsg, errmsg, _Status));
}

struct SipData SipAppData;

/* Init Regc Data */
static pj_status_t SipRegcInit(IN const int _nAccountId);

/* Callback for regc_cb */
static void onSipRegc(IN struct pjsip_regc_cbparam *param);

/* CallBack for regc_tsx_cb */
static void onSipRegcTsx(IN struct pjsip_regc_tsx_cb_param *param);

/* periodically call pjsip_endpt_handle_events to ensure that all events from both
 *transports and timer heap are handled in timely manner
 */
static int PullForEndPointEvent(IN void *_arg);

static void UpdateKeepAlive(INOUT SipAccount *_pAccount, IN const pj_bool_t _Start, IN const struct pjsip_regc_cbparam *pCbData);

static void KeepAliveTimerCallBack(IN pj_timer_heap_t *_pTimerHeap, IN pj_timer_entry *_pTimerEntry);

/* Callback to be called to handle incoming requests outside dialogs: */
static pj_bool_t onRxRequest( IN pjsip_rx_data *pRxData );

/* Callback to be called when SDP negotiation is done in the call: */
static void onSipCallOnMediaUpdate(IN pjsip_inv_session *pInviteSession,
                                  IN pj_status_t nStatus);

/* Callback to be called when invite session's state has changed: */
static void onSipCallOnStateChanged( IN pjsip_inv_session *pInviteSession,
                                   IN pjsip_event *pEvent);

/* Callback to be called when dialog has forked: */
static void onSipCallOnForked(IN pjsip_inv_session *pInviteSession, IN pjsip_event *pEvent);

/* Get local AccountId from incoming message */
static int SipGetAccountIdFromRxData(IN const pjsip_rx_data *_pRxData);

static pj_bool_t SipUpdateContactIfNat(IN SipAccount *_pAccount, IN struct pjsip_regc_cbparam *_pCbData);

static int CreateTmpSDP(pj_pool_t *pPool, SipCall *pCall, pjmedia_sdp_session **pSdp);

/* This is a PJSIP module to be registered by application to handle
 * incoming requests outside any dialogs/transactions. The main purpose
 * here is to handle incoming INVITE request message, where we will
 * create a dialog and INVITE session for it.
 */
static pjsip_module SipMod = {
        NULL, NULL,    /* prev, next.*/
        { "SipMod", 10},    /* Name.*/
        -1,    /* Id*/
        PJSIP_MOD_PRIORITY_APPLICATION, /* Priority*/
        NULL,    /* load()*/
        NULL,    /* start()*/
        NULL,    /* stop()*/
        NULL,    /* unload()*/
        &onRxRequest, /* on_rx_request()*/
        NULL,    /* on_rx_response()*/
        NULL,    /* on_tx_request.*/
        NULL,    /* on_tx_response()*/
        NULL,    /* on_tsx_state()*/
};
/* Notification on incoming messages */
static pj_bool_t LoggingOnRxMsg(pjsip_rx_data *_pRxData)
{
    PJ_LOG(4,(THIS_FILE, "RX %d bytes %s from %s %s:%d:\n"
			 "%.*s\n"
			 "--end msg--",
			 _pRxData->msg_info.len,
			 pjsip_rx_data_get_info(_pRxData),
			 _pRxData->tp_info.transport->type_name,
			 _pRxData->pkt_info.src_name,
			 _pRxData->pkt_info.src_port,
			 (int)_pRxData->msg_info.len,
			 _pRxData->msg_info.msg_buf));

    /* Always return false, otherwise messages will not get processed! */
    return PJ_FALSE;
}


/* Notification on outgoing messages */
static pj_status_t LoggingOnTxMsg(pjsip_tx_data *_pTxData)
{

    /* Important note:
     *	tp_info field is only valid after outgoing messages has passed
     *	transport layer. So don't try to access tp_info when the module
     *	has lower priority than transport layer.
     */

    PJ_LOG(4,(THIS_FILE, "TX %d bytes %s to %s %s:%d:\n"
			 "%.*s\n"
			 "--end msg--",
			 (_pTxData->buf.cur - _pTxData->buf.start),
			 pjsip_tx_data_get_info(_pTxData),
			 _pTxData->tp_info.transport->type_name,
			 _pTxData->tp_info.dst_name,
			 _pTxData->tp_info.dst_port,
			 (int)(_pTxData->buf.cur - _pTxData->buf.start),
			 _pTxData->buf.start));

    /* Always return success, otherwise message will not get sent! */
    return PJ_SUCCESS;
}

/* The module instance. */
static pjsip_module SipLogger =
{
    NULL, NULL,				/* prev, next.		*/
    { "SipLogger", 11},		/* Name.		*/
    -1,					/* Id			*/
    PJSIP_MOD_PRIORITY_TRANSPORT_LAYER-1,/* Priority	        */
    NULL,				/* load()		*/
    NULL,				/* start()		*/
    NULL,				/* stop()		*/
    NULL,				/* unload()		*/
    &LoggingOnRxMsg,			/* on_rx_request()	*/
    &LoggingOnRxMsg,			/* on_rx_response()	*/
    &LoggingOnTxMsg,			/* on_tx_request.	*/
    &LoggingOnTxMsg,			/* on_tx_response()	*/
    NULL,				/* on_tsx_state()	*/

};


int SipCreateInstance(IN const SipCallBack *_pSipCallBack)
{
        pj_log_set_level(6);
        pj_status_t Status;
        Status = pj_init();
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        Status = pjlib_util_init();
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* create pool factory before allow memory */
        pj_caching_pool_init(&SipAppData.Cp, &pj_pool_factory_default_policy, 0);
        SipAppData.pPool = pj_pool_create(&SipAppData.Cp.factory, "SipApp", 1000, 1000, NULL);

        /* Create mutex */
        Status = pj_mutex_create_recursive(SipAppData.pPool, "SipApp", &SipAppData.pMutex);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Init global sip endpoint */
        Status = pjsip_endpt_create(&SipAppData.Cp.factory, pj_gethostname()->ptr,
                                    &SipAppData.pSipEndPoint);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* start udp socket on sip port */
        pj_sockaddr Address;
        pjsip_transport *tp;
        pj_sockaddr_init(pj_AF_INET(), &Address, NULL, 0);
        Status = pjsip_udp_transport_start(SipAppData.pSipEndPoint, &Address.ipv4, NULL, 1, &tp);
        SipAppData.LocalPort = tp->local_name.port;
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        PJ_LOG(3,(THIS_FILE, "SIP UDP listening on %.*s:%d",
                  (int)tp->local_name.host.slen, tp->local_name.host.ptr,
                  tp->local_name.port));

        /* Init transaction layer */
        Status = pjsip_tsx_layer_init_module(SipAppData.pSipEndPoint);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Init UA layer module */
        Status = pjsip_ua_init_module(SipAppData.pSipEndPoint, NULL);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Add Invite  session module */
        pjsip_inv_callback InviteCallBack;

        /* Init the callback for INVITE session: */
        pj_bzero(&InviteCallBack, sizeof(InviteCallBack));
        InviteCallBack.on_state_changed = &onSipCallOnStateChanged;
        InviteCallBack.on_new_session = &onSipCallOnForked;
        InviteCallBack.on_media_update = &onSipCallOnMediaUpdate;

        /* Initialize invite session module:  */
        Status = pjsip_inv_usage_init(SipAppData.pSipEndPoint, &InviteCallBack);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Initialize 100rel support */
        Status = pjsip_100rel_init_module(SipAppData.pSipEndPoint);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /*  Register our module to receive incoming requests */
        Status = pjsip_endpt_register_module(SipAppData.pSipEndPoint, &SipMod);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Register log module */
        Status = pjsip_endpt_register_module(SipAppData.pSipEndPoint, &SipLogger);
        PJ_ASSERT_RETURN(Status == PJ_SUCCESS, Status);

        /* Add callback */
        SipAppData.OnRegStateChange = _pSipCallBack->OnRegStatusChange;
        SipAppData.OnCallStateChange = _pSipCallBack->OnCallStateChange;
        SipAppData.OnIncomingCall = _pSipCallBack->OnIncomingCall;

        /* Init Accounts */
        int i;
        for(i = 0; i < SIP_MAX_ACC; ++i) {
                SipAppData.Accounts[i].nIndex = i;
                SipAppData.Accounts[i].bValid = PJ_FALSE;
                SipAppData.Accounts[i].nCredCnt = 1;
        }
        /* Init calls */
        for(i = 0; i < SIP_MAX_CALLS; ++i) {
                SipAppData.Calls[i].nIndex = i;
                SipAppData.Calls[i].bValid = PJ_FALSE;
                SipAppData.Calls[i].nAccountId = -1;
        }

        /* Create a thraed to pull pjsip endpoint transport event*/
        pj_thread_create(SipAppData.pPool, "SipWorkThread", &PullForEndPointEvent, NULL, 0, 0, &SipAppData.pSipThread[0]);
        return 1;
}

void SipDestroyInstance()
{
        if (SipAppData.pSipEndPoint) {
                PJ_LOG(3, (THIS_FILE, "Destroy libSip instance ..."));
        }

        /* Hangup all calls */
        SipHangUpAll();

        /* offline all account */
        int i;
        for (i = 0; i < SIP_MAX_ACC; ++i) {
                if (SipAppData.Accounts[i].bValid) {
                        SipDeleteAccount(i);
                }
        }
        /* Sleep sometime to wait de-register complete */
        PJ_LOG(4, (THIS_FILE, "Destroying ...."));
        pj_thread_sleep(5000);

        /* Stop working thread */
        SipAppData.bThreadQuit = PJ_TRUE;
        if (SipAppData.pSipThread[0]) {
                pj_thread_join(SipAppData.pSipThread[0]);
                pj_thread_destroy(SipAppData.pSipThread[0]);
                SipAppData.pSipThread[0] = NULL;
        }
        PJ_LOG(4, (THIS_FILE, "Working thread has destroyed"));
        pjsip_endpt_destroy(SipAppData.pSipEndPoint);
        SipAppData.pSipEndPoint = NULL;

        /* Destroy mutex */
        if (SipAppData.pMutex) {
                pj_mutex_destroy(SipAppData.pMutex);
                SipAppData.pMutex = NULL;
        }
        /* Destroy mem pool */
        if (SipAppData.pPool) {
                pj_pool_release(SipAppData.pPool);
                SipAppData.pPool = NULL;
                pj_caching_pool_destroy(&SipAppData.Cp);
        }

        PJ_LOG(3, (THIS_FILE, "LibSip destroyed..."));

        pj_shutdown();

        pj_bzero(&SipAppData, sizeof(SipAppData));
}
int SipAddNewAccount(IN const char *_pUserName, IN const char *_pPassWord, IN const char *_pDomain)
{
        /* Input check */
        PJ_ASSERT_RETURN(_pUserName && _pPassWord && _pDomain, PJ_EINVAL);

        /* Account amount check */
        PJ_ASSERT_RETURN(SipAppData.nAccountCount < PJ_ARRAY_SIZE(SipAppData.Accounts), PJ_ETOOMANY);

        pj_mutex_lock(SipAppData.pMutex);
        /* Find empty account id. */
        int id;
        for (id=0; id < PJ_ARRAY_SIZE(SipAppData.Accounts); ++id) {
                if (SipAppData.Accounts[id].bValid == PJ_FALSE)
                        break;
        }
        /* Expect to find a slot */
        PJ_ASSERT_ON_FAIL(id < PJ_ARRAY_SIZE(SipAppData.Accounts),
                          {pj_mutex_unlock(SipAppData.pMutex); return -1;});

        SipAccount *pAccount = &SipAppData.Accounts[id];

        if (pAccount->pPool)
                pj_pool_reset(pAccount->pPool);
        else
                pAccount->pPool = pj_pool_create(&SipAppData.Cp.factory, "SipAcc%p", 512, 256, NULL);

        /* Setup time */
        pAccount->Config.nRegDelayBeforeRefresh = SIP_REG_DELAY_BEFORE_REFRESH;
        pAccount->Config.nRegRetryInterval = SIP_REG_RETRY_INTERNAL;
        pAccount->Config.nRegTimeout = SIP_REG_INTERVAL;
        pAccount->Config.nUnRegTimeout = SIP_UNREG_TIMEOUT;
        pAccount->Config.nKaInterval = SIP_KEEP_ALIVE_INTERVAL;

        /* Copy account info */
        char ID[80], Registrar[80];
        sprintf(ID, "sip:%s@%s", _pUserName, _pDomain);
        sprintf(Registrar, "sip:%s", _pDomain);
        pj_str_t PJID = pj_str(ID);
        pj_str_t PJReg = pj_str(Registrar);
        pj_str_t PJUserName = pj_str((char *)_pUserName);
        pj_str_t PJPassword = pj_str((char *)_pPassWord);
        pj_str_t PJRealm = pj_str("*");
        pj_str_t PJScheme = pj_str("digest");
        pj_str_t PJDomain = pj_str((char *)_pDomain);
        pj_str_t PJKeepAliveData = pj_str("\r\n");

        pj_strdup_with_null(pAccount->pPool, &pAccount->Config.RegUri, &PJReg);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Config.KaData, &PJKeepAliveData);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Config.Id, &PJID);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Config.SipDomain, &PJDomain);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Config.UserName, &PJUserName);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].username, &PJUserName);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].data, &PJPassword);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].realm, &PJRealm);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].scheme, &PJScheme);
        pAccount->Cred[0].data_type = 0;

        SipAppData.nAccountCount++;
        SipAppData.Accounts[id].bValid = PJ_TRUE;

        pj_mutex_unlock(SipAppData.pMutex);

        return id;
}

void SipDeleteAccount(IN const int _nAccountId)
{
        SipAccount *pAccount;

        PJ_LOG(4,(THIS_FILE, "Deleting account %d..", _nAccountId));
        //pj_mutex_lock(SipAppData.pMutex);
        pAccount = &SipAppData.Accounts[_nAccountId];

        /* Cancel keep alive timer */
        if (pAccount->KaTimer.id) {
                pjsip_endpt_cancel_timer(SipAppData.pSipEndPoint, &pAccount->KaTimer);
                pAccount->KaTimer.id = PJ_FALSE;
        }
        if (pAccount->KaTransport) {
                pjsip_transport_dec_ref(pAccount->KaTransport);
                pAccount->KaTransport = NULL;
        }

        /* Offline account */
        if (pAccount->pRegc) {
                SipRegAccount(_nAccountId, 0);
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
        }

        if (pAccount->pPool) {
                pj_pool_release(pAccount->pPool);
                pAccount->pPool = NULL;
        }
        pAccount->bValid = PJ_FALSE;
        pAccount->Contact.slen = 0;
        //pj_mutex_unlock(SipAppData.pMutex);
}

int SipRegAccount(IN const int _nAccountId, IN const int _bDeReg)
{
        SipAccount *pAccount;
        pj_status_t Status = 0;
        pjsip_tx_data *pTransData = 0;

        PJ_ASSERT_RETURN(_nAccountId >= 0 && _nAccountId < (int)PJ_ARRAY_SIZE(SipAppData.Accounts),
                         PJ_EINVAL);
        PJ_ASSERT_RETURN(SipAppData.Accounts[_nAccountId].bValid, PJ_EINVALIDOP);

        PJ_LOG(4,(THIS_FILE, "Acc %d: setting %sregistration..",
                  _nAccountId, (_bDeReg? "" : "un")));
        pj_mutex_lock(SipAppData.pMutex);

        pAccount = &SipAppData.Accounts[_nAccountId];

        /* For initial register */
        if (_bDeReg) {
                if (pAccount->pRegc == NULL) {
                        Status = SipRegcInit(_nAccountId);
                        if (Status != PJ_SUCCESS) {
                                PrintErrorMsg(Status, "Unable to create registration, Status");
                                pj_mutex_unlock(SipAppData.pMutex);
                                return -1;
                        }
                }
                if (!pAccount->pRegc) {
                        pj_mutex_unlock(SipAppData.pMutex);
                        return -1;
                }

                /* Create register request message */
                Status = pjsip_regc_register(pAccount->pRegc, 1, &pTransData);
                /* Add authorization header */
                if (Status == PJ_SUCCESS && pAccount->nCredCnt) {
                        PJ_LOG(3, (THIS_FILE, "Add Credentials info"));
                        pjsip_authorization_hdr *pAuthHeader;
                        pAuthHeader = pjsip_authorization_hdr_create(pTransData->pool);
                        pAuthHeader->scheme = pj_str("Digest");
                        pAuthHeader->credential.digest.username = pAccount->Cred[0].username;
                        pAuthHeader->credential.digest.realm = pAccount->Config.SipDomain;
                        pAuthHeader->credential.digest.uri = pAccount->Config.RegUri;
                        pAuthHeader->credential.digest.algorithm = pj_str("md5");

                        pjsip_msg_add_hdr(pTransData->msg, (pjsip_hdr*)pAuthHeader);

                }
        } else {
                if (pAccount->pRegc == NULL) {
                        PJ_LOG(3, (THIS_FILE, "Currently not registered"));
                        pj_mutex_unlock(SipAppData.pMutex);
                        return 1;
                }
                Status = pjsip_regc_unregister(pAccount->pRegc, &pTransData);
        }
        if (Status == PJ_SUCCESS) {
                Status = pjsip_regc_send(pAccount->pRegc, pTransData);
        }

        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Unable to create/send register");
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }
        PJ_LOG(4,(THIS_FILE, "Acc %d: %s sent", _nAccountId,
                  (_bDeReg ? "Registration" : "Unregistration")));

        /* Invoke callback function */
        pj_mutex_unlock(SipAppData.pMutex);
        return 1;
}

static int PullForEndPointEvent(void *_arg)
{
        PJ_UNUSED_ARG(_arg);
        while (!SipAppData.bThreadQuit) {
                pj_time_val timeout = {0, 10};
                pjsip_endpt_handle_events(SipAppData.pSipEndPoint, &timeout);
        }
        return 0;
}

static pj_status_t SipRegcInit(IN const int _nAccountId)
{
        SipAccount *pAccount;
        pj_pool_t *pPool;
        pj_status_t Status;

        pAccount = &SipAppData.Accounts[_nAccountId];

        /* Destroy exist session if any */
        if (pAccount->pRegc) {
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
        }

        /* Create Regc Data */
        Status = pjsip_regc_create(SipAppData.pSipEndPoint, pAccount, &onSipRegc, &pAccount->pRegc);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Unable to create client registration");
                return Status;
        }
        /* Create concat */
        char tmpContact[64];
        /* Get local IP address for the default IP address */
        {
                static char LocalIp[PJ_INET_ADDRSTRLEN];
                const pj_str_t *pHostName;
                pj_sockaddr_in tmpAddr;

                pHostName = pj_gethostname();
                pj_sockaddr_in_init(&tmpAddr, pHostName, 0);
                pj_inet_ntop(pj_AF_INET(), &tmpAddr.sin_addr, LocalIp,
                             sizeof(LocalIp));
                SipAppData.LocalIp = pj_str(LocalIp);
                pj_ansi_sprintf(tmpContact, "<sip:%s@%s:%d>", pAccount->Config.UserName.ptr, LocalIp, SipAppData.LocalPort);
        }
        pj_str_t Contact = pj_str(tmpContact);
        pj_strdup(pAccount->pPool, &pAccount->Contact, &Contact);
        Status = pjsip_regc_init(pAccount->pRegc,
                                 &pAccount->Config.RegUri,
                                 &pAccount->Config.Id,
                                 &pAccount->Config.Id,
                                 1,
                                 &pAccount->Contact,
                                 pAccount->Config.nRegTimeout);
        if (Status != PJ_SUCCESS) {
                PJ_LOG(3, (THIS_FILE, "Client registration initialization error", Status));
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
                pAccount->Contact.slen = 0;
                return Status;
        }

        pjsip_regc_set_reg_tsx_cb(pAccount->pRegc, onSipRegcTsx);

        /* Set credentials */
        pjsip_regc_set_credentials(pAccount->pRegc, pAccount->nCredCnt, pAccount->Cred);

        pjsip_regc_set_delay_before_refresh(pAccount->pRegc, pAccount->Config.nRegDelayBeforeRefresh);

        return PJ_SUCCESS;
}


/*
 * This callback is called by pjsip_regc when outgoing register
 * request has completed.
 */
static void onSipRegc(IN struct pjsip_regc_cbparam *_pCbData)
{
        PJ_LOG(3, (THIS_FILE, "call onSipRegc"));
        SipAccount *pAccount = (SipAccount *)_pCbData->token;
        pj_mutex_lock(SipAppData.pMutex);

        /* Print registration Status */
        if (_pCbData->status != PJ_SUCCESS) {
                PJ_LOG(3, (THIS_FILE, "SIP registration error, status = %d", _pCbData->status));
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
                pAccount->Contact.slen = 0;
                /* Stop keep alive timer */
                UpdateKeepAlive(pAccount, PJ_FALSE, NULL);
        } else if (_pCbData->code < 0 || _pCbData->code >= 300) {
                PJ_LOG(2, (THIS_FILE, "SIP registration failed, status = %d (%.*s)",
                           _pCbData->code, (int)_pCbData->reason.slen, _pCbData->reason.ptr));
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
                pAccount->Contact.slen = 0;

                /* Stop keep alive timer */
                UpdateKeepAlive(pAccount, PJ_FALSE, NULL);
        } else if (PJSIP_IS_STATUS_IN_CLASS(_pCbData->code, 200)) {
                /* For unregistration */
                if (_pCbData->expiration < 1) {
                        pjsip_regc_destroy(pAccount->pRegc);
                        pAccount->pRegc = NULL;
                        pAccount->Contact.slen = 0;
                        /* Stop keep alive timer */
                        UpdateKeepAlive(pAccount, PJ_FALSE, NULL);
                        PJ_LOG(3, (THIS_FILE, "%s: un-registration success", pAccount->Config.Id.ptr));
                } else {
                        PJ_LOG(3, (THIS_FILE, "%s: registration success, status = %d(%.*s)"
                                   "will re-register in %d seconds", pAccount->Config.Id.ptr,
                                   _pCbData->code,
                                   (int)_pCbData->reason.slen, _pCbData->reason.ptr,
                                   _pCbData->expiration));
                        /* Start keep alive timer */
                        UpdateKeepAlive(pAccount, PJ_TRUE, _pCbData);
                }
        }

        pAccount->nLastRegCode = _pCbData->code;

        if (SipAppData.OnRegStateChange) {
                (*SipAppData.OnRegStateChange)(pAccount->nIndex, (SipAnswerCode)pAccount->nLastRegCode);
        }
        pj_mutex_unlock(SipAppData.pMutex);

        /* hangup call if re-registration attempt failed */
        if (_pCbData->code == PJSIP_SC_REQUEST_TIMEOUT ||
             _pCbData->code == PJSIP_SC_INTERNAL_SERVER_ERROR ||
             _pCbData->code == PJSIP_SC_BAD_GATEWAY ||
             _pCbData->code == PJSIP_SC_SERVICE_UNAVAILABLE ||
             _pCbData->code == PJSIP_SC_SERVER_TIMEOUT ||
             _pCbData->code == PJSIP_SC_TEMPORARILY_UNAVAILABLE ||
            PJSIP_IS_STATUS_IN_CLASS(_pCbData->code, 600)) {
                SipHangUpByAccountId(pAccount->nIndex);
        }
}
static void UpdateKeepAlive(INOUT SipAccount *_pAccount, IN const pj_bool_t _Start, IN const struct pjsip_regc_cbparam *_pCbData)
{
        PJ_LOG(3, (THIS_FILE, "call UpdateKeepAlive"));
        if (_pAccount->KaTimer.id) {
                pjsip_endpt_cancel_timer(SipAppData.pSipEndPoint, &_pAccount->KaTimer);
                _pAccount->KaTimer.id = PJ_FALSE;
                if (_pAccount->KaTransport) {
                        pjsip_transport_dec_ref(_pAccount->KaTransport);
                        _pAccount->KaTransport = NULL;
                }
        }

        if (_Start) {
                pj_time_val Delay;
                pj_status_t Status;

                /* return if ka is disable */
                if (_pAccount->Config.nKaInterval == 0)
                        return;
                _pAccount->KaTransport = _pCbData->rdata->tp_info.transport;
                pjsip_transport_add_ref(_pAccount->KaTransport);

                {
                        pjsip_transaction *tsx;
                        pjsip_tx_data *req;

                        tsx = pjsip_rdata_get_tsx(_pCbData->rdata);
                        PJ_ASSERT_ON_FAIL(tsx, return);

                        req = tsx->last_tx;

                        pj_memcpy(&_pAccount->KaTarget, &req->tp_info.dst_addr,
                                  req->tp_info.dst_addr_len);
                        _pAccount->KaTargetLen = req->tp_info.dst_addr_len;

                }
                /* Setup timer */
                _pAccount->KaTimer.cb = &KeepAliveTimerCallBack;
                _pAccount->KaTimer.user_data = (void*)_pAccount;

                Delay.sec = _pAccount->Config.nKaInterval;
                Delay.msec = 0;

                Status = pjsip_endpt_schedule_timer(SipAppData.pSipEndPoint, &_pAccount->KaTimer, &Delay);
                if (Status == PJ_SUCCESS) {
                        _pAccount->KaTimer.id = PJ_TRUE;
                        PJ_LOG(3,(THIS_FILE, "Keep-alive timer started for acc %d, "
                                  "destination:%s:%d, interval:%ds",
                                  _pAccount->nIndex,
                                  _pCbData->rdata->pkt_info.src_name,
                                  _pCbData->rdata->pkt_info.src_port,
                                  _pAccount->Config.nKaInterval));
                } else {
                        _pAccount->KaTimer.id = PJ_FALSE;
                        pjsip_transport_dec_ref(_pAccount->KaTransport);
                        _pAccount->KaTransport = NULL;
                        PJ_LOG(2, (THIS_FILE, "Error starting keep-alive timer", Status));
                }
        }
}

/* Keep alive timer callback */
static void KeepAliveTimerCallBack(IN pj_timer_heap_t *_pTimerHeap, IN pj_timer_entry *_pTimerEntry)
{
        PJ_LOG(3, (THIS_FILE, "call KeepAliveTimerCallBack"));
        SipAccount *pAccount;
        pjsip_tpselector TransportSelect;
        pj_time_val Delay;
        char AddrText[PJ_INET6_ADDRSTRLEN];
        pj_status_t Status;

        pj_mutex_lock(SipAppData.pMutex);
        _pTimerEntry->id = PJ_FALSE;
        pAccount = (SipAccount *)_pTimerEntry->user_data;

        if (pAccount->KaTransport == NULL) {
                pj_mutex_unlock(SipAppData.pMutex);
        }

        /* Selet the transport to send the keep alive packet */
        pj_bzero(&TransportSelect, sizeof(TransportSelect));
        TransportSelect.type = PJSIP_TPSELECTOR_TRANSPORT;
        TransportSelect.u.transport = pAccount->KaTransport;

        PJ_LOG(4,(THIS_FILE,
                  "Sending %d bytes keep-alive packet for acc %d",
                  pAccount->Config.KaData.slen, pAccount->nIndex));

        /* Send keepalive raw(\r\n) packet */
        Status = pjsip_tpmgr_send_raw(pjsip_endpt_get_tpmgr(SipAppData.pSipEndPoint),
                                      pAccount->KaTransport->key.type, &TransportSelect,
                                      NULL, pAccount->Config.KaData.ptr,
                                      pAccount->Config.KaData.slen,
                                      &pAccount->KaTarget, pAccount->KaTargetLen,
                                      NULL, NULL);

        if (Status != PJ_SUCCESS && Status != PJ_EPENDING) {
                pj_mutex_unlock(SipAppData.pMutex);
                PrintErrorMsg(Status, "Error on sending keep-alive packet");
                return;
        }

        /* Check just in case keep-alive has been disabled. This shouldn't happen
         * though as when ka_interval is changed this timer should have been
         * cancelled.
         */
        if (pAccount->Config.nKaInterval == 0) {
                pj_mutex_unlock(SipAppData.pMutex);
                return;
        }

        /* Reschedule next timer */
        Delay.sec = pAccount->Config.nKaInterval;
        Delay.msec = 0;
        Status = pjsip_endpt_schedule_timer(SipAppData.pSipEndPoint, _pTimerEntry, &Delay);
        if (Status == PJ_SUCCESS) {
                _pTimerEntry->id = PJ_TRUE;
        } else {
                PJ_LOG(3, (THIS_FILE, "Error sending keep-alive packet, Status = %d", Status));
        }
        pj_mutex_unlock(SipAppData.pMutex);
}

/* On registration transaction callback */
static void onSipRegcTsx(IN struct pjsip_regc_tsx_cb_param *_pCbData)
{
        PJ_LOG(3, (THIS_FILE, "call onSipRegcTsx"));
        // TODO update concat
        if (_pCbData->cbparam.code >= 400 && _pCbData->cbparam.rdata) {
                SipAccount *pAccount = (SipAccount *)_pCbData->cbparam.token;
                if (SipUpdateContactIfNat(pAccount, &_pCbData->cbparam)) {
                        _pCbData->contact_cnt = 1;
                        _pCbData->contact[0] = pAccount->Contact;
                }
        }
}

static pj_bool_t isPrivateIp(const pj_str_t *pAddr)
{
            const pj_str_t PrivateNet[] =
                    {
                            { "10.", 3 },
                            {"100.", 4}, // !!!! I'm not sure about this address
                            { "127.", 4 },
                            { "172.16.", 7 }, { "172.17.", 7 }, { "172.18.", 7 }, { "172.19.", 7 },
                            { "172.20.", 7 }, { "172.21.", 7 }, { "172.22.", 7 }, { "172.23.", 7 },
                            { "172.24.", 7 }, { "172.25.", 7 }, { "172.26.", 7 }, { "172.27.", 7 },
                            { "172.28.", 7 }, { "172.29.", 7 }, { "172.30.", 7 }, { "172.31.", 7 },
                            { "192.168.", 8 }
                    };
            unsigned i;

            for (i=0; i<PJ_ARRAY_SIZE(PrivateNet); ++i) {
                    if (pj_strncmp(pAddr, &PrivateNet[i], PrivateNet[i].slen)==0)
                            return PJ_TRUE;
            }
            return PJ_FALSE;
}

static pj_bool_t SipUpdateContactIfNat(IN SipAccount *_pAccount, IN struct pjsip_regc_cbparam *_pCbData)
{
        const pj_str_t *pViaAddr;
        int nRport;
        pjsip_via_hdr *pVia;

        if (!isPrivateIp(&SipAppData.LocalIp))
                return PJ_FALSE;

        pVia = _pCbData->rdata->msg_info.via;
        if (pVia->rport_param < 1 || (pVia->recvd_param.slen ==0))
                return PJ_FALSE;
        else {
                nRport = pVia->rport_param;
                pViaAddr = &pVia->recvd_param;
        }
        if (pj_strcmp(&_pAccount->ViaAddr.host, pViaAddr) == 0)
                return PJ_FALSE;

        pj_strdup(_pAccount->pPool, &_pAccount->ViaAddr.host, pViaAddr);
        _pAccount->ViaAddr.port = nRport;
        /* Update Via */
        //pjsip_regc_set_via_sent_by(_pAccount->pRegc, &_pAccount->ViaAddr, _pCbData->rdata->tp_info.transport);

        /*Update Contact */
        char tmpContact[64];
        pj_ansi_sprintf(tmpContact, "<sip:%s@%s:%d>", _pAccount->Config.UserName.ptr,  _pAccount->ViaAddr.host.ptr, _pAccount->ViaAddr.port);
        PJ_LOG(4, (THIS_FILE, "Contact change from %s to %s", _pAccount->Contact.ptr, tmpContact));
        pj_strdup2_with_null(_pAccount->pPool, &_pAccount->Contact, tmpContact);
        //        pjsip_regc_update_contact(_pAccount->pRegc, 1, &_pAccount->Contact);
        return PJ_TRUE;
}

static int SipGetFreeCallId(void)
{
        int nCallId;
        for (nCallId = 0; nCallId < SIP_MAX_CALLS; ++nCallId) {
                if (SipAppData.Calls[nCallId].pInviteSession == NULL)
                        return nCallId;
        }
        return -1;
}
int SipMakeNewCall(IN const int _nFromAccountId, IN const char *_pDestUri)
{
        int nCallId;
        SipCall *pCall;
        pjsip_dialog *pDialog;
        pjmedia_sdp_session *pMediaSession;
        pjsip_tx_data *pTransData;
        pj_status_t Status;

        /* Check that account is valid */
        PJ_ASSERT_RETURN(_nFromAccountId >=0 || _nFromAccountId <(int)PJ_ARRAY_SIZE(SipAppData.Accounts),
                         PJ_EINVAL);

        /* Check arguments */
        PJ_ASSERT_RETURN(_pDestUri, PJ_EINVAL);
        PJ_LOG(4, (THIS_FILE, "Making call with acc %d to %s", _nFromAccountId, _pDestUri));

        pj_mutex_lock(SipAppData.pMutex);
        /* Find free call id */
        nCallId = SipGetFreeCallId();
        if (nCallId == -1) {
                PJ_LOG(3, (THIS_FILE, "Too many calls"));
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }
        pCall = &SipAppData.Calls[nCallId];
        pCall->nAccountId = _nFromAccountId;
        pj_str_t Dest = pj_str((char *)_pDestUri);
        /* Create SIP dialog */
        char LocalUri[60];

        pj_ansi_sprintf(LocalUri, "<sip:%s@%s>", SipAppData.Accounts[_nFromAccountId].Config.UserName.ptr, SipAppData.Accounts[_nFromAccountId].Config.SipDomain.ptr);
        pj_str_t Local = pj_str(LocalUri);
        Status = pjsip_dlg_create_uac(pjsip_ua_instance(), &Local, &SipAppData.Accounts[_nFromAccountId].Contact, &Dest,  &Dest, &pDialog);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create uac dialg failed");
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }
        /* TODO get Local SDP */
        CreateTmpSDP(pDialog->pool, pCall, &pMediaSession);
        /* Create invite session */
        Status = pjsip_inv_create_uac(pDialog, pMediaSession, 0, &pCall->pInviteSession);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create uac invite session");
                pjsip_dlg_terminate(pDialog);
                /* TODO destory media resouce */
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }
        pCall->pInviteSession->mod_data[SipMod.id] = pCall;

        pj_gettimeofday(&pCall->StartTime);
        pjsip_auth_clt_set_credentials(&pDialog->auth_sess, SipAppData.Accounts[_nFromAccountId].nCredCnt, SipAppData.Accounts[_nFromAccountId].Cred);
        /* Create initialization Invite request */
        Status = pjsip_inv_invite(pCall->pInviteSession, &pTransData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create invite request failed");
                pjsip_dlg_terminate(pDialog);
                /* TODO destory media resouce */
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }
        Status = pjsip_inv_send_msg(pCall->pInviteSession, pTransData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Send invite request failed");
                pjsip_dlg_terminate(pDialog);
                /* TODO destory media resouce */
                pj_mutex_unlock(SipAppData.pMutex);
                return -1;
        }

        SipAppData.Calls[nCallId].bValid = PJ_TRUE;
        pj_mutex_unlock(SipAppData.pMutex);
        return nCallId;
}

void SipHangUp(IN const int _nCallId)
{
        pjsip_tx_data *pTxData;
        pj_status_t Status;

        if (SipAppData.Calls[_nCallId].pInviteSession == NULL)
                return;

        /* TODO release media resource */
        SipAppData.Calls[_nCallId].bValid = PJ_FALSE;
        Status = pjsip_inv_end_session(SipAppData.Calls[_nCallId].pInviteSession, 603, NULL, &pTxData);
        if (Status == PJ_SUCCESS && pTxData != NULL)
                pjsip_inv_send_msg(SipAppData.Calls[_nCallId].pInviteSession, pTxData);

}

void SipHangUpAll()
{
        int i;
        for (i = 0; i < SIP_MAX_CALLS; ++i) {
                SipHangUp(SipAppData.Calls[i].nIndex);
        }
}

void SipHangUpByAccountId(int _nAccountId)
{
        int i;
        for (i = 0; i < SIP_MAX_CALLS; ++i) {
                if (SipAppData.Calls[i].bValid == PJ_TRUE && SipAppData.Calls[i].nAccountId == _nAccountId)
                        PJ_LOG(3, (THIS_FILE, "Disconnecting call of account #%d, after reregistration attempt failed", _nAccountId));
                        SipHangUp(SipAppData.Calls[i].nIndex);
        }
}
/* Callback to be called to handle incoming requests outside dialogs: */
static pj_bool_t onRxRequest(IN pjsip_rx_data *_pRxData )
{
        pjsip_dialog *pDialog = pjsip_rdata_get_dlg(_pRxData);
        pjsip_msg *pMessage = _pRxData->msg_info.msg;
        pjsip_inv_session *pInviteSession;
        int nToAccountId;
        int nCallId;
        SipCall *pCall;
        pj_status_t Status;
        unsigned nOption;
        pjsip_tx_data *pTxData;
        pjmedia_sdp_session *pSdp;

        /* Only accept INVITE method */
        if (pMessage->line.req.method.id != PJSIP_INVITE_METHOD) {
                return PJ_FALSE;
        }
        /* Don't want accept the call when shutdown is in progress */
        if (SipAppData.bThreadQuit) {
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData,
                                              PJSIP_SC_TEMPORARILY_UNAVAILABLE, NULL, NULL, NULL);
                return PJ_TRUE;
        }
        PJ_LOG(4,(THIS_FILE, "Incoming %s", _pRxData->msg_info.info));

        pj_mutex_lock(SipAppData.pMutex);
        /* Find free call id */
        nCallId = SipGetFreeCallId();
        if (nCallId >= SIP_MAX_CALLS) {
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_BUSY_HERE,
                                              NULL, NULL, NULL);
                pj_mutex_unlock(SipAppData.pMutex);
                return PJ_TRUE;
        }

        pCall = &SipAppData.Calls[nCallId];

        /* Mark call start time */
        pj_gettimeofday(&pCall->StartTime);


        /* Verify that we can handle this request */
        nOption = 0;
        Status = pjsip_inv_verify_request(_pRxData, &nOption, NULL, NULL, SipAppData.pSipEndPoint, &pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Verify request failed");
                pj_str_t Reason = pj_str("Sorry we can't handle this request");
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR,
                                              &Reason, NULL, NULL);
                pj_mutex_unlock(SipAppData.pMutex);
                return PJ_TRUE;
        }

        /* Get account id with associated incoming call */
        nToAccountId = pCall->nAccountId = SipGetAccountIdFromRxData(_pRxData);
        if (nToAccountId == -1) {
                PrintErrorMsg(Status, "Can't find correspond account Id");
                pj_str_t Reason = pj_str("Sorry we can't find right To account Id");
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR,
                                              &Reason, NULL, NULL);
                pj_mutex_unlock(SipAppData.pMutex);
                return PJ_TRUE;
        }

        /* Create UAS dialog */
        Status = pjsip_dlg_create_uas_and_inc_lock(pjsip_ua_instance(), _pRxData,
                                                   &SipAppData.Accounts[nToAccountId].Contact,
                                                   &pDialog);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create Uas dialog failed");
                pj_str_t Reason = pj_str("Sorry we can't create dialog");
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR,
                                              &Reason, NULL, NULL);
                pj_mutex_unlock(SipAppData.pMutex);
                return PJ_TRUE;
        }

        /* Creat Invite Session */
        CreateTmpSDP(pDialog->pool, pCall, &pSdp);
        Status = pjsip_inv_create_uas(pDialog, _pRxData, pSdp, 0, &pCall->pInviteSession);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create UAS invite session failed");
                pj_str_t Reason = pj_str("Sorry we can't create Invite session");
                pjsip_dlg_create_response(pDialog, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR, &Reason, &pTxData);
                pjsip_dlg_send_response(pDialog, pjsip_rdata_get_tsx(_pRxData), pTxData);
                pjsip_dlg_dec_lock(pDialog);
                pj_mutex_unlock(SipAppData.pMutex);
                return PJ_TRUE;
        }
        pCall->pInviteSession->mod_data[SipMod.id] = pCall;
        pj_mutex_unlock(SipAppData.pMutex);
        pjsip_dlg_dec_lock(pDialog);

        pj_gettimeofday(&pCall->StartTime);

        /* TODO put remote SDP to mengke */
        int nContentLen = _pRxData->msg_info.clen->len;
        if (nContentLen != 0) {
                char buf[nContentLen];
                const pjmedia_sdp_session *pRemoteSdp;
                pjmedia_sdp_neg_get_neg_remote(pCall->pInviteSession->neg, &pRemoteSdp);
                pjmedia_sdp_print(pRemoteSdp, buf, nContentLen);
                PJ_LOG(4, (THIS_FILE, "remote SDP\n%s", buf));
        }
        pCall->bValid = PJ_TRUE;

        /* Create a 180 response */
        Status = pjsip_inv_initial_answer(pCall->pInviteSession, _pRxData, PJSIP_SC_RINGING,
                                          NULL, NULL, &pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create 180 response error");
                goto onError;
        }
        /* Send the 180 response. */
        Status = pjsip_inv_send_msg(pCall->pInviteSession, pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Send 180 response error");
                goto onError;
        }

        /* Get from info */
        pjsip_uri *pFromUri;
        pjsip_sip_uri *pFromSipUri;
        pFromUri = _pRxData->msg_info.from->uri;
        pFromSipUri = (pjsip_sip_uri*)pjsip_uri_get_uri(pFromUri);
        char *pFrom = pFromSipUri->user.ptr;
        char From[20];
        int i;
        for (i = 0; i < 20; i++) {
                if (pFrom[i] == '>' || pFrom[i] == ';')
                        break;
                From[i] = pFrom[i];
        }
        From[i] = 0;
        /*TODO Create local SDP */


        /* Now create answer with user response */
        int nAnswerCode = (int)SipAppData.OnIncomingCall(nToAccountId, nCallId,  From);
        Status = pjsip_inv_answer(pCall->pInviteSession,
                                   nAnswerCode, NULL,
                                   NULL,
                                   &pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create user response error");
                goto onError;
        }

        /*  Send the response.*/
        Status = pjsip_inv_send_msg(pCall->pInviteSession, pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Send user response error");
                goto onError;

        }
        return PJ_TRUE;

 onError:
        /* TODO release media resource */
        pjsip_inv_terminate(pCall->pInviteSession, 500, PJ_FALSE);

        /* Release the session */
        pjsip_inv_terminate(pCall->pInviteSession, 500, PJ_FALSE);

        pj_mutex_unlock(SipAppData.pMutex);
        if (pDialog)
                pjsip_dlg_dec_lock(pDialog);
        return PJ_TRUE;
}

static int SipGetAccountIdFromRxData(IN const pjsip_rx_data *_pRxData)
{
        pjsip_uri *pUri;
        pjsip_sip_uri *pSipUri;
        unsigned nAccountId;

        pUri = _pRxData->msg_info.to->uri;
        pj_mutex_lock(SipAppData.pMutex);

        pSipUri = (pjsip_sip_uri*)pjsip_uri_get_uri(pUri);
        int i;
        for (i = 0; i < SipAppData.nAccountCount; ++i) {
                nAccountId = SipAppData.Accounts[i].nIndex;
                SipAccount *pAccount = &SipAppData.Accounts[nAccountId];
                if (pAccount->bValid && pj_stricmp(&pAccount->Config.UserName, &pSipUri->user) == 0
                    && pj_stricmp(&pAccount->Config.SipDomain, &pSipUri->host) == 0 ) {
                        pj_mutex_lock(SipAppData.pMutex);
                        return nAccountId;
                }
        }
        pj_mutex_lock(SipAppData.pMutex);
        return -1;
}

/* Callback to be called when SDP negotiation is done in the call: */
static void onSipCallOnMediaUpdate(IN pjsip_inv_session *_pInviteSession,
                                  pj_status_t _nStatus)
{
        const pjmedia_sdp_session *pRemoteSdp;
        //pjmedia_sdp_neg_get_active_remote(_pInviteSession->neg, &pRemoteSdp);
        /* TODO put remote SDP to mengke */
}

/* Callback to be called when invite session's state has changed: */
static void onSipCallOnStateChanged(IN pjsip_inv_session *_pInviteSession,
                                   IN pjsip_event *_pEvent)
{
        SipCall *pCall = (SipCall *)_pInviteSession->mod_data[SipMod.id];
        PJ_LOG(3, (THIS_FILE, "call state = %d, last answer = %d, callId = %d",
                   _pInviteSession->state, _pInviteSession->cause, pCall->nIndex));
        if (_pInviteSession->state == PJSIP_INV_STATE_EARLY ||
            _pInviteSession->state == PJSIP_INV_STATE_CONNECTING) {
                pj_gettimeofday(&pCall->ResponseTime);
                pj_time_val t = pCall->ResponseTime;
                PJ_TIME_VAL_SUB(t, pCall->StartTime);
                PJ_LOG(4, (THIS_FILE, "Call responsed in %d ms",PJ_TIME_VAL_MSEC(t)));
        } else if (_pInviteSession->state == PJSIP_INV_STATE_CONFIRMED) {
                pj_gettimeofday(&pCall->ConnectTime);
                pj_time_val t = pCall->ConnectTime;
                PJ_TIME_VAL_SUB(t, pCall->StartTime);
                PJ_LOG(4, (THIS_FILE, "Call conFirmed in %d ms",PJ_TIME_VAL_MSEC(t)));
        } else if (_pInviteSession->state == PJSIP_INV_STATE_DISCONNECTED) {
                PJ_LOG(3,(THIS_FILE, "Call #%d disconnected. Reason=%d (%.*s)",
                          pCall->nIndex,
                          _pInviteSession->cause,
                          (int)_pInviteSession->cause_text.slen,
                          _pInviteSession->cause_text.ptr));

                pCall->pInviteSession = NULL;
                _pInviteSession->mod_data[SipMod.id] = NULL;
                /* TODO destory media */
        }
        SipAppData.OnCallStateChange(pCall->nIndex, pCall->nAccountId, (SipInviteState)_pInviteSession->state, (SipAnswerCode)_pInviteSession->cause);
}

/* Callback to be called when dialog has forked: */
static void onSipCallOnForked(pjsip_inv_session *pInviteSession, pjsip_event *pEvent)
{

}

static int CreateTmpSDP(pj_pool_t *_pPool, SipCall *_pCall, pjmedia_sdp_session **_pSdp)
{
        pj_time_val TimeVal;
        pjmedia_sdp_session *pSdp;
        pjmedia_sdp_media *pMedia;
        pjmedia_sdp_attr *pAttr;

        PJ_ASSERT_RETURN(_pPool && _pSdp, PJ_EINVAL);


        /* Create and initialize basic SDP session */
        pSdp = pj_pool_zalloc(_pPool, sizeof(pjmedia_sdp_session));

        pj_gettimeofday(&TimeVal);
        pSdp->origin.user = pj_str("pjsip-siprtp");
        pSdp->origin.version = pSdp->origin.id = TimeVal.sec + 2208988800UL;
        pSdp->origin.net_type = pj_str("IN");
        pSdp->origin.addr_type = pj_str("IP4");
        pSdp->origin.addr = *pj_gethostname();
        pSdp->name = pj_str("pjsip");

        /* Since we only support one media stream at present, put the
         * SDP connection line in the session level.
         */
        pSdp->conn = pj_pool_zalloc (_pPool, sizeof(pjmedia_sdp_conn));
        pSdp->conn->net_type = pj_str("IN");
        pSdp->conn->addr_type = pj_str("IP4");
        pSdp->conn->addr = SipAppData.LocalIp;


        /* SDP time and attributes. */
        pSdp->time.start = pSdp->time.stop = 0;
        pSdp->attr_count = 0;

        /* Create media stream 0: */

        pSdp->media_count = 1;
        pMedia = pj_pool_zalloc (_pPool, sizeof(pjmedia_sdp_media));
        pSdp->media[0] = pMedia;

        /* Standard media info: */
        pMedia->desc.media = pj_str("audio");
        pMedia->desc.port = pj_ntohs(4000);
        pMedia->desc.port_count = 1;
        pMedia->desc.transport = pj_str("RTP/AVP");

        /* Add format and rtpmap for each codec. */
        pMedia->desc.fmt_count = 1;
        pMedia->attr_count = 0;

        {
                pjmedia_sdp_rtpmap rtpmap;
                char ptstr[10];

                sprintf(ptstr, "%d", 0);
                pj_strdup2(_pPool, &pMedia->desc.fmt[0], ptstr);
                rtpmap.pt = pMedia->desc.fmt[0];
                rtpmap.clock_rate = 8000;
                rtpmap.enc_name = pj_str("PCMU");
                rtpmap.param.slen = 0;

                pjmedia_sdp_rtpmap_to_attr(_pPool, &rtpmap, &pAttr);
                pMedia->attr[pMedia->attr_count++] = pAttr;

        }

        /* Add sendrecv attribute. */
        pAttr = pj_pool_zalloc(_pPool, sizeof(pjmedia_sdp_attr));
        pAttr->name = pj_str("sendrecv");
        pMedia->attr[pMedia->attr_count++] = pAttr;

        pMedia->desc.fmt[pMedia->desc.fmt_count++] = pj_str("121");
        /* Add rtpmap. */
        pAttr = pj_pool_zalloc(_pPool, sizeof(pjmedia_sdp_attr));
        pAttr->name = pj_str("rtpmap");
        pAttr->value = pj_str("121 telephone-event/8000");
        pMedia->attr[pMedia->attr_count++] = pAttr;
        /* Add fmtp */
        pAttr = pj_pool_zalloc(_pPool, sizeof(pjmedia_sdp_attr));
        pAttr->name = pj_str("fmtp");
        pAttr->value = pj_str("121 0-15");
        pMedia->attr[pMedia->attr_count++] = pAttr;

        /* Done */
        *_pSdp = pSdp;

        return PJ_SUCCESS;

}
