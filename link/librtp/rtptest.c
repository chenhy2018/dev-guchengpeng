#include <pjsip.h>
#include <pjmedia.h>
#include <pjmedia-codec.h>
#include <pjlib-util.h>
#include <pjlib.h>
#include "qrtc.h"
#define THIS_FILE "rtptest.c"

//#define SDP_NEG_TESG

typedef struct _App{
        PeerConnection *pPeerConnection;
        pj_caching_pool cachingPool;
        MediaConfigSet audioConfig;
        MediaConfigSet videoConfig;
        IceConfig userConfig;
}App;
App app;
pj_oshandle_t gLogFd;

#define TESTCHECK(status, a) if(status != 0){\
ReleasePeerConnectoin(a.pPeerConnection);\
return status;}


//read audio video from file start
#include <sys/time.h>
#include <unistd.h>
#define THIS_IS_AUDIO 1
#define THIS_IS_VIDEO 2
typedef int (*DataCallback)(void *pData, int nDataLen, int nFlag, int64_t timestamp);
static const uint8_t *ff_avc_find_startcode_internal(const uint8_t *p, const uint8_t *end)
{
        const uint8_t *a = p + 4 - ((intptr_t)p & 3);
        
        for (end -= 3; p < a && p < end; p++) {
                if (p[0] == 0 && p[1] == 0 && p[2] == 1)
                        return p;
        }
        
        for (end -= 3; p < end; p += 4) {
                uint32_t x = *(const uint32_t*)p;
                //      if ((x - 0x01000100) & (~x) & 0x80008000) // little endian
                //      if ((x - 0x00010001) & (~x) & 0x00800080) // big endian
                if ((x - 0x01010101) & (~x) & 0x80808080) { // generic
                        if (p[1] == 0) {
                                if (p[0] == 0 && p[2] == 1)
                                        return p;
                                if (p[2] == 0 && p[3] == 1)
                                        return p+1;
                        }
                        if (p[3] == 0) {
                                if (p[2] == 0 && p[4] == 1)
                                        return p+2;
                                if (p[4] == 0 && p[5] == 1)
                                        return p+3;
                        }
                }
        }
        
        for (end += 3; p < end; p++) {
                if (p[0] == 0 && p[1] == 0 && p[2] == 1)
                        return p;
        }
        
        return end + 3;
}

const uint8_t *ff_avc_find_startcode(const uint8_t *p, const uint8_t *end){
        const uint8_t *out= ff_avc_find_startcode_internal(p, end);
        if(p<out && out<end && !out[-1]) out--;
        return out;
}

static int getFileAndLength(char *_pFname, FILE **_pFile, int *_pLen)
{
        FILE * f = fopen(_pFname, "r");
        if ( f == NULL ) {
                return -1;
        }
        *_pFile = f;
        fseek(f, 0, SEEK_END);
        long nLen = ftell(f);
        fseek(f, 0, SEEK_SET);
        *_pLen = (int)nLen;
        return 0;
}

static int readFileToBuf(char * _pFilename, char ** _pBuf, int *_pLen)
{
        int ret;
        FILE * pFile;
        int nLen = 0;
        ret = getFileAndLength(_pFilename, &pFile, &nLen);
        if (ret != 0) {
                fprintf(stderr, "open file %s fail\n", _pFilename);
                return -1;
        }
        char *pData = malloc(nLen);
        assert(pData != NULL);
        ret = fread(pData, 1, nLen, pFile);
        if (ret <= 0) {
                fprintf(stderr, "open file %s fail\n", _pFilename);
                fclose(pFile);
                free(pData);
                return -2;
        }
        *_pBuf = pData;
        *_pLen = nLen;
        return 0;
}

static inline int64_t getCurrentMilliSecond(){
        struct timeval tv;
        gettimeofday(&tv, NULL);
        return (tv.tv_sec*1000 + tv.tv_usec/1000);
}

