#include <unistd.h>
#include <pthread.h>
#include "sip.h"
#include "sip_internal.h"
#define THIS_FILE "sip_internal.c"

extern struct SipData SipAppData;
static void PrintErrorMsg(IN const pj_status_t _Status, IN const char *_pMsg);

/* Init Regc Data */
static pj_status_t SipRegcInit(IN const int _nAccountId);

/* Callback for regc_cb */
static void onSipRegc(IN struct pjsip_regc_cbparam *param);

/* CallBack for regc_tsx_cb */
static void onSipRegcTsx(IN struct pjsip_regc_tsx_cb_param *param);

static void UpdateKeepAlive(INOUT SipAccount *_pAccount, IN const pj_bool_t _Start, IN const struct pjsip_regc_cbparam *pCbData);

static void KeepAliveTimerCallBack(IN pj_timer_heap_t *_pTimerHeap, IN pj_timer_entry *_pTimerEntry);

/* Get local AccountId from incoming message */
static int SipGetAccountIdFromRxData(IN const pjsip_rx_data *_pRxData);

static pj_bool_t SipUpdateContactIfNat(IN SipAccount *_pAccount, IN struct pjsip_regc_cbparam *_pCbData);

/* reregistration if server temporarily falied */
static void SipScheduleReRegistration(SipAccount *_pAccount);
static void SipReRegTimerCallBack(pj_timer_heap_t *_pTimerHeap, pj_timer_entry *_pTimerEntry);


static void register_thread() {
        static pj_thread_desc threaddesc;
        pj_thread_t *thread = 0;
        if( !pj_thread_is_registered())
                pj_thread_register(NULL, threaddesc, &thread);
}
SIP_ERROR_CODE OnSipAddNewAccount(IN const SipAccountConfig *_pConfig, OUT int *_pAccountId)
{
        register_thread();
        /* Input check */
        CHECK_RETURN(_pConfig->pUserName && _pConfig->pPassWord && _pConfig->pDomain, SIP_INVALID_ARG);

        /* Account amount check */
        CHECK_RETURN(SipAppData.nAccountCount < SipAppData.nMaxAccount, SIP_TOO_MANY_ACCOUNT);

        MUTEX_LOCK(SipAppData.pMutex);
        /* Find empty account id. */
        int id;
        for (id=0; id < SipAppData.nMaxAccount; ++id) {
                if (SipAppData.Accounts[id].bValid == PJ_FALSE)
                        break;
        }
        /* Expect to find a slot */
        PJ_ASSERT_ON_FAIL(id < SipAppData.nMaxAccount,
                          {MUTEX_FREE(SipAppData.pMutex); return SIP_TOO_MANY_ACCOUNT;});

        SipAccount *pAccount = &SipAppData.Accounts[id];

        if (pAccount->pPool)
                pj_pool_reset(pAccount->pPool);
        else
                pAccount->pPool = pj_pool_create(&SipAppData.Cp.factory, "SipAcc%p", 512, 256, NULL);

        /* Setup time */
        pAccount->nRegDelayBeforeRefresh = SIP_REG_DELAY_BEFORE_REFRESH;
        pAccount->nRegRetryInterval = SIP_REG_RETRY_INTERNAL;
        pAccount->nRegTimeout = SIP_REG_INTERVAL;
        pAccount->nUnRegTimeout = SIP_UNREG_TIMEOUT;
        pAccount->nKaInterval = SIP_KEEP_ALIVE_INTERVAL;
        pAccount->TimerSetting.sess_expires = SIP_SESSION_EXPIRES;
        pAccount->TimerSetting.min_se = SIP_MIN_SE;

        /* Copy account info */
        char ID[80], Registrar[80];
        sprintf(ID, "sip:%s@%s", _pConfig->pUserName, _pConfig->pDomain);
        sprintf(Registrar, "sip:%s;transport=tcp", _pConfig->pDomain);
        pj_str_t PJID = pj_str(ID);
        pj_str_t PJReg = pj_str(Registrar);
        pj_str_t PJUserName = pj_str((char *)_pConfig->pUserName);
        pj_str_t PJPassword = pj_str((char *)_pConfig->pPassWord);
        pj_str_t PJRealm = pj_str("*");
        pj_str_t PJScheme = pj_str("digest");
        pj_str_t PJDomain = pj_str((char *)_pConfig->pDomain);
        pj_str_t PJKeepAliveData = pj_str("\r\n");

        pAccount->pUserData = _pConfig->pUserData;
        pj_strdup_with_null(pAccount->pPool, &pAccount->RegUri, &PJReg);
        pj_strdup_with_null(pAccount->pPool, &pAccount->KaData, &PJKeepAliveData);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Id, &PJID);
        pj_strdup_with_null(pAccount->pPool, &pAccount->SipDomain, &PJDomain);
        pj_strdup_with_null(pAccount->pPool, &pAccount->UserName, &PJUserName);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].username, &PJUserName);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].data, &PJPassword);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].realm, &PJRealm);
        pj_strdup_with_null(pAccount->pPool, &pAccount->Cred[0].scheme, &PJScheme);
        pAccount->Cred[0].data_type = 0;

        SipAppData.nAccountCount++;
        SipAppData.Accounts[id].bValid = PJ_TRUE;
        *_pAccountId = id;
        SipAppData.Accounts[id].nMaxOngoingCall = _pConfig->nMaxOngoingCall;
        SipAppData.Accounts[id].nOngoingCall = 0;

        MUTEX_FREE(SipAppData.pMutex);

        return SIP_SUCCESS;
}

