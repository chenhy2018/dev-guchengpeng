#include "PeerConnection.h"

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

char gRemoteSdpBuf[2048] = {0};
pj_pool_t * pRemoteSdpPool = NULL;
static void sdp_from_file(pjmedia_sdp_session ** sdp, char * pFname, pj_pool_t * pRemoteSdpPool)
{
    input_confirm("input peer sdp file.(read from file)");
    
    //FILE * f = fopen("/Users/liuye/Documents/p2p/build/src/work/Debug/r.sdp", "rb");
    FILE * f = fopen(pFname, "rb");
    assert(f != NULL);
    
    fseek(f,0, SEEK_END);
    int flen = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    memset(gRemoteSdpBuf, 0, sizeof(gRemoteSdpBuf));
    int rlen = fread(gRemoteSdpBuf, 1, flen, f);
    assert(rlen == flen);
    fclose(f);
    
    pj_status_t status;
    status = pjmedia_sdp_parse(pRemoteSdpPool, gRemoteSdpBuf, strlen(gRemoteSdpBuf), sdp);
    assert(status == PJ_SUCCESS);
    
    char sdptxt[2048] = {0};
    pjmedia_sdp_print(*sdp, sdptxt, sizeof(sdptxt) - 1);
    printf("\n------------sdp from file start-----------\n");
    printf("%s", sdptxt);
    printf("\n------------sdp from file   end-----------\n");
}

#define OFFER 1
#define ANSWER 2
#define OFFERFILE "offer.sdp"
#define ANSWERFILE "answer.sdp"
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
    
    pj_caching_pool_init(&app.cachingPool, &pj_pool_factory_default_policy, 0);
    
    
    InitIceConfig(&app.userConfig);
    strcpy(app.userConfig.turnHost, "123.59.204.198");
    strcpy(app.userConfig.turnUsername, "root");
    strcpy(app.userConfig.turnPassword, "root");
    InitPeerConnectoin(&app.peerConnection, &app.userConfig, &app.cachingPool.factory);
    
    InitMediaConfig(&app.audioConfig);
    app.audioConfig.audioConfig.nSampleRate = 8000;
    app.audioConfig.audioConfig.nRtpDynamicType = 96;
    app.audioConfig.audioConfig.format = MEDIA_FORMAT_PCMU;
    status = AddAudioTrack(&app.peerConnection, &app.audioConfig);
    TESTCHECK(status, app);
    
    InitMediaConfig(&app.videoConfig);
    app.videoConfig.videoConfig.nClockRate = 90000;
    app.videoConfig.videoConfig.nRtpDynamicType = 98;
    app.videoConfig.videoConfig.format = MEDIA_FORMAT_H264;
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
        sdp_from_file(&pAnswer, ANSWERFILE,  pRemoteSdpPool);
        setRemoteDescription(&app.peerConnection, pAnswer);
    }
    
    if (role == ANSWER) {
        pjmedia_sdp_session *pOffer = NULL;
        pRemoteSdpPool =pj_pool_create(&app.cachingPool.factory, "sdpremote", 2048, 1024, NULL);
        sdp_from_file(&pOffer, OFFERFILE,  pRemoteSdpPool);
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