int start_file_test(char * _pAudioFile, char * _pVideoFile, DataCallback callback)
{
        assert(!(_pAudioFile == NULL && _pVideoFile == NULL));

        int ret;

        char * pAudioData = NULL;
        int nAudioDataLen = 0;
        if(_pAudioFile != NULL){
                ret = readFileToBuf(_pAudioFile, &pAudioData, &nAudioDataLen);
                if (ret != 0) {
                        MY_PJ_LOG(3, "map data to buffer fail:%s", _pAudioFile);
                        return -1;
                }
        }

        char * pVideoData = NULL;
        int nVideoDataLen = 0;
        if(_pVideoFile != NULL){
                ret = readFileToBuf(_pVideoFile, &pVideoData, &nVideoDataLen);
                if (ret != 0) {
                        free(pAudioData);
                        MY_PJ_LOG(3, "map data to buffer fail:%s", _pVideoFile);
                        return -2;
                }
        }

        int bAudioOk = 1;
        int bVideoOk = 1;
        if (_pVideoFile == NULL) {
                bVideoOk = 0;
        }
        if (_pAudioFile == NULL) {
                bAudioOk = 0;
        }
        int64_t nSysTimeBase = getCurrentMilliSecond();
        int64_t nNextAudioTime = nSysTimeBase;
        int64_t nNextVideoTime = nSysTimeBase;
        int64_t nNow = nSysTimeBase;
        int audioOffset = 0;

        uint8_t * nextstart = (uint8_t *)pVideoData;
        uint8_t * endptr = nextstart + nVideoDataLen;
        int cbRet = 0;
        int nIDR = 0;
        int nNonIDR = 0;
        while (bAudioOk || bVideoOk) {
                if (bAudioOk && nNow+1 > nNextAudioTime) {
                        if(audioOffset+160 <= nAudioDataLen) {
                                cbRet = callback(pAudioData + audioOffset, 160, THIS_IS_AUDIO, nNextAudioTime-nSysTimeBase);
                                if (cbRet != 0) {
                                        bAudioOk = 0;
                                        continue;
                                }
                                audioOffset += 160;
                                nNextAudioTime += 20;
                        } else {
                                bAudioOk = 0;
                        }
                }
                if (bVideoOk && nNow+1 > nNextVideoTime) {

                        uint8_t * start = NULL;
                        uint8_t * end = NULL;
                        uint8_t * sendp = NULL;
                        int eof = 0;
                        int type = -1;
                        do{
                                start = (uint8_t *)ff_avc_find_startcode((const uint8_t *)nextstart, (const uint8_t *)endptr);
                                end = (uint8_t *)ff_avc_find_startcode(start+4, endptr);

                                nextstart = end;
                                if(sendp == NULL)
                                        sendp = start;

                                if(start == end || end > endptr){
                                        eof = 1;
                                        bVideoOk = 0;
                                        break;
                                }

                                if(start[2] == 0x01){//0x 00 00 01
                                        type = start[3] & 0x1F;
                                }else{ // 0x 00 00 00 01
                                        type = start[4] & 0x1F;
                                }

                                if(type == 1 || type == 5 ){
                                        if (type == 1) {
                                                nNonIDR++;
                                        } else {
                                                nIDR++;
                                        }
                                        MY_PJ_LOG(3, "send one video(%d) frame packet:%ld", type, end - sendp);
                                        cbRet = callback(sendp, end - sendp, THIS_IS_VIDEO, nNextVideoTime-nSysTimeBase);
                                        if (cbRet != 0) {
                                                bVideoOk = 0;
                                        }
                                        nNextVideoTime += 40;
                                        break;
                                }
                        }while(1);
                }

                int64_t nSleepTime = 0;
                if (nNextAudioTime > nNextVideoTime) {
                        if (nNextVideoTime - nNow >  1)
                                nSleepTime = (nNextVideoTime - nNow - 1) * 1000;
                } else {
                        if (nNextAudioTime - nNow > 1)
                                nSleepTime = (nNextAudioTime - nNow - 1) * 1000;
                }
                if (nSleepTime != 0) {
                        MY_PJ_LOG(3, "sleeptime:%ld\n", nSleepTime);
                        usleep(nSleepTime);
                }
                nNow = getCurrentMilliSecond();
        }

        if (pAudioData) {
                free(pAudioData);
        }
        if (pVideoData) {
                free(pVideoData);
                printf("IDR:%d nonIDR:%d\n", nIDR, nNonIDR);
        }
        return 0;
}
//read audio video from file end

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
        pj_file_flush(gLogFd);
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

const char *sdpNegOffer="v=0\n"
"o=- 3736653588 3736653588 IN IP4 127.0.0.1\n"
"s=pjmedia\n"
"t=0 0\n"
"m=audio 49477 RTP/AVP 0 8\n"
"c=IN IP4 172.20.4.69\n"
"a=rtcp:49478 IN IP4 172.20.4.69\n"
"a=sendrecv\n"
"a=ice-ufrag:088e4954\n"
"a=ice-pwd:35702e2f\n"
"a=candidate:Rac140445 1 UDP 16777215 172.20.4.69 49477 typ relay raddr 222.73.202.226 rport 20584\n"
"a=candidate:Rac140445 2 UDP 16777214 172.20.4.69 49478 typ relay raddr 222.73.202.226 rport 20585\n"
"m=video 49479 RTP/AVP 98 99\n"
"c=IN IP4 172.20.4.69\n"
"a=rtcp:49480 IN IP4 172.20.4.69\n"
"a=sendrecv\n"
"a=rtpmap:98 H264/90000\n"
"a=rtpmap:99 H265/90000\n"
"a=ice-ufrag:77188b05\n"
"a=ice-pwd:6c4f3258\n"
"a=candidate:Rac140445 1 UDP 16777215 172.20.4.69 49479 typ relay raddr 222.73.202.226 rport 20586\n"
"a=candidate:Rac140445 2 UDP 16777214 172.20.4.69 49480 typ relay raddr 222.73.202.226 rport 20587\n";

const char *sdpNegAnswer="v=0\n"
"o=- 3736653593 3736653593 IN IP4 127.0.0.1\n"
"s=pjmedia\n"
"t=0 0\n"
"m=audio 49481 RTP/AVP 8 18\n"
"c=IN IP4 172.20.4.69\n"
"a=rtcp:49484 IN IP4 172.20.4.69\n"
"a=sendrecv\n"
"a=ice-ufrag:088e4954\n"
"a=ice-pwd:35702e2f\n"
"a=candidate:Rac140445 1 UDP 16777215 172.20.4.69 49481 typ relay raddr 222.73.202.226 rport 20588\n"
"a=candidate:Rac140445 2 UDP 16777214 172.20.4.69 49484 typ relay raddr 222.73.202.226 rport 20590\n"
"m=video 49483 RTP/AVP 99\n"
"c=IN IP4 172.20.4.69\n"
"a=rtcp:49482 IN IP4 172.20.4.69\n"
"a=sendrecv\n"
"a=rtpmap:99 H265/90000\n"
"a=ice-ufrag:77188b05\n"
"a=ice-pwd:6c4f3258\n"
"a=candidate:Rac140445 1 UDP 16777215 172.20.4.69 49483 typ relay raddr 222.73.202.226 rport 20589\n"
"a=candidate:Rac140445 2 UDP 16777214 172.20.4.69 49482 typ relay raddr 222.73.202.226 rport 20591\n";

