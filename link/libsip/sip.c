#include "sip.h"
#include "sip_internal.h"

struct SipData SipAppData;

#define THIS_FILE "sip.c"

/* periodically call pjsip_endpt_handle_events to ensure that all events from both
 *transports and timer heap are handled in timely manner
 */
static int PullForEndPointEvent(IN void *_arg);
static void* MQConsumer(void *_arg);
static void ReleaseMsgResource(Message *pMsg);

/* This is a PJSIP module to be registered by application to handle
 * incoming requests outside any dialogs/transactions. The main purpose
 * here is to handle incoming INVITE request message, where we will
 * create a dialog and INVITE session for it.
 */
pjsip_module SipMod = {
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

/* Sip handlers for sdk, indexed by state */
static SipAnswerCode  (*SipHandlers[])(const SipEvent *) =
{
        &OnSipRegAccount,
        &OnSipUnRegAccount,
        &OnSipMakeNewCall,
        &OnSipHangUp,
        &OnSipHangUpByAccountId,
        &OnSipHangUpAll,
        &OnSipAnswerCall,
        &OnSipDestroyInstance,

};

SipAnswerCode SipCreateInstance(IN const SipInstanceConfig *_pConfig)
{
        pj_status_t Status;
        Status = pj_init();
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_PJ_INIT_FAILED);

        Status = pjlib_util_init();
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_PJ_INIT_FAILED);

        /* create pool factory before allow memory */
        pj_caching_pool_init(&SipAppData.Cp, &pj_pool_factory_default_policy, 0);
        SipAppData.pPool = pj_pool_create(&SipAppData.Cp.factory, "SipApp", 1000, 1000, NULL);

        /* Create mutex */
        Status = pj_mutex_create_recursive(SipAppData.pPool, "SipApp", &SipAppData.pMutex);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_PJ_INIT_FAILED);

        /* Init global sip endpoint */
        Status = pjsip_endpt_create(&SipAppData.Cp.factory, pj_gethostname()->ptr,
                                    &SipAppData.pSipEndPoint);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_CREATE_ENDPOINT_FALIED);

        /* start udp socket on sip port */
        /*
        pj_sockaddr Address;
        pjsip_transport *tp;
        pj_sockaddr_init(pj_AF_INET(), &Address, NULL, 0);
        Status = pjsip_udp_transport_start(SipAppData.pSipEndPoint, &Address.ipv4, NULL, 1, &tp);
        SipAppData.LocalPort = tp->local_name.port;
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_START_TP_FAILED);

        PJ_LOG(4,(THIS_FILE, "SIP UDP listening on %.*s:%d",
                  (int)tp->local_name.host.slen, tp->local_name.host.ptr,
                  tp->local_name.port));
        */
        /* start tcp socket on sip port */
        pjsip_tpfactory *TcpFactory;
        pjsip_tcp_transport_cfg TcpConfig;
        pjsip_tcp_transport_cfg_default(&TcpConfig, pj_AF_INET());
        Status = pjsip_tcp_transport_start3(SipAppData.pSipEndPoint, &TcpConfig, &TcpFactory);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_START_TP_FAILED);
        /* Init transaction layer */
        Status = pjsip_tsx_layer_init_module(SipAppData.pSipEndPoint);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_INIT_TRANS_FAILED);

        /* Init UA layer module */
        Status = pjsip_ua_init_module(SipAppData.pSipEndPoint, NULL);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_UA_LAYER_INIT_FAILED);

        /* Add Invite  session module */
        pjsip_inv_callback InviteCallBack;

        /* Init the callback for INVITE session: */
        pj_bzero(&InviteCallBack, sizeof(InviteCallBack));
        InviteCallBack.on_state_changed = &onSipCallOnStateChanged;
        InviteCallBack.on_new_session = &onSipCallOnForked;
        InviteCallBack.on_media_update = &onSipCallOnMediaUpdate;

        /* Initialize invite session module:  */
        Status = pjsip_inv_usage_init(SipAppData.pSipEndPoint, &InviteCallBack);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_INIT_INV_SESS_FALIED);

        /* Initialize 100rel support */
        Status = pjsip_100rel_init_module(SipAppData.pSipEndPoint);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_INIT_100_REL_FALIED);

        /* Initialize session timer support */
        Status = pjsip_timer_init_module(SipAppData.pSipEndPoint);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_INIT_SESS_TIMER_FAILED);

        /*  Register our module to receive incoming requests */
        Status = pjsip_endpt_register_module(SipAppData.pSipEndPoint, &SipMod);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_REG_INCOMING_FAILED);

        /* Register log module */
        Status = pjsip_endpt_register_module(SipAppData.pSipEndPoint, &SipLogger);
        CHECK_RETURN(Status == PJ_SUCCESS, SIP_REG_LOG_FAILED);

        /* Add callback */
        SipAppData.OnRegStateChange = _pConfig->Cb.OnRegStatusChange;
        SipAppData.OnCallStateChange = _pConfig->Cb.OnCallStateChange;
        SipAppData.OnIncomingCall = _pConfig->Cb.OnIncomingCall;
        SipAppData.nMaxAccount = _pConfig->nMaxAccount;
        SipAppData.nMaxCall = _pConfig->nMaxCall;
        SipAppData.nAccountCount = 0;
        SipAppData.nCallCount = 0;
        /* Init Accounts */
        int i;

        SipAppData.Accounts = pj_pool_zalloc(SipAppData.pPool, sizeof(SipAccount) * SipAppData.nMaxAccount);
        for(i = 0; i < SipAppData.nMaxAccount; ++i) {
                SipAppData.Accounts[i].nIndex = i;
                SipAppData.Accounts[i].bValid = PJ_FALSE;
                SipAppData.Accounts[i].nCredCnt = 1;
        }
        /* Init calls */
        SipAppData.Calls = pj_pool_zalloc(SipAppData.pPool, sizeof(SipCall) * SipAppData.nMaxCall);
        for(i = 0; i < SipAppData.nMaxCall; ++i) {
                SipAppData.Calls[i].nIndex = i;
                SipAppData.Calls[i].bValid = PJ_FALSE;
                SipAppData.Calls[i].nAccountId = -1;
        }

        /* Create a thraed to pull pjsip endpoint transport event*/
        pj_thread_create(SipAppData.pPool, "SipWorkThread", &PullForEndPointEvent, NULL, 0, 0, &SipAppData.pSipThread);

        /* Create Message Queue */
        SipAppData.pMq = CreateMessageQueue(MESSAGE_QUEUE_MAX);
        /* Create a thread to consume user function */
        int Ret = pthread_create(&SipAppData.MqThread, NULL, MQConsumer, NULL);
        if (Ret != 0)
                return SIP_CREATE_MQ_THREAD_FAILED;

        return SIP_SUCCESS;
}

