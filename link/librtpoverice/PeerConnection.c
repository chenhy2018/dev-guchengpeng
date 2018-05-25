#include "PeerConnection.h"

static int waitState(IN TransportIce *_pTransportIce, IN IceState currentState)
{
    int nCnt = 0;
    do{
        if (nCnt > 5) {
            return 1; //fail
        }
        nCnt++;
        pj_thread_sleep(500);
    }while(_pTransportIce->iceState == currentState);
    
    return 0;
}


static void onIceComplete2(pjmedia_transport *tp, pj_ice_strans_op op,
                      pj_status_t status, void *user_data) {
    TransportIce *pTransportIce = (TransportIce *)user_data;
    
    if(status != PJ_SUCCESS){
        pTransportIce->iceState = ICE_STATE_FAIL;
        return;
    }
    //pTransportIce->iceState =  op;
    switch (op) {
            /** Initialization (candidate gathering) */
        case PJ_ICE_STRANS_OP_INIT:
            pTransportIce->iceState = ICE_STATE_GATHERING_OK;
            printf("--->gathering candidates finish\n");
            break;
            
            /** Negotiation */
        case PJ_ICE_STRANS_OP_NEGOTIATION:
            printf("--->PJ_ICE_STRANS_OP_NEGOTIATION\n");
            pTransportIce->iceState = ICE_STATE_NEGOTIATION_OK;
            break;
            
            /** This operation is used to report failure in keep-alive operation.
             *  Currently it is only used to report TURN Refresh failure.  */
        case PJ_ICE_STRANS_OP_KEEP_ALIVE:
            printf("--->PJ_ICE_STRANS_OP_KEEP_ALIVE\n");
            break;
            
            /** IP address change notification from STUN keep-alive operation.  */
        case PJ_ICE_STRANS_OP_ADDR_CHANGE:
            printf("--->PJ_ICE_STRANS_OP_ADDR_CHANGE\n");
            break;
    }
}

static int iceWorkerThread(void * _pArg)
{
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
        
        
        if (pUserIceConfig->nMaxHosts > 0 || pUserIceConfig->stunHost[0] != '\0') {
            pIceCfg->stun_tp_cnt = 1;
            pj_ice_strans_stun_cfg_default(&pIceCfg->stun_tp[0]);
            
            pIceCfg->stun_tp[0].max_host_cands = pUserIceConfig->nMaxHosts;
            if (pUserIceConfig->stunHost[0] != '\0') {
                pIceCfg->stun_tp[0].server = pj_str(pUserIceConfig->stunHost);
            }
            if (pUserIceConfig->nKeepAlive > 0) {
                pIceCfg->stun_tp[0].cfg.ka_interval = pUserIceConfig->nKeepAlive;
            }
        }
        
        
        if (pUserIceConfig->turnHost[0] != '\0') {
            pIceCfg->turn_tp_cnt = 1;
            pj_ice_strans_turn_cfg_default(&pIceCfg->turn_tp[0]);
            
            pIceCfg->turn_tp[0].server = pj_str(pUserIceConfig->turnHost);
            pIceCfg->turn_tp[0].port = PJ_STUN_PORT;
            
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
    status = pj_thread_create(pThreadPool, NULL, &iceWorkerThread, _pTransportIce, 0, 0, &pThread);
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
    if (_pTransportIce->pTransport) {
        pjmedia_transport_media_stop(_pTransportIce->pTransport);
        _pTransportIce->pTransport = NULL;
    }
    
    if (_pTransportIce->pPollThread) {
        pj_thread_join(_pTransportIce->pPollThread);
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
    
    status = pjmedia_endpt_create(_pPeerConnection->pPoolFactory, NULL, 1, &_pPeerConnection->pMediaEndpt);
    STATUS_CHECK(pjmedia_endpt_create, status);
    int nNoTelephoneEvent = 0;
    status = pjmedia_endpt_set_flag(_pPeerConnection->pMediaEndpt, PJMEDIA_ENDPT_HAS_TELEPHONE_EVENT_FLAG, (void *)&nNoTelephoneEvent);
    STATUS_CHECK(pjmedia_endpt_set_flag, status);
    return status;
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

void InitIceConfig(IN OUT IceConfig *_pIceConfig)
{
    pj_bzero(_pIceConfig, sizeof(IceConfig));
    
    _pIceConfig->nComponents = 2;
    _pIceConfig->nMaxHosts = 0;
    _pIceConfig->bRegular = 1;
    _pIceConfig->nKeepAlive = 300;
}

void InitPeerConnectoin(IN OUT PeerConnection * _pPeerConnection, IN IceConfig *_pIceConfig, IN pj_pool_factory * _pPoolFactory)
{
    pj_memset(_pPeerConnection, 0, sizeof(PeerConnection));
    
    _pPeerConnection->userIceConfig = *_pIceConfig;
    _pPeerConnection->pPoolFactory = _pPoolFactory;
    
    peerConnectInitIceConfig(_pPeerConnection);
    
    for ( int i = 0; i < sizeof(_pPeerConnection->nAvIndex) / sizeof(int); i++) {
        _pPeerConnection->nAvIndex[i] = -1;
    }
    
    return ;
}

void ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection)
{
    
    for ( int i = 0; i < sizeof(_pPeerConnection->nAvIndex) / sizeof(int); i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            transportIceDestroy(&_pPeerConnection->transportIce[i]);
        }
    }
    
    if (_pPeerConnection->pMediaEndpt) {
        pjmedia_endpt_destroy(_pPeerConnection->pMediaEndpt);
        _pPeerConnection->pMediaEndpt = NULL;
    }
}

int AddAudioTrack(IN OUT PeerConnection * _pPeerConnection, IN MediaConfig * _pAudioConfig)
{
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

    pj_status_t status;
    status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[nAudioIndex]);
    STATUS_CHECK(audio initTransportIce, status);

    AddMediaTrack(&_pPeerConnection->mediaStream, _pAudioConfig, nAudioIndex, TYPE_AUDIO);
    _pPeerConnection->nAvIndex[nAudioIndex] = nAudioIndex;
    return PJ_SUCCESS;
}