static void sdp_from_mem(pjmedia_sdp_session ** sdp, pj_pool_t * pSdpPool, char *pSdpStr, const char * pSrcSdpStr)
{
        pj_memcpy(pSdpStr, pSrcSdpStr, strlen(pSrcSdpStr));
        pj_status_t status;
        status = pjmedia_sdp_parse(pSdpPool, pSdpStr, strlen(pSdpStr), sdp);
        assert(status == PJ_SUCCESS);
        
        print_sdp(*sdp);
}

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

void pjmedia_sdp_neg_test_as_offer(pj_pool_factory *_pFactory)
{
        char localTextSdpBuf[2048] = {0};
        pj_pool_t *pOfferPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pOfferPool);
        pjmedia_sdp_session *pOffer = NULL;
        sdp_from_mem(&pOffer, pOfferPool, localTextSdpBuf, sdpNegOffer);
        pj_assert(pOffer != NULL);
        
        char remoteTextSdpBuf[2048] = {0};
        pj_pool_t *pAnswerPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pAnswerPool);
        pjmedia_sdp_session *pAnswer = NULL;
        sdp_from_mem(&pAnswer, pAnswerPool, remoteTextSdpBuf, sdpNegAnswer);
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
        
        pj_str_t expectFmt;
        
        pj_assert(pLocalActiveSdp->media[0]->desc.fmt_count == 1);
        expectFmt = pj_str("8");
        pj_assert(pj_strcmp(&pLocalActiveSdp->media[0]->desc.fmt[0], &expectFmt) == 0);

        pj_assert(pLocalActiveSdp->media[1]->desc.fmt_count == 1);
        expectFmt = pj_str("99");
        pj_assert(pj_strcmp(&pLocalActiveSdp->media[1]->desc.fmt[0], &expectFmt) == 0);

        
        
        const pjmedia_sdp_session * pRemoteActiveSdp = NULL;
        status = pjmedia_sdp_neg_get_active_remote(pIceNeg, &pRemoteActiveSdp);
        pj_assert(status == PJ_SUCCESS);
        printf("remote active sdp:\n");
        print_sdp(pRemoteActiveSdp);
        
        pj_assert(pRemoteActiveSdp->media[0]->desc.fmt_count == 1);
        expectFmt = pj_str("8");
        pj_assert(pj_strcmp(&pRemoteActiveSdp->media[0]->desc.fmt[0], &expectFmt) == 0);

        pj_assert(pRemoteActiveSdp->media[1]->desc.fmt_count == 1);
        expectFmt = pj_str("99");
        pj_assert(pj_strcmp(&pRemoteActiveSdp->media[1]->desc.fmt[0], &expectFmt) == 0);

        
        pj_pool_release(pOfferPool);
        pj_pool_release(pAnswerPool);
        pj_pool_release(pSdpNegPool);
        pj_pool_release(pNegPool);
}

void pjmedia_sdp_neg_test_as_answer(pj_pool_factory *_pFactory)
{
        char localTextSdpBuf[2048] = {0};
        pj_pool_t *pOfferPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pOfferPool);
        pjmedia_sdp_session *pOffer = NULL;
        sdp_from_mem(&pOffer, pOfferPool, localTextSdpBuf, sdpNegOffer);
        pj_assert(pOffer != NULL);
        
        char remoteTextSdpBuf[2048] = {0};
        pj_pool_t *pAnswerPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pAnswerPool);
        pjmedia_sdp_session *pAnswer = NULL;
        sdp_from_mem(&pAnswer, pAnswerPool, remoteTextSdpBuf, sdpNegAnswer);
        pj_assert(pAnswer != NULL);
        
        pj_status_t status;
        pj_pool_t *pSdpNegPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pSdpNegPool);
        
#if 0
        pjmedia_sdp_neg *pIceNeg = NULL;
        status = pjmedia_sdp_neg_create_w_remote_offer (pSdpNegPool, pAnswer, pOffer, &pIceNeg);
        pj_assert(status == PJ_SUCCESS);
        pj_assert(pIceNeg);
#else
        pjmedia_sdp_neg *pIceNeg = NULL;
        status = pjmedia_sdp_neg_create_w_remote_offer (pSdpNegPool, NULL, pOffer, &pIceNeg);
        pj_assert(status == PJ_SUCCESS);
        pj_assert(pIceNeg);
        print_neg_state(pIceNeg);
        status = pjmedia_sdp_neg_set_local_answer(pSdpNegPool, pIceNeg, pAnswer);
        pj_assert(status == PJ_SUCCESS);