SipAnswerCode SipDestroyInstance()
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Type = SIP_DESTROY_INSTANCE;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_DESTROY_INSTANCE;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);

        pthread_join(SipAppData.MqThread, NULL);
        return SIP_SUCCESS;
}

static int PullForEndPointEvent(void *_arg)
{
        while (!SipAppData.bThreadQuit) {
                pj_time_val timeout = {0, 10};
                pjsip_endpt_handle_events(SipAppData.pSipEndPoint, &timeout);
        }
        return 0;
}

static void* MQConsumer(void *_arg)
{
        pj_thread_desc threaddesc;
        pj_thread_t *thread = 0;
        if( !pj_thread_is_registered())
                pj_thread_register(NULL, threaddesc, &thread);
        while (1) {
                Message *pMsg = ReceiveMessageTimeout(SipAppData.pMq, 5000);
                if (!pMsg)
                        continue;
                SipEvent *pEvent = (SipEvent *)pMsg->pMessage;
                SipAnswerCode Ret = SipHandlers[pMsg->nMessageID](pEvent);
                if (Ret != SIP_SUCCESS) {
                        if (pMsg->nMessageID == SIP_REG_ACCOUNT) {
                                int nSdkAccountId = pEvent->Body.Reg.nAccountId;
                                if (SipAppData.OnRegStateChange)
                                        (SipAppData.OnRegStateChange)(nSdkAccountId, Ret, pEvent->Body.Reg.AccConfig.pUserData);
                        }
                        else if (pMsg->nMessageID == SIP_UN_REG_ACCOUNT) {
                                int nSdkAccountId = pEvent->Body.UnReg.nAccountId;
                                int nAccountId = SdkAccToInterAcc(nSdkAccountId);
                                void *pUser = nAccountId == -1? NULL: SipAppData.Accounts[nAccountId].pUserData;
                                if (SipAppData.OnRegStateChange)
                                        (SipAppData.OnRegStateChange)(nSdkAccountId, Ret, pUser);
                        }
                        else if (pMsg->nMessageID == SIP_MAKE_CALL){
                                int nSdkAccountId = pEvent->Body.MakeCall.nAccountId;
                                int nCallId = pEvent->Body.MakeCall.nCallId;
                                int nAccountId = SdkAccToInterAcc(nSdkAccountId);
                                void *pUser = nAccountId == -1? NULL: SipAppData.Accounts[nAccountId].pUserData;
                                if (SipAppData.OnCallStateChange)
                                        (SipAppData.OnCallStateChange)(nCallId, nSdkAccountId, INV_STATE_DISCONNECTED, Ret, pUser, NULL);
                        }
                        else if (pMsg->nMessageID == SIP_ANSWER_CALL) {
                                int nCallId = pEvent->Body.AnswerCall.nCallId;
                                int nSdkAccountId = SipAppData.Calls[nCallId].nAccountId;
                                int nAccountId = SdkAccToInterAcc(nSdkAccountId);
                                void *pUser = nAccountId == -1? NULL: SipAppData.Accounts[nAccountId].pUserData;
                                if (SipAppData.OnCallStateChange)
                                        (SipAppData.OnCallStateChange)(nCallId, nSdkAccountId, INV_STATE_DISCONNECTED, Ret, pUser, NULL);
                        }

                }

                /* release msg after handled */
                /*exit from  thread if Destroy */
                if (pMsg->nMessageID == SIP_DESTROY_INSTANCE) {
                        ReleaseMsgResource(pMsg);
                        break;
                }
                ReleaseMsgResource(pMsg);
        }
        return 0;
}