SIP_ERROR_CODE OnSipDeleteAccount(IN const int _nAccountId)
{
        register_thread();
        CHECK_RETURN(_nAccountId >= 0 && _nAccountId < (int)SipAppData.nMaxAccount,
                     SIP_INVALID_ARG);
        CHECK_RETURN(SipAppData.Accounts[_nAccountId].bValid, SIP_INVALID_ARG);

        SipAccount *pAccount;
        PJ_LOG(4,(THIS_FILE, "Deleting account %d..", _nAccountId));
        /* Hangup all calls */
        SipEvent pSipEvent;
        pSipEvent.Body.HangUpByAcc.nAccount = _nAccountId;
        OnSipHangUpByAccountId(&pSipEvent);

        MUTEX_LOCK(SipAppData.pMutex);
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
        /* Cancel auto-reregistration timer */
        if (pAccount->AutoReReg.ReRegTimer.id) {
                pAccount->AutoReReg.ReRegTimer.id = PJ_FALSE;
                pjsip_endpt_cancel_timer(SipAppData.pSipEndPoint, &pAccount->AutoReReg.ReRegTimer);
        }
        /* Offline account */
        if (pAccount->pRegc) {
                SipEvent pSipEvent;
                pSipEvent.Body.Reg.nAccountId = _nAccountId;
                pSipEvent.Body.Reg.Reg = 0;
                OnSipRegAccount(&pSipEvent);
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
        }

        if (pAccount->pPool) {
                pj_pool_release(pAccount->pPool);
                pAccount->pPool = NULL;
        }
        pAccount->bValid = PJ_FALSE;
        pAccount->Contact.slen = 0;
        SipAppData.nAccountCount--;
        MUTEX_FREE(SipAppData.pMutex);
        return SIP_SUCCESS;
}

SIP_ERROR_CODE OnSipRegAccount(IN const SipEvent *_pEvent)
{
        int _nAccountId = _pEvent->Body.Reg.nAccountId;
        int _bDeReg = _pEvent->Body.Reg.Reg;

        SipAccount *pAccount;
        pj_status_t Status = 0;
        pjsip_tx_data *pTransData = 0;

        PJ_LOG(4,(THIS_FILE, "Acc %d: setting %sregistration..",
                  _nAccountId, (_bDeReg? "" : "un")));
        MUTEX_LOCK(SipAppData.pMutex);

        pAccount = &SipAppData.Accounts[_nAccountId];

        /* cancel re-reg timer if any */
        if (pAccount->AutoReReg.ReRegTimer.id) {
                pAccount->AutoReReg.ReRegTimer.id = PJ_FALSE;
                pjsip_endpt_cancel_timer(SipAppData.pSipEndPoint, &pAccount->AutoReReg.ReRegTimer);
        }
        /* For initial register */
        if (_bDeReg) {
                if (pAccount->pRegc == NULL) {
                        Status = SipRegcInit(_nAccountId);
                        if (Status != PJ_SUCCESS) {
                                PrintErrorMsg(Status, "Unable to create registration, Status");
                                MUTEX_FREE(SipAppData.pMutex);
                                return SIP_CREATE_REG_FAILED;
                        }
                }
                if (!pAccount->pRegc) {
                        MUTEX_FREE(SipAppData.pMutex);
                        return SIP_CREATE_REG_FAILED;
                }

                /* Create register request message */
                Status = pjsip_regc_register(pAccount->pRegc, 1, &pTransData);
        } else {
                if (pAccount->pRegc == NULL) {
                        PJ_LOG(4, (THIS_FILE, "Currently not registered"));
                        MUTEX_FREE(SipAppData.pMutex);
                        return SIP_USR_NOT_REGISTERED;
                }
                Status = pjsip_regc_unregister(pAccount->pRegc, &pTransData);
        }
        if (Status == PJ_SUCCESS) {
                Status = pjsip_regc_send(pAccount->pRegc, pTransData);
        }

        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Unable to create/send register");
                MUTEX_FREE(SipAppData.pMutex);
                return SIP_SEND_REG_FAILED;
        }
        PJ_LOG(4,(THIS_FILE, "Acc %d: %s sent", _nAccountId,
                  (_bDeReg ? "Registration" : "Unregistration")));

        /* Invoke callback function */
        MUTEX_FREE(SipAppData.pMutex);
        return SIP_SUCCESS;
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
                pj_ansi_sprintf(tmpContact, "<sip:%s@%s:%d>", pAccount->UserName.ptr, LocalIp, SipAppData.LocalPort);
        }
        pj_str_t Contact = pj_str(tmpContact);
        pj_strdup(pAccount->pPool, &pAccount->Contact, &Contact);
        Status = pjsip_regc_init(pAccount->pRegc,
                                 &pAccount->RegUri,
                                 &pAccount->Id,
                                 &pAccount->Id,
                                 1,
                                 &pAccount->Contact,
                                 pAccount->nRegTimeout);
        if (Status != PJ_SUCCESS) {
                PJ_LOG(4, (THIS_FILE, "Client registration initialization error", Status));
                pjsip_regc_destroy(pAccount->pRegc);
                pAccount->pRegc = NULL;
                pAccount->Contact.slen = 0;
                return Status;
        }

        pjsip_regc_set_reg_tsx_cb(pAccount->pRegc, onSipRegcTsx);

        /* Set credentials */
        pjsip_regc_set_credentials(pAccount->pRegc, pAccount->nCredCnt, pAccount->Cred);

        pjsip_regc_set_delay_before_refresh(pAccount->pRegc, pAccount->nRegDelayBeforeRefresh);

        return PJ_SUCCESS;
}


