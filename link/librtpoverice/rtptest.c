#include "PeerConnection.h"

#define SDP_NEG_TESG

typedef struct _App{
    PeerConnection peerConnection;
    pj_caching_pool cachingPool;
    MediaConfig audioConfig;
    MediaConfig videoConfig;
    IceConfig userConfig;
}App;
App app;

#define TESTCHECK(status, a) if(status != 0){\
ReleasePeerConnectoin(&a.peerConnection);\
return status;}

static void input_confirm(char * pmt)
{
    char input[10];
    while(1){
        printf("%s, confirm(ok):", pmt);
        memset(input, 0, sizeof(input));
        scanf("%s", input);
        if(strcmp("ok", input) == 0){
            break;
        }
    }
}

static void write_sdp(pjmedia_sdp_session * pSdp, char * pFname)
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

static void print_sdp(const pjmedia_sdp_session *pSdp) {
    char sdptxt[2048] = { 0 };
    pjmedia_sdp_print(pSdp, sdptxt, sizeof(sdptxt) - 1);
    printf("\n------------sdp-----------\n");
    printf("%s", sdptxt);
    printf("\n------------sdp-----------\n");
}

static void sdp_from_file(pjmedia_sdp_session ** sdp, char * pFname, pj_pool_t * pSdpPool, char *pBuf, int nBufLen)
{
#ifndef SDP_NEG_TESG
    input_confirm("input peer sdp file.(read from file)");
#endif
    
    //FILE * f = fopen("/Users/liuye/Documents/p2p/build/src/work/Debug/r.sdp", "rb");
    FILE * f = fopen(pFname, "rb");
    assert(f != NULL);
    
    fseek(f,0, SEEK_END);
    int flen = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    memset(pBuf, 0, nBufLen);
    int rlen = fread(pBuf, 1, flen, f);
    assert(rlen == flen);
    fclose(f);
    
    pj_status_t status;
    status = pjmedia_sdp_parse(pSdpPool, pBuf, rlen, sdp);
    assert(status == PJ_SUCCESS);
    
    print_sdp(*sdp);
}

#define OFFER 1
#define ANSWER 2
#define OFFERFILE "offer.sdp"
#define ANSWERFILE "answer.sdp"

