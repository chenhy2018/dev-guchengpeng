#include "PeerConnection.h"
#define THIS_FILE "PeerConnection.c"

static int createMediaSdpMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport *_pTransport,
                               IN pj_pool_t *_pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media **_pSdp);
static int createSdpMline(IN OUT PeerConnection * _pPeerConnection, pj_pool_t *_pPool, pjmedia_sdp_session *pSdp);
static int negotiationSettingAfterSuccess(IN PeerConnection * _pPeerConnection);
int setLocalDescription(IN OUT PeerConnection * _pPeerConnection, IN void * _pSdp);

enum { RTCP_INTERVAL = 5000, RTCP_RAND = 2000 };


#define  LIBRTP_REGISTER_THREAD() {\
        pj_thread_desc pc_desc; \
        if(!pj_thread_is_registered()){ \
                pj_thread_t *pThread; \
                pj_thread_register("test", pc_desc, &pThread); \
        } \
} \

static void print_sdp(pjmedia_sdp_session * _pSdp, const char * _pLogPrefix)
{
        if (5 <= pj_log_get_level()) {
                char sdpStr[2048];
                memset(sdpStr, 0, 2048);
                pjmedia_sdp_print(_pSdp, sdpStr, sizeof(sdpStr));
                if (_pLogPrefix != NULL) {
                        MY_PJ_LOG(5, "%s:%s", _pLogPrefix, sdpStr);
                } else {
                        MY_PJ_LOG(5, "%s", sdpStr);
                }
        }
}

static inline int GetTransportIndex(IN PeerConnection * _pPeerConnection, IN TransportIce *pTransportIce)
{
        for (int i = 0; i < sizeof(_pPeerConnection->transportIce) / sizeof(TransportIce); i++) {
                if (&_pPeerConnection->transportIce[i] == pTransportIce) {
                        return i;
                }
        }

        return -1;
}

static void internal_write_sdp(pjmedia_sdp_session * pSdp, char * pFname)
{
        FILE * f = fopen(pFname, "wb");
        assert(f != NULL);
        
        char sdpStr[2048];
        memset(sdpStr, 0, 2048);
        int nLen = pjmedia_sdp_print(pSdp, sdpStr, sizeof(sdpStr));
        
        int nWlen = fwrite(sdpStr, 1, nLen, f);
        assert(nWlen == nLen);
        
        fclose(f);
}

static int addCandidate(TransportIce *_pTransportIce, pjmedia_sdp_session **_pSdp)
{
        pj_assert(_pTransportIce->iceState == ICE_STATE_GATHERING_OK);
        PeerConnection * pPeerConnection = (PeerConnection *)_pTransportIce->pPeerConnection;
        //int nIdx = GetTransportIndex(pPeerConnection, _pTransportIce);
        //MediaStreamTrack *pMediaTrack = &pPeerConnection->mediaStream.streamTracks[nIdx];

        pj_status_t status;
        pjmedia_sdp_session *pSdp = NULL;
        if (pPeerConnection->role == ICE_ROLE_OFFERER) {
                pSdp = pPeerConnection->pOfferSdp;
                status = createSdpMline(pPeerConnection, pPeerConnection->pSdpPool, pSdp);
                STATUS_CHECK(createSdpMline, status);

                int nMaxTracks = sizeof(pPeerConnection->nAvIndex) / sizeof(int);
                for ( int i = 0; i < nMaxTracks; i++) {
                        if (pPeerConnection->nAvIndex[i] != -1) {
                                status = pjmedia_transport_media_create(pPeerConnection->transportIce[i].pTransport,
                                                                        pPeerConnection->pSdpPool, 0, NULL, 0);
                                STATUS_CHECK(pjmedia_transport_media_create, status);

                                status = pjmedia_transport_encode_sdp(pPeerConnection->transportIce[i].pTransport,
                                                                      pPeerConnection->pSdpPool, pSdp, NULL, i);
                                STATUS_CHECK(pjmedia_transport_encode_sdp, status);
                        }
                }
                setLocalDescription(pPeerConnection, pSdp);
                internal_write_sdp(pSdp, "offer.sdp");
                print_sdp(pSdp, "offer sdp add candidate");
        } else {
                pSdp = pPeerConnection->pAnswerSdp;
                status = createSdpMline(pPeerConnection, pPeerConnection->pSdpPool, pSdp);
                STATUS_CHECK(createSdpMline, status);

                int nMaxTracks = sizeof(pPeerConnection->nAvIndex) / sizeof(int);
                for ( int i = 0; i < nMaxTracks; i++) {
                        if (pPeerConnection->nAvIndex[i] != -1) {
                                status = pjmedia_transport_media_create(pPeerConnection->transportIce[i].pTransport,
                                                                        pPeerConnection->pSdpPool, 0, pPeerConnection->pOfferSdp, 0);
                                STATUS_CHECK(pjmedia_transport_media_create, status);

                                status = pjmedia_transport_encode_sdp(pPeerConnection->transportIce[i].pTransport,
                                                                      pPeerConnection->pSdpPool, pSdp,
                                                                      pPeerConnection->pOfferSdp, i);
                                STATUS_CHECK(pjmedia_transport_encode_sdp, status);
                        }
                }
                setLocalDescription(pPeerConnection, pSdp);
                internal_write_sdp(pSdp, "answer.sdp");
                print_sdp(pSdp, "answer sdp add candidate");
        }

        *_pSdp = pSdp;
        return PJ_SUCCESS;
}

static void doUserCallback(PeerConnection * _pPeerConnection, IceState _state, void *_pData)
{
        pj_mutex_lock(_pPeerConnection->pMutex);
        _pPeerConnection->iceNegInfo.state = _state;
        _pPeerConnection->iceNegInfo.pData = _pData;
        if (_pPeerConnection->userIceConfig.userCallback) {
                pj_mutex_unlock(_pPeerConnection->pMutex);
                _pPeerConnection->userIceConfig.userCallback(_pPeerConnection->userIceConfig.pCbUserData,
                                                            CALLBACK_ICE, &_pPeerConnection->iceNegInfo);
        } else {
                pj_mutex_unlock(_pPeerConnection->pMutex);
        }
        return;
}