#endif
        pjmedia_sdp_neg_set_prefer_remote_codec_order(pIceNeg, 0);
        print_neg_state(pIceNeg);
        
        pj_pool_t *pNegPool = pj_pool_create(_pFactory, NULL, 512, 512, NULL);
        pj_assert(pNegPool);
        status = pjmedia_sdp_neg_negotiate (pNegPool, pIceNeg, 0);
        pj_assert(status == PJ_SUCCESS);
        print_neg_state(pIceNeg);
        
        const pjmedia_sdp_session * pLocalActiveSdp = NULL;
        status = pjmedia_sdp_neg_get_active_local(pIceNeg, &pLocalActiveSdp);
        pj_assert(status == PJ_SUCCESS);
        printf("local active sdp2:\n");
        print_sdp(pLocalActiveSdp);
        
        pj_str_t expectFmt;
        
        pj_assert(pLocalActiveSdp->media[0]->desc.fmt_count == 1);
        expectFmt = pj_str("8");
        pj_assert(pj_strcmp(&pLocalActiveSdp->media[0]->desc.fmt[0], &expectFmt) == 0);

        pj_assert(pLocalActiveSdp->media[1]->desc.fmt_count == 1);
        expectFmt = pj_str("99");
        pj_assert(pj_strcmp(&pLocalActiveSdp->media[1]->desc.fmt[0], &expectFmt) == 0);

        
        const pjmedia_sdp_session * pRemoteActiveSdp = NULL;
        status = pjmedia_sdp_neg_get_active_remote(pIceNeg, &pRemoteActiveSdp);
        pj_assert(status == PJ_SUCCESS);
        printf("remote active sdp2:\n");
        print_sdp(pRemoteActiveSdp);
        
#if 0
        //TODO pjmedia_sdp_neg_get_active_remote目前还是和预想的不一样
        pj_assert(pRemoteActiveSdp->media[0]->desc.fmt_count == 1);
        expectFmt = pj_str("8");
        pj_assert(pj_strcmp(&pRemoteActiveSdp->media[0]->desc.fmt[0], &expectFmt) == 0);

        pj_assert(pRemoteActiveSdp->media[1]->desc.fmt_count == 1);
        expectFmt = pj_str("100");
        pj_assert(pj_strcmp(&pRemoteActiveSdp->media[1]->desc.fmt[0], &expectFmt) == 0);
#endif
        
        pj_pool_release(pOfferPool);
        pj_pool_release(pAnswerPool);
        pj_pool_release(pSdpNegPool);
        pj_pool_release(pNegPool);
}

IceState gatherState = ICE_STATE_INIT;

static int waitState(IN IceState currentState)
{
        int nCnt = 0;
        do{
                if (nCnt > 5) {
                        return 1; //fail
                }
                nCnt++;
                pj_thread_sleep(500);
        }while(gatherState == currentState);
        
        return 0;
}

pj_oshandle_t gPcmuFd;
pj_oshandle_t gH264Fd;

static void onRxRtp(void *_pUserData, CallbackType _type, void *_pCbData)
{
        switch (_type){
                case CALLBACK_ICE:{
                        IceNegInfo *pInfo = (IceNegInfo *)_pCbData;
                        if (pInfo->state == ICE_STATE_GATHERING_OK) {
                                gatherState = ICE_STATE_GATHERING_OK;
                                break;
                        }
                        if (pInfo->state == ICE_STATE_GATHERING_FAIL) {
                                gatherState = ICE_STATE_GATHERING_FAIL;
                                break;
                        }
                        if (pInfo->state == ICE_STATE_NEGOTIATION_OK) {
                                MY_PJ_LOG(3, "==========>callback_ice: state:%d", pInfo->state);
                                for ( int i = 0; i < pInfo->nCount; i++) {
                                        MY_PJ_LOG(3, "           codec type:%d", pInfo->configs[i]->codecType);
                                }
                                break;
                        }
                }
                        break;
                case CALLBACK_RTP:{
                        RtpPacket *pPkt = (RtpPacket *)_pCbData;
                        pj_ssize_t nLen = pPkt->nDataLen;
                        if (pPkt->type == RTP_STREAM_AUDIO && nLen == 160) {
                                pj_file_write(gPcmuFd, pPkt->pData, &nLen);
                        } else if (pPkt->type == RTP_STREAM_VIDEO) {
                                pj_file_write(gH264Fd, pPkt->pData, &nLen);
                        }
                }
                        break;
                case CALLBACK_RTCP:
                        fprintf(stderr, "==========>callback_rtcp\n");
                        break;
        }
}

static int receive_data_callback(void *pData, int nDataLen, int nFlag, int64_t timestamp)
{
        RtpPacket rtpPacket;
        pj_bzero(&rtpPacket, sizeof(rtpPacket));
        if (nFlag == THIS_IS_AUDIO) {
                //printf("send %d bytes audio data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                rtpPacket.type = RTP_STREAM_AUDIO;
        } else {
                //printf("send %d bytes vidoe data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                rtpPacket.type = RTP_STREAM_VIDEO;
        }
        rtpPacket.pData = (uint8_t *)pData;
        rtpPacket.nDataLen = nDataLen;
        rtpPacket.nTimestamp = timestamp;
        return SendRtpPacket(app.pPeerConnection, &rtpPacket);
}

char * pLogFileName = NULL;