/*
 * This callback is called by pjsip_regc when outgoing register
 * request has completed.
 */
static void onSipRegc(IN struct pjsip_regc_cbparam *_pCbData)
{
        SipAccount *pAccount = (SipAccount *)_pCbData->token;
        if (SipAppData.bThreadQuit)
                return;
        MUTEX_LOCK(SipAppData.pMutex);

        /* Print registration Status */
        if (_pCbData->status != PJ_SUCCESS) {
                PJ_LOG(4, (THIS_FILE, "SIP registration error, status = %d", _pCbData->status));
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
                        PJ_LOG(4, (THIS_FILE, "%s: un-registration success", pAccount->Id.ptr));
                } else {
                        PJ_LOG(4, (THIS_FILE, "%s: registration success, status = %d(%.*s)"
                                   "will re-register in %d seconds", pAccount->Id.ptr,
                                   _pCbData->code,
                                   (int)_pCbData->reason.slen, _pCbData->reason.ptr,
                                   _pCbData->expiration));
                        /* Start keep alive timer */
                        UpdateKeepAlive(pAccount, PJ_TRUE, _pCbData);
                }
                pAccount->AutoReReg.Active = PJ_FALSE;
        }

        pAccount->nLastRegCode = _pCbData->code;
        MUTEX_FREE(SipAppData.pMutex);
        if (SipAppData.OnRegStateChange) {
                (SipAppData.OnRegStateChange)(pAccount->nIndex, (SipAnswerCode)pAccount->nLastRegCode, pAccount->pUserData);
        }

        /* hangup call if re-registration attempt failed */
        if (_pCbData->code == PJSIP_SC_REQUEST_TIMEOUT ||
             _pCbData->code == PJSIP_SC_INTERNAL_SERVER_ERROR ||
             _pCbData->code == PJSIP_SC_BAD_GATEWAY ||
             _pCbData->code == PJSIP_SC_SERVICE_UNAVAILABLE ||
             _pCbData->code == PJSIP_SC_SERVER_TIMEOUT ||
             _pCbData->code == PJSIP_SC_TEMPORARILY_UNAVAILABLE ||
            PJSIP_IS_STATUS_IN_CLASS(_pCbData->code, 600)) {
                // SipHangUpByAccountId(pAccount->nIndex); this will trigger by session timer
                MUTEX_LOCK(SipAppData.pMutex);
                SipScheduleReRegistration(pAccount);
                MUTEX_FREE(SipAppData.pMutex);

        }
}
static void SipScheduleReRegistration(SipAccount *_pAccount)
{
        pj_assert(_pAccount);
        if (!_pAccount->bValid || _pAccount->nRegRetryInterval == 0)
                return;

        /* cancel re-reg timer if any */
        if (_pAccount->AutoReReg.ReRegTimer.id) {
                _pAccount->AutoReReg.ReRegTimer.id = PJ_FALSE;
                pjsip_endpt_cancel_timer(SipAppData.pSipEndPoint, &_pAccount->AutoReReg.ReRegTimer);
        }
        _pAccount->AutoReReg.Active = PJ_TRUE;

        _pAccount->AutoReReg.ReRegTimer.cb = &SipReRegTimerCallBack;
        _pAccount->AutoReReg.ReRegTimer.user_data = _pAccount;

        pj_time_val Delay;
        Delay.sec = _pAccount->nRegRetryInterval;
        Delay.msec = 0;
        pj_time_val_normalize(&Delay);
        PJ_LOG(4, (THIS_FILE, "Schedule re-registration retry for acc %d in %u seconds", _pAccount->nIndex, Delay.sec));

        _pAccount->AutoReReg.ReRegTimer.id = PJ_TRUE;
        if (pjsip_endpt_schedule_timer(SipAppData.pSipEndPoint, &_pAccount->AutoReReg.ReRegTimer, &Delay) != PJ_SUCCESS)
                _pAccount->AutoReReg.ReRegTimer.id = PJ_FALSE;
}

