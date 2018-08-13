#include <stdio.h>
#include <libavformat/avformat.h>
#include <assert.h>
#include <sys/time.h>
#include <unistd.h>
#include "tsuploaderapi.h"
#include "adts.h"
AvArg avArg;

typedef int (*DataCallback)(void *opaque, void *pData, int nDataLen, int nFlag, int64_t timestamp, int nIsKeyFrame);
#define THIS_IS_AUDIO 1
#define THIS_IS_VIDEO 2
//#define TEST_AAC 1
#define TEST_AAC_NO_ADTS 1
#define USE_LINK_ACC 1


FILE *outTs;
int gTotalLen = 0;
char gtestToken[1024] = {0};

// start aac
static int aacfreq[13] = {96000, 88200,64000,48000,44100,32000,24000, 22050 , 16000 ,12000,11025,8000,7350};
typedef struct ADTS{
        ADTSFixheader fix;
        ADTSVariableHeader var;
}ADTS;
//end aac

enum HEVCNALUnitType {
        HEVC_NAL_TRAIL_N    = 0,
        HEVC_NAL_TRAIL_R    = 1,
        HEVC_NAL_TSA_N      = 2,
        HEVC_NAL_TSA_R      = 3,
        HEVC_NAL_STSA_N     = 4,
        HEVC_NAL_STSA_R     = 5,
        HEVC_NAL_RADL_N     = 6,
        HEVC_NAL_RADL_R     = 7,
        HEVC_NAL_RASL_N     = 8,
        HEVC_NAL_RASL_R     = 9,
        HEVC_NAL_VCL_N10    = 10,
        HEVC_NAL_VCL_R11    = 11,
        HEVC_NAL_VCL_N12    = 12,
        HEVC_NAL_VCL_R13    = 13,
        HEVC_NAL_VCL_N14    = 14,
        HEVC_NAL_VCL_R15    = 15,
        HEVC_NAL_BLA_W_LP   = 16,
        HEVC_NAL_BLA_W_RADL = 17,
        HEVC_NAL_BLA_N_LP   = 18,
        HEVC_NAL_IDR_W_RADL = 19,
        HEVC_NAL_IDR_N_LP   = 20,
        HEVC_NAL_CRA_NUT    = 21,
        HEVC_NAL_IRAP_VCL22 = 22,
        HEVC_NAL_IRAP_VCL23 = 23,
        HEVC_NAL_RSV_VCL24  = 24,
        HEVC_NAL_RSV_VCL25  = 25,
        HEVC_NAL_RSV_VCL26  = 26,
        HEVC_NAL_RSV_VCL27  = 27,
        HEVC_NAL_RSV_VCL28  = 28,
        HEVC_NAL_RSV_VCL29  = 29,
        HEVC_NAL_RSV_VCL30  = 30,
        HEVC_NAL_RSV_VCL31  = 31,
        HEVC_NAL_VPS        = 32,
        HEVC_NAL_SPS        = 33,
        HEVC_NAL_PPS        = 34,
        HEVC_NAL_AUD        = 35,
        HEVC_NAL_EOS_NUT    = 36,
        HEVC_NAL_EOB_NUT    = 37,
        HEVC_NAL_FD_NUT     = 38,
        HEVC_NAL_SEI_PREFIX = 39,
        HEVC_NAL_SEI_SUFFIX = 40,
};
enum HevcType {
        HEVC_META = 0,
        HEVC_I = 1,
        HEVC_B =2
};

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