static void onIceComplete2(pjmedia_transport *pTransport, pj_ice_strans_op op,
                           pj_status_t status, void *pUserData) {
        TransportIce *pTransportIce = (TransportIce *)pUserData;
        PeerConnection * pPeerConnection = (PeerConnection *)pTransportIce->pPeerConnection;

        if (pPeerConnection->nIsFailCallbackDone) {
                MY_PJ_LOG(1, "ice already fail and callback to user:state:%d status:%d",op, status);
                return;
        }
        if(status != PJ_SUCCESS){
                MY_PJ_LOG(1, "onIceComplete2 state:%d status:%d",op, status);
                if (!pPeerConnection->nIsFailCallbackDone) {
                        pPeerConnection->nIsFailCallbackDone = 1;
                        pTransportIce->iceState = ICE_STATE_FAIL;
                        doUserCallback(pPeerConnection, ICE_STATE_FAIL, NULL);
                }
                return;
        }


        //pTransportIce->iceState =  op;
        switch (op) {
                        /** Initialization (candidate gathering) */
                case PJ_ICE_STRANS_OP_INIT:
                        pTransportIce->iceState = ICE_STATE_GATHERING_OK;
                        pj_mutex_lock(pPeerConnection->pMutex);
                        pPeerConnection->nGatherCandidateSuccessCount++;
                        MY_PJ_LOG(3, "--->gathering candidates finish. total:%d count:%d pPeerConnection %p", pPeerConnection->mediaStream.nCount,
                                  pPeerConnection->nGatherCandidateSuccessCount, pPeerConnection);
                        pjmedia_sdp_session *pSdp = NULL;
                        if (pPeerConnection->nGatherCandidateSuccessCount == pPeerConnection->mediaStream.nCount) {
                                status = addCandidate(pTransportIce, &pSdp);
                        } else {
                                pj_mutex_unlock(pPeerConnection->pMutex);
                                return;
                        }
                        pj_mutex_unlock(pPeerConnection->pMutex);
                        if (status != PJ_SUCCESS) {
                                MY_PJ_LOG(1, "--->gathering candidates finish. but addCandidate fail:%d", status);
                                doUserCallback(pPeerConnection, ICE_STATE_GATHERING_FAIL, NULL);
                                return;
                        }
                        MY_PJ_LOG(3, "--->gathering candidates finish. addCandidate ok");
                        doUserCallback(pPeerConnection, ICE_STATE_GATHERING_OK, pSdp);

                        break;
                        
                        /** Negotiation */
                case PJ_ICE_STRANS_OP_NEGOTIATION:
                        pTransportIce->iceState = ICE_STATE_NEGOTIATION_OK;
                        pPeerConnection->nNegSuccess++;
                        MY_PJ_LOG(3, "--->negotiation finish: total:%d count:%d", pPeerConnection->mediaStream.nCount,
                                  pPeerConnection->nNegSuccess);
                        
                        if (pPeerConnection->nNegSuccess == pPeerConnection->mediaStream.nCount) {
                                status = negotiationSettingAfterSuccess(pPeerConnection);
                        } else {
                                return;
                        }
                        if (status != PJ_SUCCESS) {
                                MY_PJ_LOG(1, "--->negotiation finish, but fail. status:%d", status);
                                doUserCallback(pPeerConnection, ICE_STATE_NEGOTIATION_FAIL, NULL);
                                return;
                        }
                        MY_PJ_LOG(3, "--->negotiation ok");

                        doUserCallback(pPeerConnection, ICE_STATE_NEGOTIATION_OK, NULL);
                        break;
                        
                        /** This operation is used to report failure in keep-alive operation.
                         *  Currently it is only used to report TURN Refresh failure.  */
                case PJ_ICE_STRANS_OP_KEEP_ALIVE:
                        MY_PJ_LOG(3, "--->PJ_ICE_STRANS_OP_KEEP_ALIVE");
                        break;
                        
                        /** IP address change notification from STUN keep-alive operation.  */
                case PJ_ICE_STRANS_OP_ADDR_CHANGE:
                        MY_PJ_LOG(3, "--->PJ_ICE_STRANS_OP_ADDR_CHANGE");
                        break;
        }
}

static int iceWorkerThread(void * _pArg)
{
        fprintf(stderr, "iceWorkerThread start");
        TransportIce * pTransportIce = (TransportIce *)_pArg;
        pj_ice_strans_cfg * pIceCfg = &pTransportIce->iceConfig;
        PeerConnection * pPeerConnection = (PeerConnection*)pTransportIce->pPeerConnection;
        
        while (!pPeerConnection->bQuit) {
                enum { MAX_NET_EVENTS = 1 };
                pj_time_val maxTimeout = {0, 0};
                pj_time_val timeout = { 0, 0};
                unsigned count = 0, nNetEventCount = 0;
                int c;
                
                maxTimeout.msec = 500;
                
                /* Poll the timer to run it and also to retrieve the earliest entry. */
                timeout.sec = timeout.msec = 0;
                c = pj_timer_heap_poll( pIceCfg->stun_cfg.timer_heap, &timeout );
                if (c > 0)
                        count += c;
                
                /* timer_heap_poll should never ever returns negative value, or otherwise
                 * ioqueue_poll() will block forever!
                 */
                pj_assert(timeout.sec >= 0 && timeout.msec >= 0);
                if (timeout.msec >= 1000) timeout.msec = 999;
                
                /* compare the value with the timeout to wait from timer, and use the
                 * minimum value.
                 */
                if (PJ_TIME_VAL_GT(timeout, maxTimeout))
                        timeout = maxTimeout;
                
                /* Poll ioqueue.
                 * Repeat polling the ioqueue while we have immediate events, because
                 * timer heap may process more than one events, so if we only process
                 * one network events at a time (such as when IOCP backend is used),
                 * the ioqueue may have trouble keeping up with the request rate.
                 *
                 * For example, for each send() request, one network event will be
                 *   reported by ioqueue for the send() completion. If we don't poll
                 *   the ioqueue often enough, the send() completion will not be
                 *   reported in timely manner.
                 */
                do {
                        c = pj_ioqueue_poll( pIceCfg->stun_cfg.ioqueue, &timeout);
                        if (c < 0) {
                                pj_status_t err = pj_get_netos_error();
                                pj_thread_sleep(PJ_TIME_VAL_MSEC(timeout));
                                return err;
                        } else if (c == 0) {
                                break;
                        } else {
                                nNetEventCount += c;
                                timeout.sec = timeout.msec = 0;
                        }
                } while (c > 0 && nNetEventCount < MAX_NET_EVENTS);
                
                count += nNetEventCount;
        }
        
        return 0;
}

static pj_str_t parseIpAndPort(char *_pHosStr, pj_uint16_t *_pPort)
{
        pj_str_t ret;

        char *pPosition;
        pj_str_t host = pj_str(_pHosStr);
        if ((pPosition = pj_strchr(&host, ':')) != NULL) {
                ret.ptr = _pHosStr;
                ret.slen = (pPosition - _pHosStr);
                *_pPort = (pj_uint16_t)atoi(pPosition + 1);
        } else {
                ret = pj_str(_pHosStr);
                *_pPort = PJ_STUN_PORT;
        }
        return ret;
}

static int peerConnectInitIceConfig(IN OUT PeerConnection * _pPeerConnectoin)
{
        for (int i = 0; i < sizeof(_pPeerConnectoin->transportIce) / sizeof(TransportIce); i++) {
                pj_ice_strans_cfg * pIceCfg = &_pPeerConnectoin->transportIce[i].iceConfig;
                pj_ice_strans_cfg_default(pIceCfg);
                
                //stun turn deprecated
                pj_bzero(&pIceCfg->stun, sizeof(pIceCfg->stun));
                pj_bzero(&pIceCfg->turn, sizeof(pIceCfg->turn));
                
                pIceCfg->af = pj_AF_INET();
                
                IceConfig *pUserIceConfig = &_pPeerConnectoin->userIceConfig;
                if (pUserIceConfig->bRegular) {
                        pIceCfg->opt.aggressive = PJ_FALSE;
                } else {
                        pIceCfg->opt.aggressive = PJ_TRUE;
                }
                
                
                if (pUserIceConfig->nForceStun || pUserIceConfig->stunHost[0] != '\0') {
                        pIceCfg->stun_tp_cnt = 1;
                        pj_ice_strans_stun_cfg_default(&pIceCfg->stun_tp[0]);
                        
                        pIceCfg->stun_tp[0].max_host_cands = pUserIceConfig->nMaxHosts;
                        char *pServer = NULL;
                        if (pUserIceConfig->stunHost[0] != '\0') {
                                pServer = pUserIceConfig->stunHost;
                        } else if (pUserIceConfig->turnHost[0] != '\0') {
                                pServer = pUserIceConfig->turnHost;
                        }
                        if (pServer != NULL) {
                                pj_uint16_t nPort;
                                pIceCfg->stun_tp[0].server = parseIpAndPort(pServer, &nPort);
                                pIceCfg->stun_tp[0].port = nPort;
                        }
                        if (pUserIceConfig->nKeepAlive > 0) {
                                pIceCfg->stun_tp[0].cfg.ka_interval = pUserIceConfig->nKeepAlive;
                        }
                }
                
                
                if (pUserIceConfig->turnHost[0] != '\0') {
                        pIceCfg->turn_tp_cnt = 1;
                        pj_ice_strans_turn_cfg_default(&pIceCfg->turn_tp[0]);

                        if (pUserIceConfig->turnHost[0] != '\0') {
                                pj_uint16_t nPort;
                                pIceCfg->turn_tp[0].server = parseIpAndPort(pUserIceConfig->turnHost, &nPort);
                                pIceCfg->turn_tp[0].port = nPort;
                        }

                        if (pUserIceConfig->turnUsername[0] != '\0') {
                                pIceCfg->turn_tp[0].auth_cred.type = PJ_STUN_AUTH_CRED_STATIC;
                                pIceCfg->turn_tp[0].auth_cred.data.static_cred.username = pj_str(pUserIceConfig->turnUsername);
                                pIceCfg->turn_tp[0].auth_cred.data.static_cred.data_type = PJ_STUN_PASSWD_PLAIN;
                                pIceCfg->turn_tp[0].auth_cred.data.static_cred.data = pj_str(pUserIceConfig->turnPassword);
                        }
                        
                        if (pUserIceConfig->bTurnTcp) {
                                pIceCfg->turn_tp[0].conn_type = PJ_TURN_TP_TCP;
                        } else {
                                pIceCfg->turn_tp[0].conn_type = PJ_TURN_TP_UDP;
                        }
                }
                
                pj_status_t status;
                if (pUserIceConfig->nameserver[0] != '\0') {
                        pj_str_t nameserver = pj_str(pUserIceConfig->nameserver);
                        status = pj_dns_resolver_create(_pPeerConnectoin->pPoolFactory, "resolver", 0,
                                                        pIceCfg->stun_cfg.timer_heap,
                                                        pIceCfg->stun_cfg.ioqueue,
                                                        &pIceCfg->resolver);
                        STATUS_CHECK(pj_dns_resolver_create, status);
                        status = pj_dns_resolver_set_ns(pIceCfg->resolver, 1, &nameserver, NULL);
                        STATUS_CHECK(pj_dns_resolver_set_ns, status);
                }
        }
        return PJ_SUCCESS;
}