static void SipReRegTimerCallBack(pj_timer_heap_t *_pTimerHeap, pj_timer_entry *_pTimerEntry)
{
        SipAccount *pAccount;
        pj_status_t Status;
        pAccount = (SipAccount *)_pTimerEntry->user_data;
        pj_assert(pAccount);
        MUTEX_LOCK(SipAppData.pMutex);

        if (!pAccount->bValid || !pAccount->AutoReReg.Active || pAccount->nRegRetryInterval == 0) {
                MUTEX_FREE(SipAppData.pMutex);
                return;
        }
        if (SipRegAccount(pAccount->nIndex, 1) != SIP_SUCCESS)
                SipScheduleReRegistration(pAccount);
        MUTEX_FREE(SipAppData.pMutex);
}
static void UpdateKeepAlive(INOUT SipAccount *_pAccount, IN const pj_bool_t _Start, IN const struct pjsip_regc_cbparam *_pCbData)
{

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
                if (_pAccount->nKaInterval == 0 || (_pCbData->rdata->tp_info.transport->key.type &
                                                    ~PJSIP_TRANSPORT_IPV6)!= PJSIP_TRANSPORT_UDP)
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

                Delay.sec = _pAccount->nKaInterval;
                Delay.msec = 0;

                Status = pjsip_endpt_schedule_timer(SipAppData.pSipEndPoint, &_pAccount->KaTimer, &Delay);
                if (Status == PJ_SUCCESS) {
                        _pAccount->KaTimer.id = PJ_TRUE;
                        PJ_LOG(4,(THIS_FILE, "Keep-alive timer started for acc %d, "
                                  "destination:%s:%d, interval:%ds",
                                  _pAccount->nIndex,
                                  _pCbData->rdata->pkt_info.src_name,
                                  _pCbData->rdata->pkt_info.src_port,
                                  _pAccount->nKaInterval));
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
        SipAccount *pAccount;
        pjsip_tpselector TransportSelect;
        pj_time_val Delay;
        char AddrText[PJ_INET6_ADDRSTRLEN];
        pj_status_t Status;

        MUTEX_LOCK(SipAppData.pMutex);
        _pTimerEntry->id = PJ_FALSE;
        pAccount = (SipAccount *)_pTimerEntry->user_data;

        if (pAccount->KaTransport == NULL) {
                MUTEX_FREE(SipAppData.pMutex);
        }

        /* Selet the transport to send the keep alive packet */
        pj_bzero(&TransportSelect, sizeof(TransportSelect));
        TransportSelect.type = PJSIP_TPSELECTOR_TRANSPORT;
        TransportSelect.u.transport = pAccount->KaTransport;

        //        PJ_LOG(4,(THIS_FILE,
        //"Sending %d bytes keep-alive packet for acc %d",
                          //        pAccount->KaData.slen, pAccount->nIndex));

        /* Send keepalive raw(\r\n) packet */
        Status = pjsip_tpmgr_send_raw(pjsip_endpt_get_tpmgr(SipAppData.pSipEndPoint),
                                      pAccount->KaTransport->key.type, &TransportSelect,
                                      NULL, pAccount->KaData.ptr,
                                      pAccount->KaData.slen,
                                      &pAccount->KaTarget, pAccount->KaTargetLen,
                                      NULL, NULL);

        if (Status != PJ_SUCCESS && Status != PJ_EPENDING) {
                MUTEX_FREE(SipAppData.pMutex);
                PrintErrorMsg(Status, "Error on sending keep-alive packet");
                return;
        }

        /* Check just in case keep-alive has been disabled. This shouldn't happen
         * though as when ka_interval is changed this timer should have been
         * cancelled.
         */
        if (pAccount->nKaInterval == 0) {
                MUTEX_FREE(SipAppData.pMutex);
                return;
        }

        /* Reschedule next timer */
        Delay.sec = pAccount->nKaInterval;
        Delay.msec = 0;
        Status = pjsip_endpt_schedule_timer(SipAppData.pSipEndPoint, _pTimerEntry, &Delay);
        if (Status == PJ_SUCCESS) {
                _pTimerEntry->id = PJ_TRUE;
        } else {
                PJ_LOG(4, (THIS_FILE, "Error sending keep-alive packet, Status = %d", Status));
        }
        MUTEX_FREE(SipAppData.pMutex);
}

/* On registration transaction callback */
static void onSipRegcTsx(IN struct pjsip_regc_tsx_cb_param *_pCbData)
{
        MUTEX_LOCK(SipAppData.pMutex);
        if (_pCbData->cbparam.code >= 400 && _pCbData->cbparam.rdata) {
                SipAccount *pAccount = (SipAccount *)_pCbData->cbparam.token;
                if (SipUpdateContactIfNat(pAccount, &_pCbData->cbparam)) {
                        _pCbData->contact_cnt = 1;
                        _pCbData->contact[0] = pAccount->Contact;
                }
        }
        MUTEX_FREE(SipAppData.pMutex);
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

        //if (!isPrivateIp(&SipAppData.LocalIp))
        //      return PJ_FALSE;

        pVia = _pCbData->rdata->msg_info.via;
        if (pVia->rport_param < 1 || (pVia->recvd_param.slen ==0))
                return PJ_FALSE;
        else {
                nRport = pVia->rport_param;
                pViaAddr = &pVia->recvd_param;
        }
        if (pj_strcmp(&_pAccount->ViaAddr.host, pViaAddr) == 0)
                return PJ_FALSE;

        pj_strdup_with_null(_pAccount->pPool, &_pAccount->ViaAddr.host, pViaAddr);
        _pAccount->ViaAddr.port = nRport;
        /* Update Via */
        //pjsip_regc_set_via_sent_by(_pAccount->pRegc, &_pAccount->ViaAddr, _pCbData->rdata->tp_info.transport);

        /*Update Contact */
        char tmpContact[64];
        pj_ansi_sprintf(tmpContact, "<sip:%s@%s:%d>", _pAccount->UserName.ptr,  _pAccount->ViaAddr.host.ptr, _pAccount->ViaAddr.port);
        PJ_LOG(4, (THIS_FILE, "Contact change from %s to %s", _pAccount->Contact.ptr, tmpContact));
        pj_strdup2_with_null(_pAccount->pPool, &_pAccount->Contact, tmpContact);
        //        pjsip_regc_update_contact(_pAccount->pRegc, 1, &_pAccount->Contact);
        return PJ_TRUE;
}

int SipGetFreeCallId(void)
{
        int nCallId;
        for (nCallId = 0; nCallId < SipAppData.nMaxCall; ++nCallId) {
                if (SipAppData.Calls[nCallId].bValid == PJ_FALSE)
                        return nCallId;
        }
        return -1;
}
SIP_ERROR_CODE OnSipMakeNewCall(IN const SipEvent *_pEvent)
{
        int _nFromAccountId = _pEvent->Body.MakeCall.nAccountId;
        char *_pDestUri = _pEvent->Body.MakeCall.pDestUri;
        void *_pMedia = _pEvent->Body.MakeCall.pMedia;
        int nCallId = _pEvent->Body.MakeCall.nCallId;

        SipCall *pCall;
        pjsip_dialog *pDialog;
        pjmedia_sdp_session *pMediaSession;
        pjsip_tx_data *pTransData;
        pj_status_t Status;

        PJ_LOG(4, (THIS_FILE, "Making call with acc %d to %s", _nFromAccountId, _pDestUri));

        MUTEX_LOCK(SipAppData.pMutex);

        pCall = &SipAppData.Calls[nCallId];
        pCall->nAccountId = _nFromAccountId;
        pj_str_t Dest = pj_str((char *)_pDestUri);
        /* Create SIP dialog */
        char LocalUri[60];

        pj_ansi_sprintf(LocalUri, "<sip:%s@%s>", SipAppData.Accounts[_nFromAccountId].UserName.ptr, SipAppData.Accounts[_nFromAccountId].SipDomain.ptr);
        pj_str_t Local = pj_str(LocalUri);
        Status = pjsip_dlg_create_uac(pjsip_ua_instance(), &Local, &SipAppData.Accounts[_nFromAccountId].Contact, &Dest,  &Dest, &pDialog);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create uac dialg failed");
                MUTEX_FREE(SipAppData.pMutex);
                pCall->bValid = PJ_FALSE;
                return SIP_CREATE_DLG_FAILED;
        }
        /* TODO get Local SDP */
        //CreateTmpSDP(pDialog->pool, pCall, &pMediaSession);
        pMediaSession = (pjmedia_sdp_session *)_pMedia;
        /* Create invite session */
        unsigned nOptions = 0;
        nOptions |= PJSIP_INV_SUPPORT_100REL;
        nOptions |= PJSIP_INV_SUPPORT_TIMER;
        Status = pjsip_inv_create_uac(pDialog, pMediaSession, nOptions, &pCall->pInviteSession);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create uac invite session");
                pjsip_dlg_terminate(pDialog);
                /* TODO destory media resouce */
                pCall->bValid = PJ_FALSE;
                MUTEX_FREE(SipAppData.pMutex);
                return SIP_CREATE_INV_SESS_FAILED;
        }
        Status = pjsip_timer_init_session(pCall->pInviteSession, &SipAppData.Accounts[_nFromAccountId].TimerSetting);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Session Timer init failed");
                pjsip_dlg_terminate(pDialog);
                pCall->bValid = PJ_FALSE;
                /* TODO destory media resouce */
                MUTEX_FREE(SipAppData.pMutex);
                return SIP_SESS_TIMER_INIT_FALIED;
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
                pCall->bValid = PJ_FALSE;
                MUTEX_FREE(SipAppData.pMutex);
                return SIP_CREATE_INV_REQ_FAILED;
        }
        Status = pjsip_inv_send_msg(pCall->pInviteSession, pTransData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Send invite request failed");
                pjsip_dlg_terminate(pDialog);
                /* TODO destory media resouce */
                pCall->bValid = PJ_FALSE;
                MUTEX_FREE(SipAppData.pMutex);
                return SIP_SNED_INV_REQ_FAILED;
        }

        SipAppData.nCallCount++;
        SipAppData.Accounts[_nFromAccountId].nOngoingCall++;
        MUTEX_FREE(SipAppData.pMutex);
        return SIP_SUCCESS;
}