static void ReleaseMsgResource(Message *pMsg)
{
        SipEvent *pEvent = (SipEvent *)pMsg->pMessage;
        switch (pMsg->nMessageID) {
        case SIP_REG_ACCOUNT:
                {
                        free(pEvent->Body.Reg.AccConfig.pDomain);
                        free(pEvent->Body.Reg.AccConfig.pPassWord);
                        free(pEvent->Body.Reg.AccConfig.pUserName);
                        break;

                }
        case SIP_MAKE_CALL:
                {
                        free(pEvent->Body.MakeCall.pDestUri);
                        pEvent->Body.MakeCall.pDestUri = NULL;
                        break;
                }
        case SIP_ANSWER_CALL:
                {
                        if (pEvent->Body.AnswerCall.Reason)
                                free(pEvent->Body.AnswerCall.Reason);
                        pEvent->Body.AnswerCall.Reason = NULL;
                        break;
                }
        }
        free(pEvent);
        pEvent = NULL;
        free(pMsg);
        pMsg = NULL;
}


SipAnswerCode SipRegAccount(IN const SipAccountConfig *_pConfig, IN const int _nAccountId)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.Reg.nAccountId = _nAccountId;
        pEvent->Body.Reg.AccConfig.nMaxOngoingCall = _pConfig->nMaxOngoingCall;
        pEvent->Body.Reg.AccConfig.pUserData = _pConfig->pUserData;

        pEvent->Body.Reg.AccConfig.pDomain = (char *)malloc(strlen(_pConfig->pDomain) + 1);
        strcpy(pEvent->Body.Reg.AccConfig.pDomain, _pConfig->pDomain);
        pEvent->Body.Reg.AccConfig.pDomain[strlen(_pConfig->pDomain)] = '\0';

        pEvent->Body.Reg.AccConfig.pUserName = (char *)malloc(strlen(_pConfig->pUserName) + 1);
        strcpy(pEvent->Body.Reg.AccConfig.pUserName, _pConfig->pUserName);
        pEvent->Body.Reg.AccConfig.pUserName[strlen(_pConfig->pUserName)] = '\0';

        pEvent->Body.Reg.AccConfig.pPassWord = (char *)malloc(strlen(_pConfig->pPassWord) + 1);
        strcpy(pEvent->Body.Reg.AccConfig.pPassWord, _pConfig->pPassWord);
        pEvent->Body.Reg.AccConfig.pPassWord[strlen(_pConfig->pPassWord)] = '\0';

        pEvent->Type = SIP_REG_ACCOUNT;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }
        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_REG_ACCOUNT;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

SipAnswerCode SipUnRegAccount(IN const int _nAccountId)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.UnReg.nAccountId = _nAccountId;

        pEvent->Type = SIP_UN_REG_ACCOUNT;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }
        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_UN_REG_ACCOUNT;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

SipAnswerCode SipMakeNewCall(IN const int _nFromAccountId, IN const char *_pDestUri, IN const void *_pMedia, IN const int _nCallId)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent) {
                return SIP_MALLOC_FAILED;
        }

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.MakeCall.nAccountId = _nFromAccountId;

        pEvent->Body.MakeCall.pDestUri = (char *)malloc(strlen(_pDestUri) + 1);
        strcpy(pEvent->Body.MakeCall.pDestUri, _pDestUri);
        pEvent->Body.MakeCall.pDestUri[strlen(_pDestUri)] = '\0';

        pEvent->Body.MakeCall.pMedia = (void *)_pMedia;
        pEvent->Body.MakeCall.nCallId = _nCallId;
        pEvent->Type = SIP_MAKE_CALL;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_MAKE_CALL;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