static pj_status_t initTransportIce(IN PeerConnection * _pPeerConnectoin, OUT TransportIce * _pTransportIce)
{
        pj_assert(_pPeerConnectoin->pMediaEndpt);
        pj_status_t status;
        
        pj_pool_t * pIoQueuePool = pj_pool_create(_pPeerConnectoin->pPoolFactory, NULL, 512, 512, NULL);
        ASSERT_RETURN_CHECK(pIoQueuePool, pj_pool_create);
        _pTransportIce->pIoqueuePool = pIoQueuePool;
        pj_ioqueue_t* pIoQueue;
        status = pj_ioqueue_create(pIoQueuePool, 16, &pIoQueue);
        STATUS_CHECK(pj_ioqueue_create, status);
        _pTransportIce->pIoQueue = pIoQueue;
        
        pj_pool_t * pTimerPool = pj_pool_create(_pPeerConnectoin->pPoolFactory, NULL, 512, 512, NULL);
        ASSERT_RETURN_CHECK(pTimerPool, pj_pool_create);
        _pTransportIce->pTimerHeapPool = pTimerPool;
        pj_timer_heap_t *pTimerHeap = NULL;
        status = pj_timer_heap_create(pTimerPool, 10, &pTimerHeap);
        STATUS_CHECK(pj_timer_heap_create, status);
        _pTransportIce->pTimerHeap = pTimerHeap;
        
        
        pj_stun_config_init(&_pTransportIce->iceConfig.stun_cfg, _pPeerConnectoin->pPoolFactory, 0,
                            pIoQueue, pTimerHeap);
        
        _pTransportIce->pPeerConnection = _pPeerConnectoin;
        
        pj_thread_t * pThread;
        pj_pool_t * pThreadPool = pj_pool_create(_pPeerConnectoin->pPoolFactory, NULL, 512, 512, NULL);
        ASSERT_RETURN_CHECK(pThreadPool, pj_pool_create);
        _pTransportIce->pThreadPool = pThreadPool;
        status = pj_thread_create(pThreadPool, "iceWorkerThread", &iceWorkerThread, _pTransportIce, 0, 0, &pThread);
        STATUS_CHECK(pj_thread_create, status);
        _pTransportIce->pPollThread = pThread;
        
        pjmedia_ice_cb cb;
        cb.on_ice_complete = NULL;
        cb.on_ice_complete2 = onIceComplete2;
        
        pjmedia_transport *transport = NULL;
        status = pjmedia_ice_create3(_pPeerConnectoin->pMediaEndpt, NULL, _pPeerConnectoin->userIceConfig.nComponents,
                                     &_pTransportIce->iceConfig, &cb, 0, _pPeerConnectoin, &transport);
        STATUS_CHECK(pjmedia_ice_create3, status);
        pjmedia_ice_add_ice_cb(transport, &cb, _pTransportIce);
        _pTransportIce->pTransport = transport;
        
        return PJ_SUCCESS;
}

static void transportIceDestroy(IN OUT TransportIce * _pTransportIce)
{
        if (_pTransportIce->pPollThread) {
                pj_thread_join(_pTransportIce->pPollThread);
        }

        if (_pTransportIce->pTransport) {
                pjmedia_transport_media_stop(_pTransportIce->pTransport);
                _pTransportIce->pTransport = NULL;
        }
        
        if (_pTransportIce->pPollThread) {
                pj_thread_destroy(_pTransportIce->pPollThread);
                _pTransportIce->pPollThread = NULL;
        }
        
        if (_pTransportIce->pIoQueue) {
                pj_ioqueue_destroy(_pTransportIce->pIoQueue);
                _pTransportIce->pIoQueue = NULL;
        }
        
        if (_pTransportIce->pTimerHeap ) {
                pj_timer_heap_destroy(_pTransportIce->pTimerHeap);
                _pTransportIce->pTimerHeap = NULL;
        }
        
        if (_pTransportIce->pIoqueuePool) {
                pj_pool_release(_pTransportIce->pIoqueuePool);
                _pTransportIce->pIoqueuePool = NULL;
        }
        
        if (_pTransportIce->pThreadPool) {
                pj_pool_release(_pTransportIce->pThreadPool);
                _pTransportIce->pThreadPool = NULL;
        }
        
        if (_pTransportIce->pTimerHeapPool) {
                pj_pool_release(_pTransportIce->pTimerHeapPool);
                _pTransportIce->pTimerHeapPool = NULL;
        }
        
        if (_pTransportIce->pNegotiationPool) {
                pj_pool_release(_pTransportIce->pNegotiationPool);
                _pTransportIce->pNegotiationPool = NULL;
        }
}

static pj_status_t createMediaEndpt(IN OUT PeerConnection * _pPeerConnection)
{
        if (_pPeerConnection->pMediaEndpt != NULL) {
                return PJ_SUCCESS;
        }
        
        pj_status_t status;
        
        pj_pool_t *pPool = pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 128, 128, NULL);
        ASSERT_RETURN_CHECK(pPool, pj_pool_create);
        _pPeerConnection->pMutexPool = pPool;
        status = pj_mutex_create(pPool, NULL, PJ_MUTEX_DEFAULT, &_pPeerConnection->pMutex);
        STATUS_CHECK(pj_mutex_create, status);

        status = pjmedia_endpt_create(_pPeerConnection->pPoolFactory, NULL, 1, &_pPeerConnection->pMediaEndpt);
        STATUS_CHECK(pjmedia_endpt_create, status);
        int nNoTelephoneEvent = 0;
        status = pjmedia_endpt_set_flag(_pPeerConnection->pMediaEndpt, PJMEDIA_ENDPT_HAS_TELEPHONE_EVENT_FLAG, (void *)&nNoTelephoneEvent);
        STATUS_CHECK(pjmedia_endpt_set_flag, status);
        return status;
}

static pj_status_t IceConfigIsValid(IN IceConfig *_pIceConfig)
{
        if (_pIceConfig == NULL) {
                return PJ_EINVAL;
        }
        if (_pIceConfig->userCallback == NULL) {
                MY_PJ_LOG(1, "not set userCallback");
                return PJ_EINVAL;
        }
        if (_pIceConfig->nComponents <= 0 || _pIceConfig->nComponents > 2) {
                MY_PJ_LOG(1, "ice component:%d", _pIceConfig->nComponents);
                return PJ_EINVAL;
        }
        if (_pIceConfig->turnHost[0] == '\0' && _pIceConfig->stunHost[0] == '\0') {
                MY_PJ_LOG(1, "not set stun and turn server");
                return PJ_EINVAL;
        }
        return PJ_SUCCESS;
}