SIP_ERROR_CODE OnSipAnswerCall(IN const SipEvent *_pEvent)
{
        int _nCallId = _pEvent->Body.AnswerCall.nCallId;
        int _StatusCode = _pEvent->Body.AnswerCall.StatusCode;
        char *_pReason = _pEvent->Body.AnswerCall.Reason;
        void *_pMedia = _pEvent->Body.AnswerCall.pMedia;

        /* Check that callId is valid */
        PJ_LOG(4, (THIS_FILE, "callId = %d answer code = %d, reason = %s \n", _nCallId, _StatusCode, _pReason));
        pj_str_t Reason = pj_str((char *)_pReason);
        pj_str_t *pReason;
        pj_status_t Status;
        pjsip_tx_data *pTxData;
        SipCall *pCall = &SipAppData.Calls[_nCallId];
        pjmedia_sdp_session *pSdp;

        if (!_pReason)
                pReason = NULL;
        else
                pReason = &Reason;

        if (_StatusCode == OK)
                pSdp = (pjmedia_sdp_session*)_pMedia;
        else
                pSdp = NULL;
        MUTEX_LOCK(SipAppData.pMutex);
        Status = pjsip_inv_answer(pCall->pInviteSession,
                                  _StatusCode, pReason,
                                  pSdp,
                                  &pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create user response error");
                Status = SIP_CREATE_RES_FAILED;
                goto onError;
        }

        /*  Send the response.*/
        Status = pjsip_inv_send_msg(pCall->pInviteSession, pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Send user response error");
                Status = SIP_SEND_RES_FAILED;
                goto onError;

        }
        MUTEX_FREE(SipAppData.pMutex);
        return SIP_SUCCESS;

 onError:
        /* Release the session */
        pjsip_inv_terminate(pCall->pInviteSession, 500, PJ_FALSE);
        MUTEX_FREE(SipAppData.pMutex);
        return Status;
}
SIP_ERROR_CODE OnSipHangUp(IN const SipEvent *_pEvent)
{
        pjsip_tx_data *pTxData;
        pj_status_t Status;
        int nCallId = _pEvent->Body.HangUp.nCallId;
        if (SipAppData.Calls[nCallId].pInviteSession == NULL)
                return SIP_SUCCESS;

        /* TODO release media resource */
        MUTEX_LOCK(SipAppData.pMutex);
        SipAppData.Calls[nCallId].bValid = PJ_FALSE;
        Status = pjsip_inv_end_session(SipAppData.Calls[nCallId].pInviteSession, 0, NULL, &pTxData);
        if (Status == PJ_SUCCESS && pTxData != NULL)
                pjsip_inv_send_msg(SipAppData.Calls[nCallId].pInviteSession, pTxData);
        MUTEX_FREE(SipAppData.pMutex);

        return SIP_SUCCESS;
}

