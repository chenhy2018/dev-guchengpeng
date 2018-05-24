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


int main(int argc, char **argv)
{
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
    pjmedia_sdp_session *pOffer = NULL;
    status = createOffer(&app.peerConnection, pSdpPool, &pOffer);
    TESTCHECK(status, app);
    setLocalDescription(&app.peerConnection, pOffer);
    
    return 0;
}