SipAnswerCode SipAnswerCall(IN const int _nCallId, IN const SipAnswerCode _StatusCode, IN const char *_pReason, IN const void *_pMedia)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.AnswerCall.nCallId = _nCallId;
        pEvent->Body.AnswerCall.StatusCode = _StatusCode;
        if (_pReason) {
                pEvent->Body.AnswerCall.Reason = (char *)malloc(strlen(_pReason) + 1);
                strcpy(pEvent->Body.AnswerCall.Reason, _pReason);
                pEvent->Body.AnswerCall.Reason[strlen(_pReason)] = '\0';

        }
        pEvent->Body.AnswerCall.pMedia = (void *)_pMedia;
        pEvent->Type = SIP_ANSWER_CALL;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_ANSWER_CALL;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}
SipAnswerCode SipHangUp(IN const int _nCallId)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.HangUp.nCallId = _nCallId;
        pEvent->Type = SIP_HANGUP;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_HANGUP;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

SipAnswerCode SipHangUpAll()
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Type = SIP_HANGUP_ALL;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_HANGUP_ALL;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

SipAnswerCode SipHangUpByAccountId(int _nAccountId)
{
        SipEvent *pEvent = (SipEvent *)malloc(sizeof(SipEvent));
        if (!pEvent)
                return SIP_MALLOC_FAILED;

        memset(pEvent, 0, sizeof(SipEvent));
        pEvent->Body.HangUpByAcc.nAccount = _nAccountId;
        pEvent->Type = SIP_HANGUP_BY_ACCOUNT;

        Message *pMessage = (Message *)malloc(sizeof(Message));
        if (!pMessage) {
                free(pEvent);
                pEvent = NULL;
                return SIP_MALLOC_FAILED;
        }

        memset(pMessage, 0, sizeof(Message));
        pMessage->nMessageID = SIP_HANGUP_BY_ACCOUNT;
        pMessage->pMessage = (void*)pEvent;
        SendMessage(SipAppData.pMq, pMessage);
        return SIP_SUCCESS;
}

int CreateTmpSDP(OUT void **_pSdp)
{
        pj_pool_t *_pPool = SipAppData.pPool;
        pj_time_val TimeVal;
        pjmedia_sdp_session *pSdp;
        pjmedia_sdp_media *pMedia;
        pjmedia_sdp_attr *pAttr;

        CHECK_RETURN(_pPool && _pSdp, PJ_EINVAL);

        MUTEX_LOCK(SipAppData.pMutex);
        /* Create and initialize basic SDP session */
        pSdp = pj_pool_zalloc(_pPool, sizeof(pjmedia_sdp_session));

        pj_gettimeofday(&TimeVal);
        pSdp->origin.user = pj_str("pjsip-siprtp");
        pSdp->origin.version = pSdp->origin.id = TimeVal.sec + 2208988800UL;
        pSdp->origin.net_type = pj_str("IN");
        pSdp->origin.addr_type = pj_str("IP4");
        pSdp->origin.addr = pj_str("127.0.0.1");
        pSdp->name = pj_str("pjsip");

        /* Since we only support one media stream at present, put the
         * SDP connection line in the session level.
         */
        pSdp->conn = pj_pool_zalloc (_pPool, sizeof(pjmedia_sdp_conn));
        pSdp->conn->net_type = pj_str("IN");
        pSdp->conn->addr_type = pj_str("IP4");
        pSdp->conn->addr = pj_str("172.20.4.69");


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
                pj_strdup2_with_null(_pPool, &pMedia->desc.fmt[0], ptstr);
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
        *_pSdp = (void *)pSdp;
        MUTEX_FREE(SipAppData.pMutex);
        return PJ_SUCCESS;

}

void SipSetLogLevel(IN const int _nLevel)
{
        pj_log_set_level(_nLevel);
}
void PrintSdp(IN const void *_pSdp)
{
        if (_pSdp){
                char buf[500];
                const pjmedia_sdp_session *pRemoteSdp = (pjmedia_sdp_session *)_pSdp;
                pjmedia_sdp_print(pRemoteSdp, buf, 500);
                printf("remote SDP = SDP = %s\n", buf);
        }
}