#ifdef SDP_NEG_TESG
static void print_neg_state(pjmedia_sdp_neg * pNeg) {
    pjmedia_sdp_neg_state state = pjmedia_sdp_neg_get_state(pNeg);
    const char * str =  pjmedia_sdp_neg_state_str(state);
    switch (state) {
        case PJMEDIA_SDP_NEG_STATE_NULL:
            printf("PJMEDIA_SDP_NEG_STATE_NULL:%s\n", str);
            break;
        case PJMEDIA_SDP_NEG_STATE_LOCAL_OFFER:
            printf("PJMEDIA_SDP_NEG_STATE_LOCAL_OFFER:%s\n", str);
            break;
        case PJMEDIA_SDP_NEG_STATE_REMOTE_OFFER:
            printf("PJMEDIA_SDP_NEG_STATE_REMOTE_OFFER:%s\n", str);
            break;
        case PJMEDIA_SDP_NEG_STATE_WAIT_NEGO:
            printf("PJMEDIA_SDP_NEG_STATE_WAIT_NEGO:%s\n", str);
            break;
        case PJMEDIA_SDP_NEG_STATE_DONE:
            printf("PJMEDIA_SDP_NEG_STATE_DONE:%s\n", str);
            break;
    }
}
void pjmedia_sdp_neg_test(pj_pool_factory *_pFactory){
    char localTextSdpBuf[2048] = {0};
    pj_pool_t *pOfferPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
    pj_assert(pOfferPool);
    pjmedia_sdp_session *pOffer = NULL;
    sdp_from_file(&pOffer, OFFERFILE, pOfferPool, localTextSdpBuf, sizeof(localTextSdpBuf));
    pj_assert(pOffer != NULL);
    
    char remoteTextSdpBuf[2048] = {0};
    pj_pool_t *pAnswerPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
    pj_assert(pAnswerPool);
    pjmedia_sdp_session *pAnswer = NULL;
    sdp_from_file(&pAnswer, ANSWERFILE, pAnswerPool, remoteTextSdpBuf, sizeof(remoteTextSdpBuf));
    pj_assert(pAnswer != NULL);
    
    pj_status_t status;
    pj_pool_t *pSdpNegPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
    pj_assert(pSdpNegPool);
    
    pjmedia_sdp_neg *pIceNeg = NULL;
    status = pjmedia_sdp_neg_create_w_local_offer (pSdpNegPool, pOffer, &pIceNeg);
    pj_assert(status == PJ_SUCCESS);
    pj_assert(pIceNeg);
    pjmedia_sdp_neg_set_prefer_remote_codec_order(pIceNeg, 0);
    print_neg_state(pIceNeg);
    
    status = pjmedia_sdp_neg_set_remote_answer (pSdpNegPool, pIceNeg, pAnswer);
    pj_assert(status == PJ_SUCCESS);
    print_neg_state(pIceNeg);
    
    pj_pool_t *pNegPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
    pj_assert(pNegPool);
    status = pjmedia_sdp_neg_negotiate (pNegPool, pIceNeg, 0);
    pj_assert(status == PJ_SUCCESS);
    print_neg_state(pIceNeg);
    
    const pjmedia_sdp_session * pLocalActiveSdp = NULL;
    status = pjmedia_sdp_neg_get_active_local(pIceNeg, &pLocalActiveSdp);
    pj_assert(status == PJ_SUCCESS);
    printf("local active sdp:\n");
    print_sdp(pLocalActiveSdp);
    

    const pjmedia_sdp_session * pRemoteActiveSdp = NULL;
    status = pjmedia_sdp_neg_get_active_remote(pIceNeg, &pRemoteActiveSdp);
    pj_assert(status == PJ_SUCCESS);
    printf("remote active sdp:\n");
    print_sdp(pRemoteActiveSdp);

}
#endif //SDP_NEG_TEST