void InitIceConfig(IN OUT IceConfig *_pIceConfig)
{
        pj_assert(_pIceConfig != NULL);
        pj_bzero(_pIceConfig, sizeof(IceConfig));
        
        _pIceConfig->nComponents = 2;
        _pIceConfig->nMaxHosts = 5;
        _pIceConfig->bRegular = 1;
        _pIceConfig->nKeepAlive = 300;
}

int InitPeerConnectoin(OUT PeerConnection ** _pPeerConnection, IN IceConfig *_pIceConfig)
{
        if (_pPeerConnection == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();
        pj_status_t status = IceConfigIsValid(_pIceConfig);
        if (status != PJ_SUCCESS) {
                MY_PJ_LOG(1, "invalid IceConfig");
                return status;
        }

        PeerConnection * pPeerConnection =  (PeerConnection *)malloc(sizeof(PeerConnection));
        ASSERT_RETURN_CHECK(pPeerConnection, malloc);
        
        pj_bzero(pPeerConnection, sizeof(PeerConnection));
        pj_caching_pool_init(&pPeerConnection->cachingPool, &pj_pool_factory_default_policy, 0);
        
        pPeerConnection->userIceConfig = *_pIceConfig;

        pPeerConnection->pPoolFactory = &pPeerConnection->cachingPool.factory;
        
        peerConnectInitIceConfig(pPeerConnection);
        
        for ( int i = 0; i < sizeof(pPeerConnection->nAvIndex) / sizeof(int); i++) {
                pPeerConnection->nAvIndex[i] = -1;
        }
        *_pPeerConnection = pPeerConnection;
        InitMediaStream(&pPeerConnection->mediaStream);
        return 0;
}

int ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection)
{
        if (_pPeerConnection == NULL)
        {
                return PJ_SUCCESS;
        }
        LIBRTP_REGISTER_THREAD();

        _pPeerConnection->bQuit = 1;
        for ( int i = 0; i < sizeof(_pPeerConnection->nAvIndex) / sizeof(int); i++) {
                if (_pPeerConnection->nAvIndex[i] != -1) {
                        transportIceDestroy(&_pPeerConnection->transportIce[i]);
                }
        }
        
        if (_pPeerConnection->pMediaEndpt) {
                pjmedia_endpt_destroy(_pPeerConnection->pMediaEndpt);
                _pPeerConnection->pMediaEndpt = NULL;
        }
        
        if (_pPeerConnection->pNegPool) {
                pj_pool_release(_pPeerConnection->pNegPool);
                _pPeerConnection->pNegPool = NULL;
        }

        if (_pPeerConnection->pSdpPool) {
                pj_pool_release(_pPeerConnection->pSdpPool);
                _pPeerConnection->pSdpPool = NULL;
        }

        if (_pPeerConnection->pMutexPool) {
                pj_pool_release(_pPeerConnection->pMutexPool);
                _pPeerConnection->pMutexPool = NULL;
        }

        if (_pPeerConnection->pMutex) {
                pj_mutex_destroy(_pPeerConnection->pMutex);
                _pPeerConnection->pMutex = NULL;
        }
        
        for ( int i = 0 ; i < _pPeerConnection->mediaStream.nCount; i++) {
                pj_pool_t *pTmp = _pPeerConnection->mediaStream.streamTracks[i].pPacketizerPool;
                if (pTmp) {
                        pj_pool_release(pTmp);
                        _pPeerConnection->mediaStream.streamTracks[i].pPacketizerPool = NULL;
                }
        }

        pj_caching_pool_destroy (&_pPeerConnection->cachingPool);

        free(_pPeerConnection);

        return PJ_SUCCESS;
}

int AddAudioTrack(IN OUT PeerConnection * _pPeerConnection, IN MediaConfigSet * _pAudioConfig)
{
        if (_pPeerConnection == NULL || _pAudioConfig == NULL) {
                return PJ_EINVAL;
        }
        pj_status_t status;
        LIBRTP_REGISTER_THREAD();
        status = MediaConfigSetIsValid(_pAudioConfig);
        if (status != PJ_SUCCESS) {
                MY_PJ_LOG(1, "invalid MediaConfigSet");
                return status;
        }
        createMediaEndpt(_pPeerConnection);
        
        //TODO dupicated check
        int nAudioIndex = -1;
        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] == -1) {
                        nAudioIndex = i;
                        break;
                }
        }
        if(nAudioIndex == -1) {
                return -1;
        }
        
        AddMediaTrack(&_pPeerConnection->mediaStream, _pAudioConfig, nAudioIndex, RTP_STREAM_AUDIO, _pPeerConnection);
        _pPeerConnection->nAvIndex[nAudioIndex] = nAudioIndex;
        return PJ_SUCCESS;
}

int AddVideoTrack(IN OUT PeerConnection * _pPeerConnection, IN MediaConfigSet * _pVideoConfig)
{
        if (_pPeerConnection == NULL || _pVideoConfig == NULL) {
                return PJ_EINVAL;
        }
        pj_status_t status;
        LIBRTP_REGISTER_THREAD();
        status = MediaConfigSetIsValid(_pVideoConfig);
        if (status != PJ_SUCCESS) {
                MY_PJ_LOG(1, "invalid MediaConfigSet");
                return status;
        }

        createMediaEndpt(_pPeerConnection);

        //TODO dupicated check
        int nVideoIndex = -1;
        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] == -1) {
                        nVideoIndex = i;
                        break;
                }
        }
        if(nVideoIndex == -1) {
                return -1;
        }
        
        AddMediaTrack(&_pPeerConnection->mediaStream, _pVideoConfig, nVideoIndex, RTP_STREAM_VIDEO, _pPeerConnection);
        _pPeerConnection->nAvIndex[nVideoIndex] = nVideoIndex;
        return PJ_SUCCESS;
}

static int createMediaSdpMLine(IN pjmedia_endpt *_pMediaEndpt, IN pjmedia_transport *_pTransport,
                               IN pj_pool_t *_pPool, IN MediaStreamTrack *_pMediaTrack, OUT pjmedia_sdp_media **_pSdp)
{
        pjmedia_transport_info tpinfo;
        pjmedia_transport_info_init(&tpinfo);
        pjmedia_transport_get_info(_pTransport, &tpinfo);
        
        pj_status_t status = PJ_SUCCESS;
        if (_pMediaTrack->type == RTP_STREAM_AUDIO) {
                status = CreateSdpAudioMLine(_pMediaEndpt, &tpinfo, _pPool, _pMediaTrack, _pSdp);
        } else {
                status = CreateSdpVideoMLine(_pMediaEndpt, &tpinfo, _pPool, _pMediaTrack, _pSdp);
        }
        
        return status;
}

static int createSdpMline(IN OUT PeerConnection * _pPeerConnection, pj_pool_t *_pPool, pjmedia_sdp_session *pSdp)
{
        pj_status_t status;
        pjmedia_sdp_media * pAudioSdp = NULL;
        pjmedia_sdp_media * pVideoSdp = NULL;
        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] != -1) {
                        
                        if (_pPeerConnection->mediaStream.streamTracks[i].type == RTP_STREAM_AUDIO) {
                                status = createMediaSdpMLine(_pPeerConnection->pMediaEndpt, _pPeerConnection->transportIce[i].pTransport, _pPool,
                                                             &_pPeerConnection->mediaStream.streamTracks[i], &pAudioSdp);
                                STATUS_CHECK(pjmedia_endpt_create_audio_sdp, status);
                                pSdp->media[pSdp->media_count++] = pAudioSdp;
                        } else {
                                status = createMediaSdpMLine(_pPeerConnection->pMediaEndpt, _pPeerConnection->transportIce[i].pTransport, _pPool,
                                                             &_pPeerConnection->mediaStream.streamTracks[i], &pVideoSdp);
                                STATUS_CHECK(pjmedia_endpt_create_video_sdp, status);
                                pSdp->media[pSdp->media_count++] = pVideoSdp;
                        }
                }
        }
        
        
        print_sdp(pSdp, "basesdp with mline");
        return PJ_SUCCESS;
}

