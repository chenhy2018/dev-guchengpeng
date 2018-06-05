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
        MediaConfig audioConfig;
        MediaConfig videoConfig;
        IceConfig userConfig;
}App;
App app;

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


pj_oshandle_t gPcmuFd;
pj_oshandle_t gH264Fd;
static void onRxRtp(void *_pUserData, CallbackType _type, void *_pCbData)
{
        switch (_type){
                case CALLBACK_ICE:{
                        IceNegInfo *pInfo = (IceNegInfo *)_pCbData;
                        fprintf(stderr, "==========>callback_ice: state:%d\n", pInfo->state);
                        for ( int i = 0; i < pInfo->nCount; i++) {
                                fprintf(stderr, "           codec type:%d\n", pInfo->configs[i]->codecType);
                        }
                }
                        break;
                case CALLBACK_RTP:{
                        RtpPacket *pPkt = (RtpPacket *)_pCbData;
                        pj_ssize_t nLen = pPkt->nDataLen;
                        if (pPkt->type == STREAM_AUDIO && nLen == 160) {
                                pj_file_write(gPcmuFd, pPkt->pData, &nLen);
                        } else if (pPkt->type == STREAM_VIDEO) {
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
                printf("send %d bytes audio data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                rtpPacket.type = STREAM_AUDIO;
        } else {
                printf("send %d bytes vidoe data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                rtpPacket.type = STREAM_VIDEO;
        }
        rtpPacket.pData = (uint8_t *)pData;
        rtpPacket.nDataLen = nDataLen;
        rtpPacket.nTimestamp = timestamp;
        return SendPacket(app.pPeerConnection, &rtpPacket);
}

char * pLogFileName = NULL;
pj_oshandle_t gLogFd;
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
                app.audioConfig.configs[1].nSampleOrClockRate = 8000;
                app.audioConfig.configs[1].codecType = MEDIA_FORMAT_PCMA;
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
                pjmedia_sdp_session *pOffer = NULL;
                status = createOffer(app.pPeerConnection, &pOffer);
                TESTCHECK(status, app);
                setLocalDescription(app.pPeerConnection, pOffer);
                write_sdp(pOffer, OFFERFILE);
                
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
                
                pjmedia_sdp_session *pAnswer = NULL;
                status = createAnswer(app.pPeerConnection, pOffer, &pAnswer);
                TESTCHECK(status, app);
                setLocalDescription(app.pPeerConnection, pAnswer);
                write_sdp(pAnswer, ANSWERFILE);
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