int AddVideoTrack(IN OUT PeerConnection * _pPeerConnection, IN MediaConfig * _pVideoConfig)
{
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
    
    pj_status_t status;
    status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[nVideoIndex]);
    STATUS_CHECK(video initTransportIce, status);
    
    AddMediaTrack(&_pPeerConnection->mediaStream, _pVideoConfig, nVideoIndex, TYPE_VIDEO);
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
    if (_pMediaTrack->type == TYPE_AUDIO) {
        status = CreateSdpAudioMLine(_pMediaEndpt, &tpinfo, _pPool, _pMediaTrack, _pSdp);
    } else {
        status = CreateSdpVideoMLine(_pMediaEndpt, &tpinfo, _pPool, _pMediaTrack, _pSdp);
    }
    
    return status;
}

static int createSdp(IN OUT PeerConnection * _pPeerConnection, IN pj_pool_t * _pPool, OUT pjmedia_sdp_session **_pOffer)
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
    
    pjmedia_sdp_session *pSdp = *_pOffer;
    
    pjmedia_sdp_media * pAudioSdp = NULL;
    pjmedia_sdp_media * pVideoSdp = NULL;
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            if (waitState(&_pPeerConnection->transportIce[i], ICE_STATE_INIT)) {
                PJ_LOG(3,(__FILE__, "wait ICE_STATE_GATHERING_OK timeout"));
                return -1;
            }
            
            if (_pPeerConnection->mediaStream.streamTracks[i].type == TYPE_AUDIO) {
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
    
    return PJ_SUCCESS;
}

int createOffer(IN OUT PeerConnection * _pPeerConnection, IN pj_pool_t * _pPool, OUT pjmedia_sdp_session **_pOffer)
{
    pj_status_t  status;
    
    status = createSdp(_pPeerConnection, _pPool, _pOffer);
    STATUS_CHECK(createSdp, status);
    
    char sdpStr[2048];
    memset(sdpStr, 0, 2048);
    pjmedia_sdp_print(*_pOffer, sdpStr, sizeof(sdpStr));
    printf("%s\n", sdpStr);
    
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            status = pjmedia_transport_media_create(_pPeerConnection->transportIce[i].pTransport, _pPool, 0, NULL, 0);
            STATUS_CHECK(pjmedia_transport_media_create, status);
            
            status = pjmedia_transport_encode_sdp(_pPeerConnection->transportIce[i].pTransport, _pPool, *_pOffer, NULL, i);
            STATUS_CHECK(pjmedia_transport_encode_sdp, status);
        }
    }
    
    memset(sdpStr, 0, 2048);
    pjmedia_sdp_print(*_pOffer, sdpStr, sizeof(sdpStr));
    printf("----------------\n%s\n", sdpStr);
    
    return PJ_SUCCESS;
}