SIP_ERROR_CODE OnSipHangUpAll(IN const SipEvent *_pEvent)
{
        int i;
        for (i = 0; i < SipAppData.nMaxCall; ++i) {
                if (SipAppData.Calls[i].bValid == PJ_TRUE && SipAppData.Calls[i].pInviteSession != NULL) {
                        SipEvent Event;
                        Event.Body.HangUp.nCallId = SipAppData.Calls[i].nIndex;
                        OnSipHangUp(&Event);
                }
        }
        return SIP_SUCCESS;
}

SIP_ERROR_CODE OnSipHangUpByAccountId(IN const SipEvent *_pEvent)
{
        int nAccountId = _pEvent->Body.HangUpByAcc.nAccount;
        int i;
        for (i = 0; i < SipAppData.nMaxCall; ++i) {
                if ((SipAppData.Calls[i].bValid == PJ_TRUE) && (SipAppData.Calls[i].nAccountId == nAccountId)) {
                        PJ_LOG(4, (THIS_FILE, "Disconnecting call of account #%d, callId = %d", nAccountId, SipAppData.Calls[i].nIndex));
                        SipEvent Event;
                        Event.Body.HangUp.nCallId = SipAppData.Calls[i].nIndex;
                        OnSipHangUp(&Event);
                }
        }
        return SIP_SUCCESS;
}
/* Callback to be called to handle incoming requests outside dialogs: */
pj_bool_t onRxRequest(IN pjsip_rx_data *_pRxData )
{
        pjsip_dialog *pDialog = pjsip_rdata_get_dlg(_pRxData);
        pjsip_msg *pMessage = _pRxData->msg_info.msg;
        pjsip_inv_session *pInviteSession;
        int nToAccountId;
        int nCallId;
        SipCall *pCall;
        pj_status_t Status;
        pjsip_tx_data *pTxData;

        if (SipAppData.bThreadQuit)
                return PJ_TRUE;
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

        MUTEX_LOCK(SipAppData.pMutex);
        /* Find free call id */
        nCallId = SipGetFreeCallId();
        if (nCallId == -1) {
                PJ_LOG(1, (THIS_FILE, "Too many calls"));
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_BUSY_HERE,
                                              NULL, NULL, NULL);
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }

        pCall = &SipAppData.Calls[nCallId];

        /* Mark call start time */
        pj_gettimeofday(&pCall->StartTime);


        /* Verify that we can handle this request */
        unsigned nOption = 0;
        Status = pjsip_inv_verify_request(_pRxData, &nOption, NULL, NULL, SipAppData.pSipEndPoint, &pTxData);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Verify request failed");
                pj_str_t Reason = pj_str("Sorry we can't handle this request");
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR,
                                              &Reason, NULL, NULL);
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }

        /* Get account id with associated incoming call */
        nToAccountId = pCall->nAccountId = SipGetAccountIdFromRxData(_pRxData);
        if (nToAccountId == -1) {
                PrintErrorMsg(Status, "Can't find correspond account Id");
                pj_str_t Reason = pj_str("Sorry we can't find right account Id");
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR,
                                              &Reason, NULL, NULL);
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }

        if (SipAppData.Accounts[nToAccountId].nOngoingCall >= SipAppData.Accounts[nToAccountId].nMaxOngoingCall) {
                PJ_LOG(1, (THIS_FILE, "Too many call for this account"));
                pjsip_endpt_respond_stateless(SipAppData.pSipEndPoint, _pRxData, PJSIP_SC_BUSY_HERE,
                                              NULL, NULL, NULL);
                MUTEX_FREE(SipAppData.pMutex);
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
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }

        /* Creat Invite Session */
        //CreateTmpSDP(pDialog->pool, pCall, &pSdp);
        unsigned nOptions = 0;
        nOptions |= PJSIP_INV_SUPPORT_100REL;
        nOptions |= PJSIP_INV_SUPPORT_TIMER;
        Status = pjsip_inv_create_uas(pDialog, _pRxData, NULL, nOptions, &pCall->pInviteSession);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Create UAS invite session failed");
                pj_str_t Reason = pj_str("Sorry we can't create Invite session");
                pjsip_dlg_create_response(pDialog, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR, &Reason, &pTxData);
                pjsip_dlg_send_response(pDialog, pjsip_rdata_get_tsx(_pRxData), pTxData);
                pjsip_dlg_dec_lock(pDialog);
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }

        Status = pjsip_timer_init_session(pCall->pInviteSession, &SipAppData.Accounts[nToAccountId].TimerSetting);
        if (Status != PJ_SUCCESS) {
                PrintErrorMsg(Status, "Session Timer init failed");
                pj_str_t Reason = pj_str("Sorry Init Session Timer Falied");
                pjsip_dlg_create_response(pDialog, _pRxData, PJSIP_SC_INTERNAL_SERVER_ERROR, &Reason, &pTxData);
                pjsip_dlg_send_response(pDialog, pjsip_rdata_get_tsx(_pRxData), pTxData);
                pjsip_dlg_dec_lock(pDialog);
                MUTEX_FREE(SipAppData.pMutex);
                return PJ_TRUE;
        }
        pCall->pInviteSession->mod_data[SipMod.id] = pCall;
        pjsip_dlg_dec_lock(pDialog);

        pj_gettimeofday(&pCall->StartTime);
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

        /* TODO put remote SDP to mengke */
        const pjmedia_sdp_session *pRemoteSdp = NULL;
        pjsip_rdata_sdp_info *pSdpInfo = NULL;
        pSdpInfo = pjsip_rdata_get_sdp_info(_pRxData);
        if (pSdpInfo)
                pRemoteSdp = pSdpInfo->sdp;
        int Code =  SipAppData.OnIncomingCall(nToAccountId, nCallId,  From, SipAppData.Accounts[pCall->nAccountId].pUserData, (void *)pRemoteSdp);
        /* Faild in callback function */
        if (Code != OK) {
                pj_str_t Reason = pj_str("Faild in callback function");
                Status = pjsip_inv_answer(pCall->pInviteSession,
                                          Code, &Reason,
                                          NULL,
                                          &pTxData);
                if (Status != PJ_SUCCESS) {
                        PrintErrorMsg(Status, "Create user response error");
                        goto onError;

                }

                Status = pjsip_inv_send_msg(pCall->pInviteSession, pTxData);
                if (Status != PJ_SUCCESS) {
                        PrintErrorMsg(Status, "Send user response error");
                        goto onError;
                }
        }

        SipAppData.nCallCount++;
        SipAppData.Accounts[nToAccountId].nOngoingCall++;
        MUTEX_FREE(SipAppData.pMutex);
        return PJ_TRUE;

 onError:
        /* Release the session */
        pjsip_inv_terminate(pCall->pInviteSession, 500, PJ_FALSE);

        MUTEX_FREE(SipAppData.pMutex);
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
        pSipUri = (pjsip_sip_uri*)pjsip_uri_get_uri(pUri);
        int i;
        for (i = 0; i < SipAppData.nAccountCount; ++i) {
                nAccountId = SipAppData.Accounts[i].nIndex;
                SipAccount *pAccount = &SipAppData.Accounts[nAccountId];
                if (pAccount->bValid && pj_stricmp(&pAccount->UserName, &pSipUri->user) == 0
                    && pj_stricmp(&pAccount->SipDomain, &pSipUri->host) == 0 ) {
                        return nAccountId;
                }
        }
        return -1;
}