static int createBaseSdp(IN OUT PeerConnection * _pPeerConnection, IN pj_pool_t * _pPool, OUT pjmedia_sdp_session **_pOffer)
{
        pj_assert(_pPeerConnection && _pOffer);
        pj_assert(_pPeerConnection->pMediaEndpt);
        
        pj_str_t originStrAddr = pj_str("localhost");
        pj_sockaddr originAddr;
        pj_status_t status;
        status = pj_sockaddr_parse(pj_AF_INET(), 0, &originStrAddr, &originAddr);
        STATUS_CHECK(pj_sockaddr_parse, status);
        
        status = pjmedia_endpt_create_base_sdp(_pPeerConnection->pMediaEndpt, _pPool, NULL, &originAddr, _pOffer);
        STATUS_CHECK(pjmedia_endpt_create_base_sdp, status);

        print_sdp(*_pOffer, "basesdp");

        //createSdpMline(_pPeerConnection, _pPool, *_pOffer);
        
        return PJ_SUCCESS;
}

static void createSdpPool(IN OUT PeerConnection * _pPeerConnection)
{
        pj_pool_t *pPool = _pPeerConnection->pSdpPool;
        if(pPool == NULL) {
                pPool = pj_pool_create(&_pPeerConnection->cachingPool.factory,
                                       NULL, 4096, 1024, NULL);
                pj_assert(pPool != NULL);
                _pPeerConnection->pSdpPool = pPool;
        }
        
        return;
}

int createOffer(IN OUT PeerConnection * _pPeerConnection)
{
        pj_status_t  status;
        if (_pPeerConnection == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        pj_pool_t *pPool = _pPeerConnection->pSdpPool;
        if(pPool == NULL) {
                pPool = pj_pool_create(&_pPeerConnection->cachingPool.factory,
                                                  NULL, 4096, 1024, NULL);
                ASSERT_RETURN_CHECK(pPool, pj_pool_create);
                _pPeerConnection->pSdpPool = pPool;
        }

        if (_pPeerConnection->role != ICE_ROLE_NONE) {
                MY_PJ_LOG(2, "already created offer");
                return PJ_SUCCESS;
        }
        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] != -1) {
                        status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[_pPeerConnection->nAvIndex[i]]);
                        STATUS_CHECK(video initTransportIce, status);
                }
        }

        _pPeerConnection->role = ICE_ROLE_OFFERER;
        pjmedia_sdp_session *pOffer = NULL;
        status = createBaseSdp(_pPeerConnection, pPool, &pOffer);
        STATUS_CHECK(createSdp, status);

        _pPeerConnection->pOfferSdp = pOffer;
        
        return PJ_SUCCESS;
}

int createAnswer(IN OUT PeerConnection * _pPeerConnection, IN void *_pOffer)
{
        pj_status_t  status;
        if (_pPeerConnection == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        pj_pool_t *pPool = _pPeerConnection->pSdpPool;
        if(pPool == NULL) {
                pPool = pj_pool_create(&_pPeerConnection->cachingPool.factory,
                                                  NULL, 4096, 1024, NULL);
                ASSERT_RETURN_CHECK(pPool, pj_pool_create);
                _pPeerConnection->pSdpPool = pPool;
        }

        if (_pPeerConnection->role != ICE_ROLE_NONE) {
                MY_PJ_LOG(2, "already created answer");
                return PJ_SUCCESS;
        }
        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] != -1) {
                        status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[_pPeerConnection->nAvIndex[i]]);
                        STATUS_CHECK(video initTransportIce, status);
                }
        }

        pjmedia_sdp_session *pAnswer = NULL;
        _pPeerConnection->role = ICE_ROLE_ANSWERER;
        status = createBaseSdp(_pPeerConnection, pPool, &pAnswer);
        STATUS_CHECK(createSdp, status);

        _pPeerConnection->pAnswerSdp = pAnswer;
        
        return PJ_SUCCESS;
}


static void on_rx_rtcp(void *pUserData, void *pPkt, pj_ssize_t size)
{
        MediaStreamTrack *pMediaTrack = (MediaStreamTrack *)pUserData;
        
        if (size < 0) {
                MY_PJ_LOG(1, "Error receiving RTCP packet:%d", size);
                return;
        }
        
        /* Update RTCP session */
        pjmedia_rtcp_rx_rtcp(&pMediaTrack->rtcpSession, pPkt, size);
        return;
}

static void on_rx_rtp(void *pUserData, void *pPkt, pj_ssize_t size)
{
        pj_status_t status;
        const pjmedia_rtp_hdr *pRtpHeader;
        const void *pPayload;
        unsigned nPayloadLen;

        MediaStreamTrack *pMediaTrack = (MediaStreamTrack *)pUserData;

        /* Check for errors */
        if (size < 0) {
                MY_PJ_LOG(1, "RTP recv() error:%d", size);
                return;
        }

        /* Decode RTP packet. */
        status = pjmedia_rtp_decode_rtp(&pMediaTrack->rtpSession,
                                        pPkt, (int)size,
                                        &pRtpHeader, &pPayload, &nPayloadLen);
        if (status != PJ_SUCCESS) {
                MY_PJ_LOG(1, "RTP decode error:%d", status);
                return;
        }
        
        uint32_t nRtpTs = pj_ntohl(pRtpHeader->ts);
        MY_PJ_LOG(5, "-->receiveSize:%d  rtp seq:%d ts=%d", size, pj_ntohs(pRtpHeader->seq), nRtpTs);

        //MY_PJ_LOG(4, "Rx seq=%d", pj_ntohs(hdr->seq));
        /* Update the RTCP session. */
        pjmedia_rtcp_rx_rtp(&pMediaTrack->rtcpSession, pj_ntohs(pRtpHeader->seq),
                            nRtpTs, nPayloadLen);
        
        /* Update RTP session */
        pjmedia_rtp_session_update(&pMediaTrack->rtpSession, pRtpHeader, NULL);

        int nIsDiscard;
        JitterBufferPush(&pMediaTrack->jbuf, pPayload, nPayloadLen, pj_ntohs(pRtpHeader->seq),
                         nRtpTs, &nIsDiscard);
        if (nIsDiscard) {
                MY_PJ_LOG(2, "rtp packet disacrded by jitter buffer");
                return;
        }

        if (pMediaTrack->nMostLastRecvTimeAcc !=0 &&
            (nRtpTs < pMediaTrack->nLastRecvPktTimestamp && pMediaTrack->nLastRecvPktTimestamp - nRtpTs > 1000000000)) {
                pMediaTrack->nMostLastRecvTimeAcc = pMediaTrack->nLastRecvPktTimestamp;
        }

        pj_bool_t bGetFrame = PJ_TRUE;
        int nTestCnt = 0;
        while(bGetFrame) {
                JBFrameStatus popFrameType;
                int nSeq = 0;
                pj_uint32_t nTs = 0;
                int nFrameSize = 0;
                
                nFrameSize = sizeof(pMediaTrack->jbuf.getBuf);
                JitterBufferPop(&pMediaTrack->jbuf, pMediaTrack->jbuf.getBuf,
                                &nFrameSize, &nSeq, &nTs, &popFrameType);

                switch (popFrameType) {
                        case JBFRAME_STATE_MISSING:
                                pPayload = NULL;
                                nPayloadLen = 0;
                                bGetFrame = PJ_FALSE;
                                break;
                        case JBFRAME_STATE_CACHING:
                        case JBFRAME_STATE_EMPTY:
                                bGetFrame = PJ_FALSE;
                                return;
                        case JBFRAME_STATE_NORMAL:
                                pPayload = pMediaTrack->jbuf.getBuf;
                                nPayloadLen = nFrameSize;
                                bGetFrame = PJ_TRUE;
                                break;
                }
                if (!bGetFrame) {
                        break;
                }

                MY_PJ_LOG(4, "%d-->get_frame:%d  rtp seq:%d, ts=%d", ++nTestCnt, nPayloadLen, nSeq, nTs);

                //deal with payload
                pj_bool_t bTryAgain = PJ_FALSE;
                do{
                        pj_uint8_t *pBitstream = NULL;
                        unsigned nBitstreamPos = 0;
                        status = MediaUnPacketize(pMediaTrack->pMediaPacketier, pPayload, nPayloadLen, &pBitstream, &nBitstreamPos, pRtpHeader->m, &bTryAgain);
                        if (nBitstreamPos == 0) {
                                //MY_PJ_LOG(3, "MediaUnPacketize:%d, receiveSize:%d", status, size);
                                break;
                        }

                        RtpPacket rtpPacket;
                        pj_bzero(&rtpPacket, sizeof(rtpPacket));
                        rtpPacket.type = pMediaTrack->type;
                        rtpPacket.pData = pBitstream;
                        rtpPacket.nDataLen = nBitstreamPos;

                        int nIdx = pMediaTrack->mediaConfig.nUseIndex;
                        MediaConfig * pAvParam = &pMediaTrack->mediaConfig.configs[nIdx];
                        rtpPacket.format = pAvParam->codecType;
                        uint64_t nTs64 = nTs + pMediaTrack->nMostLastRecvTimeAcc;
                        rtpPacket.nTimestamp = nTs64 * 1000 / pAvParam->nSampleOrClockRate;

                        //MY_PJ_LOG(5, "rtp data receive:%ld, payLen:%d", size, nPayloadLen);
                        PeerConnection * pPeerConnection = (PeerConnection *)pMediaTrack->pPeerConnection;
                        pPeerConnection->userIceConfig.userCallback(pPeerConnection->userIceConfig.pCbUserData,
                                                                    CALLBACK_RTP,
                                                                    &rtpPacket);
                }while(bTryAgain);
        };
        
        return;
}