int createAnswer(IN OUT PeerConnection * _pPeerConnection, IN pj_pool_t * _pPool,
                 IN pjmedia_sdp_session *_pOffer, OUT pjmedia_sdp_session **_pAnswer)
{
    pj_status_t  status;
    
    status = createSdp(_pPeerConnection, _pPool, _pAnswer);
    STATUS_CHECK(createSdp, status);
    
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            status = pjmedia_transport_media_create(_pPeerConnection->transportIce[i].pTransport, _pPool, 0, _pOffer, 0);
            STATUS_CHECK(pjmedia_transport_media_create, status);
            
            status = pjmedia_transport_encode_sdp(_pPeerConnection->transportIce[i].pTransport, _pPool, *_pAnswer, _pOffer, i);
            STATUS_CHECK(pjmedia_transport_encode_sdp, status);
        }
    }
    
    return PJ_SUCCESS;
}

void setLocalDescription(IN OUT PeerConnection * _pPeerConnectoin, IN pjmedia_sdp_session * _pLocalSdp)
{
    _pPeerConnectoin->pOfferSdp = _pLocalSdp;
}

void setRemoteDescription(IN OUT PeerConnection * _pPeerConnectoin, IN pjmedia_sdp_session * _pRemoteSdp)
{
    _pPeerConnectoin->pAnswerSdp = _pRemoteSdp;
}

/*
 * will init rtp rtcp session is negotiation ok
 */
int StartNegotiation(IN PeerConnection * _pPeerConnection)
{
    pj_status_t status;
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            TransportIce *pTransportIce = &_pPeerConnection->transportIce[i];
            pj_pool_t * pIceNegPool = pj_pool_create(_pPeerConnection->pPoolFactory, NULL, 512, 512, NULL);
            ASSERT_RETURN_CHECK(pIceNegPool, pj_pool_create);
            pTransportIce->pNegotiationPool = pIceNegPool;
            status = pjmedia_transport_media_start(pTransportIce->pTransport, pIceNegPool,
                                            _pPeerConnection->pOfferSdp, _pPeerConnection->pAnswerSdp, i);
            STATUS_CHECK(pjmedia_transport_media_start, status);
            
            if (waitState(&_pPeerConnection->transportIce[i], ICE_STATE_GATHERING_OK)){
                PJ_LOG(3,(__FILE__, "wait ICE_STATE_NEGOTIATION_OK timeout"));
                return -1;
            }

            pjmedia_transport_info tpinfo;
            pjmedia_transport_info_init(&tpinfo);
            pjmedia_transport_get_info(pTransportIce->pTransport, &tpinfo);
            
            status = pjmedia_transport_attach(pTransportIce->pTransport, NULL,
                                     &tpinfo.sock_info.rtp_addr_name,
                                     &tpinfo.sock_info.rtcp_addr_name,
                                     sizeof(tpinfo.sock_info.rtp_addr_name),
                                     NULL, //void (*rtp_cb)(void *user_data, void *pkt,pj_ssize_t),
                                     NULL //void (*rtcp_cb)(void *usr_data,void*pkt,pj_ssize_t)
                                     );
            STATUS_CHECK(pjmedia_transport_attach, status);

            //init rtp sesstoin
            MediaStreamTrack * pMediaTrack = &_pPeerConnection->mediaStream.streamTracks[i];
            pjmedia_rtp_session_init(&pMediaTrack->rtpSession, pMediaTrack->mediaConfig.nRtpDynamicType,
                                     pj_rand());

            pjmedia_rtcp_init(&pMediaTrack->rtcpSession, NULL, pMediaTrack->mediaConfig.nSampleOrClockRate,
                              160, 0); //TODO 160 instead by cacl
        }
    }

    return PJ_SUCCESS;
}