/* Callback to be called when SDP negotiation is done in the call: */
void onSipCallOnMediaUpdate(IN pjsip_inv_session *_pInviteSession,
                                  pj_status_t _nStatus)
{
        const pjmedia_sdp_session *pRemoteSdp;
        //pjmedia_sdp_neg_get_active_remote(_pInviteSession->neg, &pRemoteSdp);
        /* TODO put remote SDP to mengke */
}

/* Callback to be called when invite session's state has changed: */
void onSipCallOnStateChanged(IN pjsip_inv_session *_pInviteSession,
                                   IN pjsip_event *_pEvent)
{
        if (SipAppData.bThreadQuit)
                return;

        SipCall *pCall = (SipCall *)_pInviteSession->mod_data[SipMod.id];
        const pjmedia_sdp_session *pRemoteSdp = NULL;
        pjsip_rx_data *RxData = NULL;
        pjsip_rdata_sdp_info *pSdpInfo = NULL;
        PJ_LOG(4, (THIS_FILE, "call state = %d, last answer = %d, callId = %d, eventId = %d",
                   _pInviteSession->state, _pInviteSession->cause, pCall->nIndex, _pEvent->type));

        MUTEX_LOCK(SipAppData.pMutex);
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
                PJ_LOG(4,(THIS_FILE, "Call #%d disconnected. Reason=%d (%.*s)",
                          pCall->nIndex,
                          _pInviteSession->cause,
                          (int)_pInviteSession->cause_text.slen,
                          _pInviteSession->cause_text.ptr));
                pCall->pInviteSession = NULL;
                pCall->bValid = PJ_FALSE;
                _pInviteSession->mod_data[SipMod.id] = NULL;
                SipAppData.nCallCount--;
                SipAppData.Accounts[pCall->nAccountId].nOngoingCall--;
        }
        if (_pInviteSession->state == PJSIP_INV_STATE_CONNECTING && _pInviteSession->role == PJSIP_ROLE_UAC) {
                RxData = _pEvent->body.rx_msg.rdata;
                pSdpInfo = pjsip_rdata_get_sdp_info(RxData);
                if (pSdpInfo)
                        pRemoteSdp = pSdpInfo->sdp;
        }
        MUTEX_FREE(SipAppData.pMutex);
        SipAppData.OnCallStateChange(pCall->nIndex, pCall->nAccountId, (SipInviteState)_pInviteSession->state, (SipAnswerCode)_pInviteSession->cause, SipAppData.Accounts[pCall->nAccountId].pUserData, (void*)pRemoteSdp);
}

/* Callback to be called when dialog has forked: */
void onSipCallOnForked(pjsip_inv_session *pInviteSession, pjsip_event *pEvent)
{

}

static void PrintErrorMsg(IN const pj_status_t _Status, IN const char *_pMsg)
{
        char errmsg[PJ_ERR_MSG_SIZE];
        pj_strerror(_Status, errmsg, sizeof(errmsg));
        PJ_LOG(1, (THIS_FILE, "%s: %s [Status=%d]", _pMsg, errmsg, _Status));
}
