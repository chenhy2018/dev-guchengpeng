#include "PeerConnection.h"


static void onIceComplete2(pjmedia_transport *tp, pj_ice_strans_op op,
                      pj_status_t status, void *user_data) {
    PeerConnectoin *pPeerConnectoin = (PeerConnectoin *)user_data;
    
    if(status != PJ_SUCCESS){
        pPeerConnectoin->iceState = ICE_STATE_FAIL;
        return;
    }
    pPeerConnectoin->iceState =  op;
    switch (op) {
            /** Initialization (candidate gathering) */
        case PJ_ICE_STRANS_OP_INIT:
            pPeerConnectoin->iceState = ICE_STATE_GATHERING_OK;
            printf("--->gathering candidates finish\n");
            break;
            
            /** Negotiation */
        case PJ_ICE_STRANS_OP_NEGOTIATION:
            printf("--->PJ_ICE_STRANS_OP_NEGOTIATION\n");
            pPeerConnectoin->iceState = ICE_STATE_NEGOTIATION_OK;
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
    PeerConnectoin * pPeerConnection = (PeerConnectoin *)_pArg;
    pj_ice_strans_cfg * pIceCfg = &pPeerConnection->iceConfig;
    
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

static int peerConnectInitIceConfig(IN OUT PeerConnectoin * _pPeerConnectoin)
{
    pj_assert(_pPeerConnectoin->pMediaEndpt);
    
    pj_ice_strans_cfg * pIceCfg = &_pPeerConnectoin->iceConfig;
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
    return PJ_SUCCESS;
}

static pj_status_t initTransportIce(IN PeerConnectoin * _pPeerConnectoin, OUT TransportIce * _pTransportIce)
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
    
    
    pj_stun_config_init(&_pPeerConnectoin->iceConfig.stun_cfg, _pPeerConnectoin->pPoolFactory, 0,
                        pIoQueue, pTimerHeap);
    
    pj_thread_t * pThread;
    pj_pool_t * pThreadPool = pj_pool_create(_pPeerConnectoin->pPoolFactory, NULL, 2048, 1024, NULL);
    ASSERT_RETURN_CHECK(pThreadPool, pj_pool_create);
    _pTransportIce->pThreadPool = pThreadPool;
    status = pj_thread_create(pThreadPool, NULL, &iceWorkerThread, _pPeerConnectoin, 0, 0, &pThread);
    STATUS_CHECK(pj_thread_create, status);
    _pTransportIce->pPollThread = pThread;
    
    pjmedia_ice_cb cb;
    cb.on_ice_complete = NULL;
    cb.on_ice_complete2 = onIceComplete2;
    
    pjmedia_transport *transport = NULL;
    status = pjmedia_ice_create3(_pPeerConnectoin->pMediaEndpt, NULL, _pPeerConnectoin->userIceConfig.nComponents,
                                 &_pPeerConnectoin->iceConfig, &cb, 0, _pPeerConnectoin, &transport);
    STATUS_CHECK(pjmedia_ice_create3, status);
    pjmedia_ice_add_ice_cb(transport, &cb, _pPeerConnectoin);
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
}

void InitIceConfig(IN OUT IceConfig *_pIceConfig)
{
    pj_bzero(_pIceConfig, sizeof(IceConfig));
    
    _pIceConfig->nComponents = 2;
    _pIceConfig->nMaxHosts = 0;
    _pIceConfig->bRegular = 1;
    _pIceConfig->nKeepAlive = 300;
}

int InitPeerConnectoin(IN OUT PeerConnectoin * _pPeerConnection, IN IceConfig *_pIceConfig, IN pj_pool_factory * _pPoolFactory)
{
    pj_memset(_pPeerConnection, 0, sizeof(PeerConnectoin));
    
    _pPeerConnection->userIceConfig = *_pIceConfig;
    _pPeerConnection->pPoolFactory = _pPoolFactory;
    
    pj_status_t status;
    status = pjmedia_endpt_create(_pPeerConnection->pPoolFactory, NULL, 1, &_pPeerConnection->pMediaEndpt);
    STATUS_CHECK(pjmedia_endpt_create, status);

    pjmedia_endpt_set_flag(_pPeerConnection->pMediaEndpt, PJMEDIA_ENDPT_HAS_TELEPHONE_EVENT_FLAG, (void *)0);
    
    peerConnectInitIceConfig(_pPeerConnection);
    for ( int i = 0; i < sizeof(_pPeerConnection->nAvIndex) / sizeof(int); i++) {
        _pPeerConnection->nAvIndex[i] = -1;
    }
    
    return status;
}

void ReleasePeerConnectoin(IN OUT PeerConnectoin * _pPeerConnection)
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

int AddAudioTrack(IN OUT PeerConnectoin * _pPeerConnection, IN MediaConfig * _pAudioConfig)
{
    //TODO dupicated check
    int nAudioIndex;
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] == -1) {
            nAudioIndex = i;
        }
    }
    if(nAudioIndex == nMaxTracks) {
        return -1;
    }

    pj_status_t status;
    status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[nAudioIndex]);
    if(status != PJ_SUCCESS){
        ReleasePeerConnectoin(_pPeerConnection);
    }
    AddMediaTrack(&_pPeerConnection->mediaStream, _pAudioConfig, nAudioIndex, TYPE_AUDIO);
    return PJ_SUCCESS;
}