static int negotiationSettingAfterSuccess(IN PeerConnection * _pPeerConnection)
{
        pj_status_t status;

        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] == -1) {
                        continue;
                }
                TransportIce *pTransportIce = &_pPeerConnection->transportIce[i];

                pjmedia_transport_info tpinfo;
                pjmedia_transport_info_init(&tpinfo);
                pjmedia_transport_get_info(pTransportIce->pTransport, &tpinfo);
                
                status = pjmedia_transport_attach(pTransportIce->pTransport,
                                                  &_pPeerConnection->mediaStream.streamTracks[i],
                                                  &tpinfo.sock_info.rtp_addr_name,
                                                  &tpinfo.sock_info.rtcp_addr_name,
                                                  sizeof(tpinfo.sock_info.rtp_addr_name),
                                                  on_rx_rtp, //void (*rtp_cb)(void *user_data, void *pkt,pj_ssize_t),
                                                  on_rx_rtcp //void (*rtcp_cb)(void *usr_data,void*pkt,pj_ssize_t)
                                                  );
                STATUS_CHECK(pjmedia_transport_attach, status);
                
                //init rtp sesstoin
                MediaStreamTrack * pMediaTrack = &_pPeerConnection->mediaStream.streamTracks[i];
                int nIdx = pMediaTrack->mediaConfig.nUseIndex;
                int nRtpDynamicType = pMediaTrack->mediaConfig.configs[nIdx].codecType;
                pjmedia_rtp_session_init(&pMediaTrack->rtpSession, nRtpDynamicType, pj_rand());
                
                int nSampleOrClockRate = pMediaTrack->mediaConfig.configs[nIdx].nSampleOrClockRate;
                pjmedia_rtcp_init(&pMediaTrack->rtcpSession, NULL, nSampleOrClockRate,
                                  160, //TODO Average number of samples per frame. I don't know???
                                  //How do I set it if payload is video
                                  0);
                
                status = createJitterBuffer(pMediaTrack, _pPeerConnection->pPoolFactory);
                STATUS_CHECK(createJitterBuffer_pjmedia_jbuf_create, status);
        }

        return PJ_SUCCESS;
}

/*
 * will init rtp rtcp session is negotiation ok
 */
int StartNegotiation(IN PeerConnection * _pPeerConnection)
{
        if (_pPeerConnection == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
        for ( int i = 0; i < nMaxTracks; i++) {
                if (_pPeerConnection->nAvIndex[i] != -1) {
                        TransportIce *pTransportIce = &_pPeerConnection->transportIce[i];
                        pj_pool_t * pIceNegPool = pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 512, 512, NULL);
                        ASSERT_RETURN_CHECK(pIceNegPool, pj_pool_create);
                        pTransportIce->pNegotiationPool = pIceNegPool;
                        pj_status_t status = pjmedia_transport_media_start(pTransportIce->pTransport, pIceNegPool,
                                                               _pPeerConnection->pLocalSdp, _pPeerConnection->pRemoteSdp, i);
                        STATUS_CHECK(pjmedia_transport_media_start, status);
                }
        }
        
        return PJ_SUCCESS;
}

static int checkAndNeg(IN OUT PeerConnection * _pPeerConnection)
{
        pj_assert(_pPeerConnection->role != ICE_ROLE_NONE);
        
        if (_pPeerConnection->pNegPool == NULL) {
                _pPeerConnection->pNegPool =  pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 512, 512, NULL);
                ASSERT_RETURN_CHECK(_pPeerConnection->pNegPool, pj_pool_create);
        }
        
        pj_status_t status;
        
        if (_pPeerConnection->role == ICE_ROLE_OFFERER) {
                if (_pPeerConnection->pIceNeg == NULL) {
                        
                        status = pjmedia_sdp_neg_create_w_local_offer (_pPeerConnection->pNegPool,
                                                                       _pPeerConnection->pLocalSdp, &_pPeerConnection->pIceNeg);
                        STATUS_CHECK(pjmedia_sdp_neg_create_w_local_offer, status);
                        status = pjmedia_sdp_neg_set_remote_answer (_pPeerConnection->pNegPool,
                                                                    _pPeerConnection->pIceNeg, _pPeerConnection->pRemoteSdp);
                        STATUS_CHECK(pjmedia_sdp_neg_set_remote_answer, status);
                }
        } else if (_pPeerConnection->role == ICE_ROLE_ANSWERER) {
                status = pjmedia_sdp_neg_create_w_remote_offer(_pPeerConnection->pNegPool,
                                                               _pPeerConnection->pRemoteSdp, _pPeerConnection->pLocalSdp,
                                                               &_pPeerConnection->pIceNeg);
                STATUS_CHECK(pjmedia_sdp_neg_create_w_remote_offer, status);
        }
        
        status = pjmedia_sdp_neg_negotiate (_pPeerConnection->pNegPool, _pPeerConnection->pIceNeg, 0);
        STATUS_CHECK(pjmedia_sdp_neg_set_remote_answer, status);
        
        // which codec is agree
        const pjmedia_sdp_session * pActiveSdp = NULL;

        status = pjmedia_sdp_neg_get_active_local(_pPeerConnection->pIceNeg, &pActiveSdp);
        STATUS_CHECK(pjmedia_sdp_neg_get_active_local, status);

        status = SetActiveCodec(&_pPeerConnection->mediaStream, pActiveSdp);
        STATUS_CHECK(pjmedia_sdp_neg_get_active_remote, status);
        _pPeerConnection->iceNegInfo.nCount = pActiveSdp->media_count;
        for (int i = 0; i < _pPeerConnection->iceNegInfo.nCount; i++) {
                int nTmp =_pPeerConnection->mediaStream.streamTracks[i].mediaConfig.nUseIndex;
                _pPeerConnection->iceNegInfo.configs[i] = &_pPeerConnection->mediaStream.streamTracks[i].mediaConfig.configs[nTmp];
        }
        
        MediaStreamTrack *pVideoTrack = GetVideoTrack(&_pPeerConnection->mediaStream);
        if (pVideoTrack) {
                int nIdx = pVideoTrack->mediaConfig.nUseIndex;
                if (pVideoTrack->mediaConfig.configs[nIdx].codecType == MEDIA_FORMAT_H264){

                        pVideoTrack->pPacketizerPool = pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 200*1024, 200*1024, NULL);
                        status = CreatePacketizer("h264", 4, pVideoTrack->pPacketizerPool, &pVideoTrack->pMediaPacketier);
                        STATUS_CHECK(createPacketizer, status);
                }
        }
        
        MediaStreamTrack *pAudioTrack = GetAudioTrack(&_pPeerConnection->mediaStream);
        if (pAudioTrack) {
                int nIdx = pAudioTrack->mediaConfig.nUseIndex;
                if (pAudioTrack->mediaConfig.configs[nIdx].codecType == MEDIA_FORMAT_PCMU ||
                    pAudioTrack->mediaConfig.configs[nIdx].codecType == MEDIA_FORMAT_PCMA){

                        pAudioTrack->pPacketizerPool = pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 512, 512, NULL);
                        status = CreatePacketizer("pcmu", 4, pAudioTrack->pPacketizerPool, &pAudioTrack->pMediaPacketier);
                        STATUS_CHECK(createPacketizer, status);
                }
        }