void log_to_file(int _nLevel, const char *_pData, int _nLen)
{
        pj_ssize_t nLen = _nLen;
        pj_file_write(gLogFd, _pData, &nLen);
}

void set_log_to_file(int role)
{
        pj_status_t status;
        if (role == ANSWER) {
                pLogFileName = "answer.log";
        } else {
                pLogFileName = "offer.log";
        }
        pj_pool_t * logpool = pj_pool_create(&app.cachingPool.factory, "log", 2000, 2000, NULL);
        char logfileName[128] = {0};
        sprintf(logfileName, "/Users/liuye/Documents/p2p/build/src/work/Debug/%s", pLogFileName);
        status = pj_file_open(logpool, logfileName, PJ_O_WRONLY, &gLogFd);
        if(status != PJ_SUCCESS){
                char errmsg[PJ_ERR_MSG_SIZE] = {0};
                pj_strerror(status, errmsg, PJ_ERR_MSG_SIZE);
                printf("pj_file_open %s fail:%s\n", logfileName, errmsg);
                assert(status == PJ_SUCCESS);
        }
        pj_log_set_log_func(log_to_file);
}

void test_sdp_neg()
{
        //start pjmedia_sdp_neg test
        printf("--------pjmedia_sdp_neg_test1----------\n");
        pjmedia_sdp_neg_test_as_offer(&app.cachingPool.factory);
        printf("--------pjmedia_sdp_neg_test2----------\n");
        pjmedia_sdp_neg_test_as_answer(&app.cachingPool.factory);
        printf("%s\n", sdpNegOffer);
        printf("%s\n", sdpNegAnswer);
        printf("--------end pjmedia_sdp_neg_test----------\n");
        //end pjmedia_sdp_neg test
}

#define HAS_AUDIO 0x01
#define HAS_VIDEO 0x02