int AddVideoTrack(IN OUT PeerConnectoin * _pPeerConnection, IN MediaConfig * _pVideoConfig)
{
    //TODO dupicated check
    int nVideoIndex;
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] == -1) {
            nVideoIndex = i;
        }
    }
    if(nVideoIndex == nMaxTracks) {
        return -1;
    }
    
    pj_status_t status;
    status = initTransportIce(_pPeerConnection, &_pPeerConnection->transportIce[nVideoIndex]);
    if(status != PJ_SUCCESS){
        ReleasePeerConnectoin(_pPeerConnection);
    }
    AddMediaTrack(&_pPeerConnection->mediaStream, _pVideoConfig, nVideoIndex, TYPE_VIDEO);
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
        status = pjmedia_endpt_create_audio_sdp(_pMediaEndpt, _pPool, &tpinfo.sock_info, 0, _pSdp);
    } else {
        status = pjmedia_endpt_create_video_sdp(_pMediaEndpt, _pPool, &tpinfo.sock_info, 0, _pSdp);
    }
    
    return status;
}

static int createSdp(IN OUT PeerConnectoin * _pPeerConnection, IN pj_pool_t * _pPool, OUT pjmedia_sdp_session **_pOffer)
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

int createOffer(IN OUT PeerConnectoin * _pPeerConnection, IN pj_pool_t * _pPool, OUT pjmedia_sdp_session **_pOffer)
{
    pj_status_t  status;
    
    status = createSdp(_pPeerConnection, _pPool, _pOffer);
    STATUS_CHECK(createSdp, status);
    
    int nMaxTracks = sizeof(_pPeerConnection->nAvIndex) / sizeof(int);
    for ( int i = 0; i < nMaxTracks; i++) {
        if (_pPeerConnection->nAvIndex[i] != -1) {
            status = pjmedia_transport_media_create(_pPeerConnection->transportIce[i].pTransport, _pPool, 0, NULL, 0);
            STATUS_CHECK(pjmedia_transport_media_create, status);
            
            status = pjmedia_transport_encode_sdp(_pPeerConnection->transportIce[i].pTransport, _pPool, *_pOffer, NULL, i);
            STATUS_CHECK(pjmedia_transport_encode_sdp, status);
        }
    }
    
    return PJ_SUCCESS;
}

int createAnswer(IN OUT PeerConnectoin * _pPeerConnection, IN pj_pool_t * _pPool,
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

int setLocalDescription(IN OUT PeerConnectoin * _pPeerConnectoin, IN pjmedia_sdp_session * _pLocalSdp)
{
    _pPeerConnectoin->pOfferSdp = _pLocalSdp;
}

int setRemoteDescription(IN OUT PeerConnectoin * _pPeerConnectoin, IN pjmedia_sdp_session * _pRemoteSdp)
{
    _pPeerConnectoin->pAnswerSdp = _pRemoteSdp;
}