static inline int64_t getCurrentMilliSecond(){
        struct timeval tv;
        gettimeofday(&tv, NULL);
        return (tv.tv_sec*1000 + tv.tv_usec/1000);
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

static int is_h265_picture(int t)
{
        switch (t) {
                case HEVC_NAL_VPS:
                case HEVC_NAL_SPS:
                case HEVC_NAL_PPS:
                case HEVC_NAL_SEI_PREFIX:
                        return HEVC_META;
                case HEVC_NAL_IDR_W_RADL:
                case HEVC_NAL_CRA_NUT:
                        return HEVC_I;
                case HEVC_NAL_TRAIL_N:
                case HEVC_NAL_TRAIL_R:
                case HEVC_NAL_RASL_N:
                case HEVC_NAL_RASL_R:
                        return HEVC_B;
                default:
                        return -1;
        }
}

int start_file_test(char * _pAudioFile, char * _pVideoFile, DataCallback callback, void *opaque)
{
        assert(!(_pAudioFile == NULL && _pVideoFile == NULL));
        
        int ret;
        
        char * pAudioData = NULL;
        int nAudioDataLen = 0;
        if(_pAudioFile != NULL){
                ret = readFileToBuf(_pAudioFile, &pAudioData, &nAudioDataLen);
                if (ret != 0) {
                        printf("map data to buffer fail:%s", _pAudioFile);
                        return -1;
                }
        }
        
        char * pVideoData = NULL;
        int nVideoDataLen = 0;
        if(_pVideoFile != NULL){
                ret = readFileToBuf(_pVideoFile, &pVideoData, &nVideoDataLen);
                if (ret != 0) {
                        free(pAudioData);
                        printf( "map data to buffer fail:%s", _pVideoFile);
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
        int isAAC = 0;
        int64_t aacFrameCount = 0;
        if (memcmp(_pAudioFile + strlen(_pAudioFile) - 3, "aac", 3) == 0)
                isAAC = 1;
        while (bAudioOk || bVideoOk) {
                if (bAudioOk && nNow+1 > nNextAudioTime) {
                        if (isAAC) {
                                ADTS adts;
                                if(audioOffset+7 <= nAudioDataLen) {
                                        ParseAdtsfixedHeader((unsigned char *)(pAudioData + audioOffset), &adts.fix);
                                        int hlen = adts.fix.protection_absent == 1 ? 7 : 9;
                                        ParseAdtsVariableHeader((unsigned char *)(pAudioData + audioOffset), &adts.var);
                                        if (audioOffset+hlen+adts.var.aac_frame_length <= nAudioDataLen) {
#ifdef TEST_AAC_NO_ADTS
                                                cbRet = callback(opaque, pAudioData + audioOffset + hlen, adts.var.aac_frame_length - hlen,
                                                                 THIS_IS_AUDIO, nNextAudioTime-nSysTimeBase, 0);
#else
                                                cbRet = callback(opaque, pAudioData + audioOffset, adts.var.aac_frame_length,
                                                                 THIS_IS_AUDIO, nNextAudioTime-nSysTimeBase, 0);
#endif
                                                if (cbRet != 0) {
                                                        bAudioOk = 0;
                                                        continue;
                                                }
                                                audioOffset += adts.var.aac_frame_length;
                                                aacFrameCount++;
                                                int64_t d = ((1024*1000.0)/aacfreq[adts.fix.sampling_frequency_index]) * aacFrameCount;
                                                nNextAudioTime = nSysTimeBase + d;
                                        } else {
                                                bAudioOk = 0;
                                        }
                                } else {
                                        bAudioOk = 0;
                                }
                        } else {
                                if(audioOffset+160 <= nAudioDataLen) {
                                        cbRet = callback(opaque, pAudioData + audioOffset, 160, THIS_IS_AUDIO, nNextAudioTime-nSysTimeBase, 0);
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
                                
                                if (avArg.nVideoFormat == TK_VIDEO_H264) {
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
                                                //printf("send one video(%d) frame packet:%ld", type, end - sendp);
                                                cbRet = callback(opaque, sendp, end - sendp, THIS_IS_VIDEO, nNextVideoTime-nSysTimeBase, type == 5);
                                                if (cbRet != 0) {
                                                        bVideoOk = 0;
                                                }
                                                nNextVideoTime += 40;
                                                break;
                                        }
                                }else{
                                        int dlen = 3;
                                        if(start[2] == 0x01){//0x 00 00 01
                                                type = start[3] & 0x7E;
                                        }else{ // 0x 00 00 00 01
                                                dlen = 4;
                                                type = start[4] & 0x7E;
                                        }
                                        type = (type >> 1);
                                        int hevctype = is_h265_picture(type);
                                        if (hevctype == -1) {
                                                printf("unknown type:%d\n", type);
                                                continue;
                                        }
                                        //printf("%d------------->%d\n",dlen, type);
                                        if(hevctype == HEVC_I || hevctype == HEVC_B ){
                                                if (type == 20) {
                                                        nNonIDR++;
                                                } else {
                                                        nIDR++;
                                                }
                                                //printf("send one video(%d) frame packet:%ld", type, end - sendp);
                                                cbRet = callback(opaque, sendp, end - sendp, THIS_IS_VIDEO, nNextVideoTime-nSysTimeBase, hevctype == HEVC_I);
                                                if (cbRet != 0) {
                                                        bVideoOk = 0;
                                                }
                                                nNextVideoTime += 40;
                                                break;
                                        }
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
                        //printf("sleeptime:%lld\n", nSleepTime);
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

static int64_t firstTimeStamp = -1;
static int segStartCount = 0;
static int nByteCount = 0;

static int dataCallback(void *opaque, void *pData, int nDataLen, int nFlag, int64_t timestamp, int nIsKeyFrame)
{
        int ret = 0;
        nByteCount += nDataLen;
        if (nFlag == THIS_IS_AUDIO){
                //fprintf(stderr, "push audio ts:%lld\n", timestamp);
                ret = PushAudio(pData, nDataLen, timestamp);
        } else {
                if (firstTimeStamp == -1){
                        firstTimeStamp = timestamp;
                }
                int nNewSegMent = 0;
                if (nIsKeyFrame && timestamp - firstTimeStamp > 30000 && segStartCount == 0) {
                        nNewSegMent = 1;
                        segStartCount++;
                }
                //fprintf(stderr, "push video ts:%lld\n", timestamp);
                ret = PushVideo(pData, nDataLen, timestamp, nIsKeyFrame, nNewSegMent);
        }
        return ret;
}

static void * upadateToken() {
        int ret = 0;
        while(1) {
                sleep(30);
                ret = GetUploadToken(gtestToken, sizeof(gtestToken));
                if (ret != 0) {
                        printf("update token file<<<<<<<<<<<<<\n");
                        return NULL;
                }
                printf("token:%s\n", gtestToken);
                ret = UpdateToken(gtestToken);
                if (ret != 0) {
                        printf("update token file<<<<<<<<<<<<<\n");
                        return NULL;
                }
        }
        return NULL;
}


int main(int argc, char* argv[])
{

        int ret = 0;
        
        SetLogLevelToDebug();
        
        pthread_t updateTokenThread;
        pthread_attr_t attr;
        pthread_attr_init (&attr);
        pthread_attr_setdetachstate (&attr, PTHREAD_CREATE_DETACHED);
        ret = pthread_create(&updateTokenThread, &attr, upadateToken, NULL);
        if (ret != 0) {
                printf("create update token thread fail\n");
                return ret;
        }
        pthread_attr_destroy (&attr);
        
#ifdef USE_LINK_ACC
        SetAk("JAwTPb8dmrbiwt89Eaxa4VsL4_xSIYJoJh4rQfOQ");
        SetSk("G5mtjT3QzG4Lf7jpCAN5PZHrGeoSH9jRdC96ecYS");
        
        //计算token需要，所以需要先设置
        SetBucketName("ipcamera");
#else
        SetAk("p4y2X-zKqiDoWeIvmhhXkR3mHbHx_Yw36YHFz9e1");
        SetSk("hVZOIZfLw_0xzJhtVv1ddmlSkbiUrHLdNJewZRkp");
        SetBucketName("bucket");
#endif

        
        ret = GetUploadToken(gtestToken, sizeof(gtestToken));
        if (ret != 0)
                return ret;
        printf("token:%s\n", gtestToken);
        
#ifdef TEST_AAC
        avArg.nAudioFormat = TK_AUDIO_AAC;
        avArg.nChannels = 1;
        avArg.nSamplerate = 16000;
#else
        avArg.nAudioFormat = TK_AUDIO_PCMU;
        avArg.nChannels = 1;
        avArg.nSamplerate = 8000;
#endif
        avArg.nVideoFormat = TK_VIDEO_H264;

        ret = InitUploader("testuid3", "testdeviceid", gtestToken, &avArg);
        if (ret != 0) {
                return ret;
        }
#ifdef __APPLE__
        char * pVFile = "/Users/liuye/Documents/material/h265_aac_1_16000_h264.h264";
  #ifdef TEST_AAC
        char * pAFile = "/Users/liuye/Documents/material/h265_aac_1_16000_a.aac";
  #else
        char * pAFile = "/Users/liuye/Documents/material/h265_aac_1_16000_pcmu_8000.mulaw";
  #endif
        if (avArg.nVideoFormat == TK_VIDEO_H265) {
                pVFile = "/Users/liuye/Documents/material/h265_aac_1_16000_v.h265";
        }
        start_file_test(pAFile, pVFile, dataCallback, NULL);

#else

        char * pVFile = "/liuye/Documents/material/h265_aac_1_16000_h264.h264";
  #ifdef TEST_AAC
        char * pAFile = "/liuye/Documents/material/h265_aac_1_16000_a.aac";
  #else
        char * pAFile = "/liuye/Documents/material/h265_aac_1_16000_pcmu_8000.mulaw";
  #endif
        if (avArg.nVideoFormat == TK_VIDEO_H265) {
                pVFile = "/liuye/Documents/material/h265_aac_1_16000_v.h265";
        }
        start_file_test(pAFile, pVFile, dataCallback, NULL);
#endif
        
        
        UninitUploader();
        loginfo("should total:%d\n", nByteCount);

        return 0;
}