//jitter buffer
#include "JitterBuffer.h"
void test_jitter_buffer()
{
        pj_pool_t * pool = pj_pool_create(&app.cachingPool.factory, "jbuf", 256 * 50, 256 * 10, NULL);
        JitterBuffer jb;
        
        JitterBufferInit(&jb, 50, 20, pool, 256);
        
        int seqs[] = {
#if 1
                65499, 65501, 65500, 65503, 65502, 65504, 65505, 65506, 65507, 65509,
                65508, 65510, 65520, 65513, 65514, 65516, 65515, 65518, 65517, 65519,
                65511, 65512, 65521, 65522, 65523, 65525, 65526, 65527, 65529, 65528,
                65530, 65532, 65533, 65534, 65535, 0,     2,     3,     4,     1,
                6,     7,     8,     9,     10,    13,    12,    11,    15,    14,
                16,    17,    18,    19,    30,    20,    22,    21,    25,    24,
                23,    22,    26,    27,    28,    29,    31,    33,    32,    34,
                35,    36,    37,    38,    39,    40,    41,    42,    43,    50,
                49,    44,    45,    46,    47,    48,    51,    52,    53,    54,
                55,    56,    57,    58,    59,    60,    61,    62,    63,    64,
                65,    66,    67,    68,    69,    70,    71,    72,    73,    74,
                75,    76,    83,    78,    79,    80,    81,    82,    77,    84,
                85,    86,    87,    89,    88,    90,    91,    92,    93,    94,
                95,    96,    97,    98,    99,    100,   101,   102,   103,   104,
                105,   106,   107,   108,   109,   111,   110,   112,   113,   114
#else
                8001,8002,8005,8003,8004,8007,8006,8009,8008,8011,
                8010,8012,8013,8014,8015,8016,8017,8019,8020,8021,
                8022,8023,8024,8025,8026,8027,8028,8029,8030,8031,
                8032,8033,8035,8034,8036,8037,8038,8039,8040,8042,
                8041,8043,8044,8045,8046,8047,8049,8048,8050,8051,
                8052,8053,8054,8055,8056,8057,8058,8059,8060,8062,
                8061,8064,8066,8063,8065,8067,8068,8069,8070,8071,
                8072,8073,8074,8076,8075,8077,8079,8078,8080,8081,
                8082,8083,8084,8085,8086,8088,8087,8089,8090,8091,
                8092,8093,8094,8095,8096,8097,8098,8099,8100,8101,
                8102,8103,8104,8105,8106,8107,8109,8108,8110,8111,
                8112,8113,8114,8116,8115,8117,8118,8119,8120,8121,
                8123,8122,8124,8125,8127,8126,8128,8130,8129,8131,
                8132,8133,8134,8135,8136,8137,8138,8139,8140,8142,
                8141,8143,8144,8145,8146,8147,8148,8149,8150,8151,
                8152,8153,8154,8155,8156,8157,8158,8159,8160,8161,
                8162,8163,8164,8165,8166,8168,8167,8169,8170,8171,
                8172,8173,8175,8174,8176,8177,8178,8179,8180,8182,
                8181,8183,8184,8185,8186,8187,8189,8188,8190,8191,
                8192,8193,8194,8196,8195,8197,8198,8199,8200,8201,
                8203,8202,8204,8209,8206,8205,8207,8208,8211,8210,
                8212,8214,8213,8215,8216,8217,8218,8219,8221,8222,
                8223,8224,8225,8226,8227,8228,8229,8230,8231,8232,
                8233,8235,8234,8236,8237,8238,8239,8240,8242,8241,
                8243,8244,8245,8247,8249,8248,8250,8251,8252,8253,
                8254,8256,8255,8257,8258,8259,8260,8261,8262,8263,
                8264,8265,8266,8267,8268,8269,8270,8271,8273,8272,
                8275,8274,8276,8277,8279,8278,8280,8282,8281,8283,
                8284,8285,8286,8287,8288,8289,8290,8291,8292,8293,
                8294,8295,8296,8298,8297,8299,8300,8301,8302,8303,
                8304,8305,8307,8306,8308,8309,8310,8311,8312,8314,
                8313,8315,8316,8317,8318,8319,8320,8321,8322,8323,
                8324,8325,8326,8328,8327,8329,8331,8330,8332,8333,
                8334,8335,8336,8337,8338,8339,8341,8340,8342,8343,
                8344,8345,8346,8347,8348,8349,8350,8351,8353,8352,
                8354,8355,8356,8357,8358,8359,8360,8361,8362,8363,
                8364,8365,8366,8367,8368,8369,8370,8371,8372,8373,
                8374,8375,8376,8377,8378,8379,8380,8381,8382,8383,
                8384,8386,8385,8387,8388,8389,8390,8391,8392,8393,
                8394,8395,8396,8397,8398,8400,8399,8401,8403,8402,
                8404,8406,8405,8407,8408,8409,8411,8410,8412,8413,
                8414,8415,8416,8417,8418,8419,8420,8421,8422,8423,
                8424,8425,8426,8427,8428,8429,8430,8431,8432,8433,
                8434,8435,8436,8437,8438,8439,8440,8441,8442,8443,
                8444,8445,8446,8447,8448,8449,8450,8451,8452,8453,
                8454,8457,8455,8456,8458,8459,8460,8461,8463,8464,
                8462,8465,8466,8467,8468,8469,8470,8471,8472,8473,
                8474,8475,8476,8477,8478,8479,8480,8481,8482,8483,
                8484,8485,8486,8487,8488,8489,8490,8491,8492,8493,
                8494,8495,8496,8497,8498,8499,8500,8501,8502,8503,
                8504,8505,8507,8506,8508,8509,8510,8511,8512,8513,
                8515,8514,8516,8517,8518,8520,8519,8521,8522,8523,
                8525,8524,8526,8527,8528,8530,8529,8531,8532,8533,
                8534,8535,8536,8537,8538,8539,8540,8541,8542,8543,
                8544,8546,8545,8547,8549,8548,8551,8552,8554,8553,
                8555,8556,8557,8558,8559,8560,8561,8562,8563,8564,
                8565,8566,8568,8567,8569,8570,8571,8572,8573,8574,
                8575,8576,8577,8578,8579,8580,8581,8582,8583,8585,
                8584,8586,8587,8588,8589,8590,8591,8592,8593,8594,
                8596,8595,8598,8597,8599,8600,8601,8602,8603,8604,
                8605,8606,8607,8608,8609,8610,8611,8612,8613,8614,
                8615,8616,8617,8618,8619,8620,8621,8622,8623,8624,
                8626,8625,8627,8628,8629,8630,8631,8632,8633,8634,
                8635,8636,8637,8638,8639,8640,8641,8642,8643,8644,
                8645,8646,8647,8648,8649,8650,8651,8652,8653,8654,
                8655,8656,8657,8658,8659,8660,8661,8662,8663,8664,
                8665,8666,8667,8668,8670,8669,8671,8672,8673,8675,
                8674,8676,8677,8680,8678,8679,8681,8682,8684,8683,
                8686,8685,8687,8689,8688,8690,8691,8692,8693,8694,
                8695,8696,8697,8698,8699,8700,8701,8702,8703,8704,
                8705,8706,8707,8708,8709,8710,8711,8712,8713,8714,
                8715,8716,8717,8718,8719,8720,8721,8722,8723,8724,
                8725,8726,8727,8728,8729,8730,8731,8732,8733,8734,
                8735,8737,8736,8738,8739,8740,8741,8742,8744,8743,
                8747,8746,8745,8748,8749,8750,8752,8754,8753,8751,
                8755,8757,8756,8759,8758,8760,8761,8762,8763,8764,
                8765,8766,8767,8768,8769,8770,8771,8772,8773,8774,
                8775,8776,8778,8777,8779,8780,8781,8782,8783,8784,
                8785,8786,8787,8788,8789,8790,8791,8792,8793,8794,
                8795,8796,8797,8798,8799,8800,8801,8802,8803,8804,
                8805,8806,8807,8808,8809,8810,8811,8812,8813,8814,
                8815,8816,8818,8817,8819,8820,8821,8822,8824,8823,
                8826,8825,8827,8828,8829,8830,8831,8832,8833,8834,
                8835,8836,8837,8838,8839,8840,8841,8842,8843,8844,
                8845,8846,8847,8848,8849,8850,8851,8852,8853,8854,
                8855,8856,8857,8858,8859,8860,8862,8861,8863,8864,
                8866,8865,8867,8868,8869,8870,8871,8872,8873,8874,
                8875,8876,8877,8878,8879,8881,8880,8882,8883,8884,
                8885,8886,8888,8887,8889,8890,8891,8892,8893,8896,
                8894,8895,8897,8898,8899,8900,8901,8902,8903,8904,
                8905,8906,8907,8908,8909,8910,8911,8912,8913,8914,
                8916,8915,8917,8918,8919,8920,8921,8922,8923,8924,
                8925,8926,8927,8929,8928,8930,8931,8932,8933,8934,
                8936,8935,8937,8938,8939,8940,8941,8943,8942,8944,
                8945,8946,8947,8948,8949,8950,8951,8952,8953,8954
#endif
        };
        
        int nTotalSeqCount = sizeof(seqs) / sizeof(int);
        for (int i = 0; i < nTotalSeqCount; i++) {
                char buf[20] = {0};
                sprintf(buf, "%d %d", seqs[i], seqs[i]);
                int nIsDiscard;
                JitterBufferPush(&jb, buf, strlen(buf), seqs[i], seqs[i]+10000, &nIsDiscard);
                if (nIsDiscard) {
                        MY_PJ_LOG(3, "seq:%d dicarded", seqs[i]);
                }
                printf("======>push frame:seq:%d  idx:%03d jblen:%d\n", seqs[i], i, jb.nCurrentSize);
                
                pj_bool_t bGetFrame = PJ_TRUE;
                while(bGetFrame) {
                        char getBuf[257] = {0};
                        int nFrameSize = sizeof(getBuf) - 1;
                        uint32_t nTs;
                        uint32_t nFrameSeq = 0;
                        JBFrameStatus popFrameType;
                        JitterBufferPop(&jb, getBuf, &nFrameSize, &nFrameSeq, &nTs, &popFrameType);
                        
                        switch (popFrameType) {
                                case JBFRAME_STATE_MISSING:
                                        pj_thread_sleep(50);
                                        printf("missing\n");
                                        bGetFrame = PJ_FALSE;
                                        break;
                                case JBFRAME_STATE_CACHING:
                                        printf("JBFRAME_STATE_CACHING\n");
                                        bGetFrame = PJ_FALSE;
                                        break;
                                case JBFRAME_STATE_EMPTY:
                                        printf("JBFRAME_STATE_EMPTY\n");
                                        bGetFrame = PJ_FALSE;
                                        break;
                                case JBFRAME_STATE_NORMAL:
                                        printf("-->get one frame:seq:%04d, size:%02d, ts:%05d, content:%s\n", nFrameSeq, nFrameSize, nTs, getBuf);
                                        bGetFrame = PJ_TRUE;
                                        break;
                        }
                }
                pj_thread_sleep(50);
        }
}