int main(int argc, char **argv)
{
    if(argc != 2){
        printf("usage as:%s [1 for offer|2 for answer]\n", argv[0]);
        return -1;
    }
    int role = atoi(argv[1]);
    if(role != OFFER && role != ANSWER){
        printf("usage as:%s [1 for offer|2 for answer]\n", argv[0]);
        return -1;
    }
    
    pj_status_t status;
    status = pj_init();
    PJ_ASSERT_RETURN(status == PJ_SUCCESS, 1);
    status = pjlib_util_init();
    PJ_ASSERT_RETURN(status == PJ_SUCCESS, 1);
    
    pj_pool_t * pRemoteSdpPool = NULL;
    char textSdpBuf[2048] = {0};
    
    pj_caching_pool_init(&app.cachingPool, &pj_pool_factory_default_policy, 0);
    
    
    //InitIceConfig(&app.userConfig);
    //strcpy(app.userConfig.turnHost, "127.0.0.1");
    //strcpy(app.userConfig.turnHost, "123.59.204.198");
    //strcpy(app.userConfig.turnUsername, "root");
    //strcpy(app.userConfig.turnPassword, "root");
    //there is default ice config
    InitPeerConnectoin(&app.peerConnection, &app.cachingPool.factory, NULL); //&app.userConfig
    
#ifdef SDP_NEG_TESG
    printf("go into pjmedia_sdp_neg_test\n");
    pjmedia_sdp_neg_test(&app.cachingPool.factory);
    return 0;
#endif //SDP_NEG_TEST
    
    InitMediaConfig(&app.audioConfig);
    app.audioConfig.configs[0].nSampleOrClockRate = 8000;
    app.audioConfig.configs[0].nRtpDynamicType = 0;
    app.audioConfig.configs[0].format = MEDIA_FORMAT_PCMU;
    app.audioConfig.configs[1].nSampleOrClockRate = 8000;
    app.audioConfig.configs[1].nRtpDynamicType = 8;
    app.audioConfig.configs[1].format = MEDIA_FORMAT_PCMA;
    app.audioConfig.nCount = 2;
    if ( role == ANSWER ){
        app.audioConfig.configs[0] = app.audioConfig.configs[1];
        app.audioConfig.configs[1].nSampleOrClockRate = 8000;
        app.audioConfig.configs[1].nRtpDynamicType = 18;
        app.audioConfig.configs[1].format = MEDIA_FORMAT_G729;
    }
    status = AddAudioTrack(&app.peerConnection, &app.audioConfig);
    TESTCHECK(status, app);
    
    InitMediaConfig(&app.videoConfig);
    app.videoConfig.configs[0].nSampleOrClockRate = 90000;
    app.videoConfig.configs[0].nRtpDynamicType = 98;
    app.videoConfig.configs[0].format = MEDIA_FORMAT_H264;
    app.videoConfig.configs[1].nSampleOrClockRate = 90000;
    app.videoConfig.configs[1].nRtpDynamicType = 99;
    app.videoConfig.configs[1].format = MEDIA_FORMAT_H265;
    app.videoConfig.nCount = 2;
    if ( role == ANSWER ){
        app.videoConfig.nCount = 1;
        app.videoConfig.configs[0] = app.videoConfig.configs[1];
    }
    status = AddVideoTrack(&app.peerConnection, &app.videoConfig);
    TESTCHECK(status, app);
    
    pj_pool_t * pSdpPool = pj_pool_create(&app.cachingPool.factory, NULL, 1024, 512, NULL);
    ASSERT_RETURN_CHECK(pSdpPool, pj_pool_create);
    if (role == OFFER) {
        pjmedia_sdp_session *pOffer = NULL;
        status = createOffer(&app.peerConnection, pSdpPool, &pOffer);
        TESTCHECK(status, app);
        setLocalDescription(&app.peerConnection, pOffer);
        write_sdp(pOffer, OFFERFILE);
        
        pjmedia_sdp_session *pAnswer = NULL;
        pRemoteSdpPool =pj_pool_create(&app.cachingPool.factory, "sdpremote", 2048, 1024, NULL);
        sdp_from_file(&pAnswer, ANSWERFILE,  pRemoteSdpPool, textSdpBuf, sizeof(textSdpBuf));
        setRemoteDescription(&app.peerConnection, pAnswer);
    }
    
    if (role == ANSWER) {
        pjmedia_sdp_session *pOffer = NULL;
        pRemoteSdpPool =pj_pool_create(&app.cachingPool.factory, "sdpremote", 2048, 1024, NULL);
        sdp_from_file(&pOffer, OFFERFILE,  pRemoteSdpPool, textSdpBuf, sizeof(textSdpBuf));
        setRemoteDescription(&app.peerConnection, pOffer);
        
        pjmedia_sdp_session *pAnswer = NULL;
        status = createAnswer(&app.peerConnection, pSdpPool, pOffer, &pAnswer);
        TESTCHECK(status, app);
        setLocalDescription(&app.peerConnection, pAnswer);
        write_sdp(pAnswer, ANSWERFILE);
    }
    
    input_confirm("confirm to negotiation:");
    StartNegotiation(&app.peerConnection);
    
    char packet[120];
    while(1){
        memset(packet, 0, sizeof(packet));
        memset(packet, 0x30, 12);
        printf("input:");
        scanf("%s", packet+12);
        if(packet[12] == 'q'){
            break;
        }
        pjmedia_transport_send_rtp(app.peerConnection.transportIce[0].pTransport, packet, strlen(packet));
        memset(packet, 0x31, 12);
        pjmedia_transport_send_rtp(app.peerConnection.transportIce[1].pTransport, packet, strlen(packet));
    }
    
    input_confirm("quit");
    
    return 0;
}