#ifdef AUTO_NEGOTIATION
        StartNegotiation(_pPeerConnection);
#endif
        return PJ_SUCCESS;
}

int setLocalDescription(IN OUT PeerConnection * _pPeerConnection, IN void * _pSdp)
{
        if (_pPeerConnection == NULL || _pSdp == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        createSdpPool(_pPeerConnection);
        pjmedia_sdp_session *  pSdp = (pjmedia_sdp_session *) _pSdp;
        if (pSdp != _pPeerConnection->pOfferSdp && pSdp != _pPeerConnection->pAnswerSdp) {
                _pPeerConnection->pLocalSdp = pjmedia_sdp_session_clone(_pPeerConnection->pSdpPool, _pSdp);
                if (_pPeerConnection->pLocalSdp == NULL) {
                        MY_PJ_LOG(1, "PJ_NO_MEMORY_EXCEPTION, clone sdp fail");
                        pj_assert(_pPeerConnection->pLocalSdp != NULL);
                }
        } else {
                _pPeerConnection->pLocalSdp = pSdp;
        }

        if (_pPeerConnection->pRemoteSdp) {
                return checkAndNeg(_pPeerConnection);
        }
        return PJ_SUCCESS;
}

int setRemoteDescription(IN OUT PeerConnection * _pPeerConnection, IN void * _pSdp)
{
        if (_pPeerConnection == NULL || _pSdp == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        createSdpPool(_pPeerConnection);
        pjmedia_sdp_session *  pSdp = (pjmedia_sdp_session *) _pSdp;
        if (pSdp != _pPeerConnection->pOfferSdp && pSdp != _pPeerConnection->pAnswerSdp) {
                _pPeerConnection->pRemoteSdp = pjmedia_sdp_session_clone(_pPeerConnection->pSdpPool, _pSdp);
                if (_pPeerConnection->pRemoteSdp == NULL) {
                        MY_PJ_LOG(1, "PJ_NO_MEMORY_EXCEPTION, clone sdp fail");
                        pj_assert(_pPeerConnection->pLocalSdp != NULL);
                }
        } else {
                _pPeerConnection->pRemoteSdp = pSdp;
        }

        if(_pPeerConnection->pLocalSdp){
                return checkAndNeg(_pPeerConnection);
        }
        return PJ_SUCCESS;
}

static inline void firstScheduleNextSendRtcpTime(MediaStreamTrack * _pMediaTrack, pj_timestamp _now)
{
        // init timestamp
        if(_pMediaTrack->hzPerSecond.u64 == 0){
                pj_get_timestamp_freq(&_pMediaTrack->hzPerSecond);
                _pMediaTrack->nextRtcpTimestamp = _now;
                _pMediaTrack->nextRtcpTimestamp.u64 += (_pMediaTrack->hzPerSecond.u64 * (RTCP_INTERVAL+(pj_rand()%RTCP_RAND)) / 1000);
        }
}

static pj_status_t checkAndSendRtcp(MediaStreamTrack *_pMediaTrack, TransportIce *_pTransportIce, pj_timestamp _now)
{
        firstScheduleNextSendRtcpTime(_pMediaTrack, _now);
        if (_now.u64 >= _pMediaTrack->nextRtcpTimestamp.u64) {
                if (_pMediaTrack->nextRtcpTimestamp.u64 <= _now.u64) {
                        void *pRtcpPkt;
                        int nRtcpLen;
                        pj_ssize_t size;
                        pj_status_t status;

                        /* Build RTCP packet */
                        pjmedia_rtcp_build_rtcp(&_pMediaTrack->rtcpSession, &pRtcpPkt, &nRtcpLen);

                        /* Send packet */
                        size = nRtcpLen;
                        status = pjmedia_transport_send_rtcp(_pTransportIce->pTransport,
                                                             pRtcpPkt, size);
                        STATUS_CHECK(pjmedia_transport_send_rtcp, status);

                        /* Schedule next send */
                        _pMediaTrack->nextRtcpTimestamp.u64 = _now.u64 + (_pMediaTrack->hzPerSecond.u64 * (RTCP_INTERVAL+(pj_rand()%RTCP_RAND)) / 1000);
                }
        }
        return PJ_SUCCESS;
}

static inline uint64_t getTimestampGapFromLastPacket(IN MediaStreamTrack *_pMediaTrack, uint64_t _timestamp)
{
        MY_PJ_LOG(5, "nLastSendPktTimestamp1:%lld, %d", _pMediaTrack->nLastSendPktTimestamp,_pMediaTrack->nLastSendPktTimestamp == ULLONG_MAX);
        if (_pMediaTrack->nLastSendPktTimestamp == ULLONG_MAX) {
                MY_PJ_LOG(5, "nLastSendPktTimestamp:%lld", _pMediaTrack->nLastSendPktTimestamp);
                _pMediaTrack->nLastSendPktTimestamp = _timestamp;
                return TS_BASE_VALUE;
        }
        uint64_t diff = _timestamp - _pMediaTrack->nLastSendPktTimestamp;
        _pMediaTrack->nLastSendPktTimestamp = _timestamp;
        return diff;
}

static inline uint64_t getMediaTrackElapseTime(IN MediaStreamTrack *_pMediaTrack, uint64_t _timestamp)
{
        return _timestamp - _pMediaTrack->nFirstSendPktTimestamp;
}

// _nPktTimestampGap millisecond?
static inline uint32_t calcRtpTimestampLen(uint64_t _nPktTimestampGap, int nRate)
{
        uint32_t rate = (uint32_t)nRate;
        return _nPktTimestampGap * nRate / 1000;
}

static void dealWithTimestamp(IN OUT MediaStreamTrack *_pMediaTrack, IN pj_timestamp _now,
                              IN int _nRate, IN RtpPacket * _pPacket, IN uint32_t *pRtpTsLen)
{
        uint64_t nPktTimestampGap = getTimestampGapFromLastPacket(_pMediaTrack, _pPacket->nTimestamp);
        
        *pRtpTsLen = calcRtpTimestampLen(nPktTimestampGap, _nRate);
        //MY_PJ_LOG(5, "tslen:%d pkt:%lld gap:%lld now:%lld", *pRtpTsLen, _pPacket->nTimestamp, nPktTimestampGap, _now.u64);
        
        uint32_t nElapse = getMediaTrackElapseTime(_pMediaTrack, _pPacket->nTimestamp);
        pj_timestamp exptectNow;
        exptectNow.u64 = _pMediaTrack->nSysTimeBase.u64 + nElapse * _pMediaTrack->hzPerSecond.u64 / 1000;
        
        pj_uint64_t nLate = 0;
        if (exptectNow.u64 > _now.u64) {
                nLate = ((exptectNow.u64 - _now.u64) * 1000) / _pMediaTrack->hzPerSecond.u64;
                if ( nLate > 1) {
                        MY_PJ_LOG(5, "audio data late:%lld-%lld=%lld",exptectNow.u64, _now.u64, nLate);
                }
        } else {
                nLate = ((_now.u64 - exptectNow.u64) * 1000) / _pMediaTrack->hzPerSecond.u64;
                if ( nLate > 1) {
                        MY_PJ_LOG(5, "audio data early:%lld-%lld=%lld",_now.u64, exptectNow.u64, nLate);
                }
        }
}

static pj_status_t sendPacket(IN OUT MediaStreamTrack *_pMediaTrack, IN TransportIce * _pTransportIce,
                              IN int _nRtpType, IN int _nMarker, IN int _nRtpTsLen, IN const void *_pData, IN int _nDataLen)
{
        //start to send rtp
        pj_status_t status;
        const void *pVoidHeader;
        const pjmedia_rtp_hdr *pRtpHeader;
        pj_ssize_t size;
        int nHeaderLen;
        
        /* Format RTP header */
        status = pjmedia_rtp_encode_rtp( &_pMediaTrack->rtpSession, _nRtpType,
                                        _nMarker,
                                        _nDataLen,
                                        _nRtpTsLen,
                                        &pVoidHeader, &nHeaderLen);
        STATUS_CHECK(pjmedia_rtp_encode_rtp, status);
        
        pRtpHeader = (const pjmedia_rtp_hdr*) pVoidHeader;
        MY_PJ_LOG(5, "send data(%d) len:%d with seq=%d ts=%d tsLen=%d", _nRtpType, _nDataLen,
                  pj_ntohs(pRtpHeader->seq), pj_ntohl(pRtpHeader->ts), _nRtpTsLen);
        
        char packet[1500];
        /* Copy RTP header to packet */
        pj_memcpy(packet, pRtpHeader, nHeaderLen);
        
        /* Zero the payload */
        pj_memcpy(packet+nHeaderLen, _pData, _nDataLen);
        
        /* Send RTP packet */
        size = nHeaderLen + _nDataLen;
        status = pjmedia_transport_send_rtp(_pTransportIce->pTransport,
                                            packet, size);
        STATUS_CHECK(pjmedia_transport_send_rtp, status);
        
        /* Update RTCP SR */
        pjmedia_rtcp_tx_rtp( &_pMediaTrack->rtcpSession, _nDataLen);
        
        return PJ_SUCCESS;
}

static int SendAudioPacket(IN PeerConnection *_pPeerConnection, IN RtpPacket * _pPacket)
{
        MediaStreamTrack * pMediaTrack = GetAudioTrack(&_pPeerConnection->mediaStream);
        if (pMediaTrack == NULL) {
                MY_PJ_LOG(1, "no audio track in stream");
                return -1;
        }
        int nTransportIndex = GetMediaTrackIndex(&_pPeerConnection->mediaStream, pMediaTrack);
        if (nTransportIndex < 0){
                MY_PJ_LOG(1, "no found match track in stream");
                return -2;
        }
        TransportIce * pTransportIce = &_pPeerConnection->transportIce[nTransportIndex];
        
        MediaConfigSet *pAudioConfig = &pMediaTrack->mediaConfig;
        int nIdx = pMediaTrack->mediaConfig.nUseIndex;
        int nSampleRate = pAudioConfig->configs[nIdx].nSampleOrClockRate;
        int nRtpType = pAudioConfig->configs[nIdx].codecType;
        //int nChannel = pAudioConfig->configs[nIdx].nChannel;
        //int nBitDepth = pAudioConfig->configs[nIdx].nBitDepth;
        //unsigned nMsecInterval = _pPacket->nDataLen * 1000 /nChannel / (nBitDepth / 8) / nSampleRate;

        pj_timestamp now;
        pj_get_timestamp(&now);
        
        if (pMediaTrack->nSysTimeBase.u64 == 0) {
                pMediaTrack->nSysTimeBase = now;
                pMediaTrack->nFirstSendPktTimestamp = _pPacket->nTimestamp;
        }

        checkAndSendRtcp(pMediaTrack, pTransportIce, now);

        uint32_t nRtpTsLen = 0;
        dealWithTimestamp(pMediaTrack, now, nSampleRate, _pPacket, &nRtpTsLen);

        pj_status_t status;
        status =  sendPacket(pMediaTrack, pTransportIce, nRtpType, 0, nRtpTsLen, _pPacket->pData, _pPacket->nDataLen);
        STATUS_CHECK(pjmedia_rtp_encode_rtp, status);

        return 0;
}

static int SendVideoPacket(IN PeerConnection *_pPeerConnection, IN OUT RtpPacket * _pPacket)
{
        MediaStreamTrack *pMediaTrack = GetVideoTrack(&_pPeerConnection->mediaStream);
        int nTransportIndex = GetMediaTrackIndex(&_pPeerConnection->mediaStream, pMediaTrack);
        if (nTransportIndex < 0){
                MY_PJ_LOG(1, "no found match track in stream");
                return -2;
        }
        TransportIce * pTransportIce = &_pPeerConnection->transportIce[nTransportIndex];

        MediaConfigSet *pVideoConfig = &pMediaTrack->mediaConfig;
        int nIdx = pMediaTrack->mediaConfig.nUseIndex;
        int nClockRate = pVideoConfig->configs[nIdx].nSampleOrClockRate;
        int nRtpType = pVideoConfig->configs[nIdx].codecType;

        pj_timestamp now;
        pj_get_timestamp(&now);

        if (pMediaTrack->nSysTimeBase.u64 == 0) {
                pMediaTrack->nSysTimeBase = now;
                pMediaTrack->nFirstSendPktTimestamp = _pPacket->nTimestamp;
        }

        checkAndSendRtcp(pMediaTrack, pTransportIce, now);

        uint32_t nRtpTsLen = 0;
        dealWithTimestamp(pMediaTrack, now, nClockRate, _pPacket, &nRtpTsLen);

        int nLeft = _pPacket->nDataLen;
        unsigned nOffset = 0;
        const pj_uint8_t * pPayload;
        pj_size_t nPayloadLen;
        unsigned nBitsPos;
        int nTsLlen = nRtpTsLen;

        while (nLeft != 0) {

                pPayload = NULL;
                nPayloadLen = 0;
                nBitsPos = 0;

                pj_status_t status;
                status = MediaPacketize(pMediaTrack->pMediaPacketier,
                                                (pj_uint8_t *)_pPacket->pData + nOffset,
                                                nLeft,
                                                &nBitsPos,
                                                &pPayload,
                                                &nPayloadLen
                                                );
                STATUS_CHECK(pjmedia_h264_packetize, status);
                nLeft -= nBitsPos;
                nOffset += nBitsPos;
                
                
                int marker = 0;
                if (nOffset == _pPacket->nDataLen && nOffset != nBitsPos){
                        marker = 1;
                }

                status =  sendPacket(pMediaTrack, pTransportIce, nRtpType, marker, nTsLlen, pPayload, nPayloadLen);
                STATUS_CHECK(pjmedia_rtp_encode_rtp, status);

                nTsLlen = 0;
        }
        
        return PJ_SUCCESS;
}

int SendRtpPacket(IN PeerConnection *_pPeerConnection, IN OUT RtpPacket * _pPacket)
{
        if (_pPeerConnection == NULL || _pPacket == NULL) {
                return PJ_EINVAL;
        }
        LIBRTP_REGISTER_THREAD();

        if (_pPacket->type == RTP_STREAM_AUDIO) {
                return SendAudioPacket(_pPeerConnection, _pPacket);
        } else {
                return SendVideoPacket(_pPeerConnection, _pPacket);
        }
}