int SendAudio(IN PeerConnection *_pPeerConnection, uint8_t *_pData, int _nLen)
{
    enum { RTCP_INTERVAL = 5000, RTCP_RAND = 2000 };
    char packet[1500];

    MediaStreamTrack * pMediaTrack = GetAudioTrack(&_pPeerConnection->mediaStream);
    if (pMediaTrack == NULL) {
        PJ_LOG(3, (__FILE__, "no audio track in stream"));
        return -1;
    }
    int nTransportIndex = GetMediaTrackIndex(&_pPeerConnection->mediaStream, pMediaTrack);
    if (nTransportIndex < 0){
        PJ_LOG(3, (__FILE__, "no found match track in stream"));
        return -2;
    }
    TransportIce * pTransportIce = &_pPeerConnection->transportIce[nTransportIndex];

    MediaConfig *pAudioConfig = &pMediaTrack->mediaConfig;
    unsigned nMsecInterval = _nLen * 1000 / pAudioConfig->nChannel / (pAudioConfig->nBitDepth / 8) / pAudioConfig->nSampleOrClockRate;

    if(pMediaTrack->hzPerSecond.u64 == 0){
        pj_get_timestamp_freq(&pMediaTrack->hzPerSecond);

        pj_get_timestamp(&pMediaTrack->nextRtpTimestamp);
        pMediaTrack->nextRtpTimestamp.u64 += (pMediaTrack->hzPerSecond.u64 * nMsecInterval / 1000);

        pMediaTrack->nextRtcpTimestamp = pMediaTrack->nextRtpTimestamp;
        pMediaTrack->nextRtcpTimestamp.u64 += (pMediaTrack->hzPerSecond.u64 * (RTCP_INTERVAL+(pj_rand()%RTCP_RAND)) / 1000);
    }

    pj_timestamp now;
    if (pMediaTrack->nextRtpTimestamp.u64 >= pMediaTrack->nextRtcpTimestamp.u64) {
        pj_get_timestamp(&now);
        if (pMediaTrack->nextRtcpTimestamp.u64 <= now.u64) {
            void *pRtcpPkt;
            int nRtcpLen;
            pj_ssize_t size;
            pj_status_t status;

            /* Build RTCP packet */
            pjmedia_rtcp_build_rtcp(&pMediaTrack->rtcpSession, &pRtcpPkt, &nRtcpLen);

            /* Send packet */
            size = nRtcpLen;
            status = pjmedia_transport_send_rtcp(pTransportIce->pTransport,
                                                 pRtcpPkt, size);
            STATUS_CHECK(pjmedia_transport_send_rtcp, status);

            /* Schedule next send */
            pMediaTrack->nextRtcpTimestamp.u64 += (pMediaTrack->hzPerSecond.u64 * (RTCP_INTERVAL+(pj_rand()%RTCP_RAND)) / 1000);
        }
    }

    pj_timestamp lesser;
    pj_time_val timeout;

    lesser = pMediaTrack->nextRtpTimestamp;
    pj_get_timestamp(&now);

    /* Determine how long to sleep */
    if (lesser.u64 <= now.u64) {
        timeout.sec = timeout.msec = 0;
        //printf("immediate "); fflush(stdout);
    } else {
        pj_uint64_t tick_delay;
        tick_delay = lesser.u64 - now.u64;
        timeout.sec = 0;
        timeout.msec = (pj_uint32_t)(tick_delay * 1000 / pMediaTrack->hzPerSecond.u64);
        pj_time_val_normalize(&timeout);

        //printf("%d:%03d ", timeout.sec, timeout.msec); fflush(stdout);
    }
    printf("timeout:%ld %ld\n", timeout.sec, timeout.msec);
    pj_thread_sleep(PJ_TIME_VAL_MSEC(timeout)); //TODO deal sleep

    //start to send rtp
    pj_status_t status;
    const void *p_hdr;
    const pjmedia_rtp_hdr *hdr;
    pj_ssize_t size;
    int hdrlen;

    /* Format RTP header */
    status = pjmedia_rtp_encode_rtp( &pMediaTrack->rtpSession, 0, //pt is 0 for pcmu
                                    0, /* marker bit */
                                    160,
                                    160,
                                    &p_hdr, &hdrlen);
    STATUS_CHECK(pjmedia_rtp_encode_rtp, status);


    //PJ_LOG(4,(THIS_FILE, "\t\tTx seq=%d", pj_ntohs(hdr->seq)));
    hdr = (const pjmedia_rtp_hdr*) p_hdr;

    /* Copy RTP header to packet */
    pj_memcpy(packet, hdr, hdrlen);

    /* Zero the payload */
    pj_memcpy(packet+hdrlen, _pData, 160);

    /* Send RTP packet */
    size = hdrlen + 160;
    status = pjmedia_transport_send_rtp(pTransportIce->pTransport,
                                        packet, size);
    STATUS_CHECK(pjmedia_transport_send_rtp, status);



    /* Update RTCP SR */
    pjmedia_rtcp_tx_rtp( &pMediaTrack->rtcpSession, 160);

    /* Schedule next send */
    pMediaTrack->nextRtpTimestamp.u64 += (nMsecInterval * pMediaTrack->hzPerSecond.u64 / 1000);

    return 0;
}