//end jitter buffer
int main(int argc, char **argv)
{
        if(argc == 1){
                printf("usage as:%s (1 for offer|2 for answer) [1for audio|2for video|3 audio and vido] [turnip] [tcp|udp]\n", argv[0]);
                return -1;
        }
        int role = atoi(argv[1]);
        if(role != OFFER && role != ANSWER){
                printf("usage as:%s (1 for offer|2 for answer)  [1for audio|2for video|3 audio and vido] [turnip]  [tcp|udp]\n", argv[0]);
                return -1;
        }
        int hasAudioVideo = 3;
        if (argc > 2) {
                hasAudioVideo = atoi(argv[2]);
        }
        if (hasAudioVideo < 0 || hasAudioVideo > 3) {
                printf("usage as:%s (1 for offer|2 for answer)  [1for audio|2for video|3 audio and vido] [turnip]  [tcp|udp]\n", argv[0]);
                return -1;
        }
        char * pTurnHost = NULL;
        if (argc > 3) {
                pTurnHost = argv[3];
        }
        
        
        pj_status_t status;
        status = pj_init();
        PJ_ASSERT_RETURN(status == PJ_SUCCESS, 1);
        status = pjlib_util_init();
        PJ_ASSERT_RETURN(status == PJ_SUCCESS, 1);

        pj_caching_pool_init(&app.cachingPool, &pj_pool_factory_default_policy, 0);

        set_log_to_file(role);
        
        //test_jitter_buffer();
        //return 0;

        test_sdp_neg();
        //---------------------start------------

        //offer send to answer
        if (role == ANSWER) {
                //test receive pcmu
                pj_pool_t * apool = pj_pool_create(&app.cachingPool.factory, "rxrtpa", 2000, 2000, NULL);
                status = pj_file_open(apool, "/Users/liuye/Documents/p2p/build/src/work/Debug/rxrtp.mulaw", PJ_O_WRONLY, &gPcmuFd);
                if(status != PJ_SUCCESS){
                        printf("pj_file_open fail:%d\n", status);
                        return status;
                }
                //end test recive h264
                //test receive pcmu
                pj_pool_t * vpool = pj_pool_create(&app.cachingPool.factory, "rxrtpv", 2000, 2000, NULL);
                status = pj_file_open(vpool, "/Users/liuye/Documents/p2p/build/src/work/Debug/rxrtp.h264", PJ_O_WRONLY, &gH264Fd);
                if(status != PJ_SUCCESS){
                        printf("pj_file_open fail:%d\n", status);
                        return status;
                }
                //end test recive h264
        }
        
        InitIceConfig(&app.userConfig);
        app.userConfig.nForceStun = 1;
        strcpy(app.userConfig.turnHost, "123.59.204.198");
        strcpy(app.userConfig.turnUsername, "root");
        strcpy(app.userConfig.turnPassword, "root");
        if (pTurnHost != NULL) {
                strcpy(app.userConfig.turnHost, pTurnHost);
        }
        if (argc > 4) {
                if (strcmp(argv[4], "tcp") == 0) {
                        MY_PJ_LOG(3, "use tcp");
                        app.userConfig.bTurnTcp = 1;
                }
        }
        app.userConfig.userCallback = onRxRtp;
        status = InitPeerConnectoin(&app.pPeerConnection, &app.userConfig);
        pj_assert(status == 0);
        
        
        if(hasAudioVideo & HAS_AUDIO){
                InitMediaConfig(&app.audioConfig);
                app.audioConfig.configs[0].nSampleOrClockRate = 8000;
                app.audioConfig.configs[0].codecType = MEDIA_FORMAT_PCMU;
                app.audioConfig.configs[0].nChannel = 1;
                app.audioConfig.configs[1].nSampleOrClockRate = 8000;
                app.audioConfig.configs[1].codecType = MEDIA_FORMAT_PCMA;
                app.audioConfig.configs[1].nChannel = 1;
                app.audioConfig.nCount = 2;
                if ( role == ANSWER ){
                        app.audioConfig.configs[0] = app.audioConfig.configs[1];
                        app.audioConfig.configs[1].nSampleOrClockRate = 8000;
                        app.audioConfig.configs[1].codecType = MEDIA_FORMAT_G729;
                }
                status = AddAudioTrack(app.pPeerConnection, &app.audioConfig);
                TESTCHECK(status, app);
        }
        
        pj_pool_t * pRemoteSdpPool = NULL;
        char textSdpBuf[2048] = {0};
        if(hasAudioVideo & HAS_VIDEO){
                InitMediaConfig(&app.videoConfig);
                app.videoConfig.configs[0].nSampleOrClockRate = 90000;
                app.videoConfig.configs[0].codecType = MEDIA_FORMAT_H264;
                app.videoConfig.configs[1].nSampleOrClockRate = 90000;
                app.videoConfig.configs[1].codecType = MEDIA_FORMAT_H265;
                app.videoConfig.nCount = 2;
                if ( role == ANSWER ){
                        app.videoConfig.nCount = 1;
                        //app.videoConfig.configs[0] = app.videoConfig.configs[1];
                }
                status = AddVideoTrack(app.pPeerConnection, &app.videoConfig);
                TESTCHECK(status, app);
        }
        

        if (role == OFFER) {
                status = createOffer(app.pPeerConnection);
                TESTCHECK(status, app);
                waitState(ICE_STATE_INIT);
                
                pjmedia_sdp_session *pAnswer = NULL;
                pRemoteSdpPool =pj_pool_create(&app.cachingPool.factory, "sdpremote", 2048, 1024, NULL);
                sdp_from_file(&pAnswer, ANSWERFILE,  pRemoteSdpPool, textSdpBuf, sizeof(textSdpBuf));
                setRemoteDescription(app.pPeerConnection, pAnswer);
        }
        
        if (role == ANSWER) {
                pjmedia_sdp_session *pOffer = NULL;
                pRemoteSdpPool =pj_pool_create(&app.cachingPool.factory, "sdpremote", 2048, 1024, NULL);
                sdp_from_file(&pOffer, OFFERFILE,  pRemoteSdpPool, textSdpBuf, sizeof(textSdpBuf));
                setRemoteDescription(app.pPeerConnection, pOffer);

                status = createAnswer(app.pPeerConnection, pOffer);
                TESTCHECK(status, app);
                waitState(ICE_STATE_INIT);
        }
        
        input_confirm("confirm to negotiation:");
        StartNegotiation(app.pPeerConnection);

        if (role == ANSWER) {
#if 0
                char packet[120];
                while(1){
                        memset(packet, 0, sizeof(packet));
                        memset(packet, 0x30, 12);
                        printf("input:");
                        scanf("%s", packet+12);
                        if(packet[12] == 'q'){
                                break;
                        }
                        //pjmedia_transport_send_rtp(app.pPeerConnection->transportIce[0].pTransport, packet, strlen(packet));
                        memset(packet, 0x31, 12);
                        //pjmedia_transport_send_rtp(app.pPeerConnection->transportIce[1].pTransport, packet, strlen(packet));
                }
#endif
        } else {
                input_confirm("confirm to sendfile:");
                if ((hasAudioVideo & HAS_AUDIO) && (hasAudioVideo & HAS_VIDEO)) {
                        start_file_test("/Users/liuye/Documents/p2p/build/src/mysiprtp/Debug/8000_1.mulaw",
                                        "/Users/liuye/Documents/p2p/build/src/mysiprtp/Debug/hks.h264",
                                        receive_data_callback);
                } else if (hasAudioVideo & HAS_AUDIO) {
                        start_file_test("/Users/liuye/Documents/p2p/build/src/mysiprtp/Debug/8000_1.mulaw",
                                        NULL, receive_data_callback);
                } else {
                        start_file_test(NULL, "/Users/liuye/Documents/p2p/build/src/mysiprtp/Debug/hks.h264",
                                        receive_data_callback);
                }
        }
        
        input_confirm("quit");
        if (role == ANSWER) {
                pj_file_close(gPcmuFd);
#ifdef HAS_VIDEO
                pj_file_close(gH264Fd);
#endif
        }
        ReleasePeerConnectoin(app.pPeerConnection);
        return 0;
}
