// Last Update:2018-06-05 22:52:26
/**
 * @file test.c
 * @brief 
 * @author
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
#include "sdk_interface_p2p.h"
#include "dbg.h"
#include "unit_test.h"
#include <unistd.h> 
#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>
#include <assert.h>
#include <sys/time.h>

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))

int RegisterTestSuitCallback( TestSuit *this );
int RegisterTestSuitInit( TestSuit *this );
int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase );

typedef struct {
    char *id;
    char *password;
    char *sigHost;
    char *mediaHost;
    char *imHost;
    int timeOut;
    unsigned char init;
    int accountid;
    int callid;
    int64_t timecount;
    int count;
    int misscount;
} RegisterData;
int allcount = 0;

typedef struct {
    TestCase father;
    RegisterData data;
} RegisterTestCase;
#define MAX_COUNT 1
#define HOST "39.107.247.14"
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", 0 },
        { "1741", "6fKzQXiH", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1016", "1016", HOST, HOST, HOST, 100, 0 }
    },
    {   
        { "normal", 0 },
        { "1017", "1017", HOST, HOST, HOST, 100, 1 }
    },
    {   
        { "invalid_account", 0 },
        { "1018", "1018", HOST, HOST, HOST, 100, 0 }
    },
    {   
        { "normal", 0 },
        { "1019", "1019", HOST, HOST, HOST, 100, 1 }
    },
};

TestSuit gRegisterTestSuit =
{
    "Register",
    RegisterTestSuitCallback,
    RegisterTestSuitInit,
    RegisterTestSuitGetTestCase,
    (void*)&gRegisterTestCases,
};


int RegisterTestSuitInit( TestSuit *this )
{
    this->total = ARRSZ(gRegisterTestCases);
    this->index = 0;

    return 0;
}

int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    RegisterTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}
#define THIS_IS_AUDIO 1
#define THIS_IS_VIDEO 2
typedef int (*DataCallback)(void *pData, int nDataLen, int nFlag, int64_t timestamp, AccountID id, int callID);
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

int start_file_test(char * _pAudioFile, char * _pVideoFile, DataCallback callback, AccountID id, int callID)
{
        assert(!(_pAudioFile == NULL && _pVideoFile == NULL));

        int ret;

        char * pAudioData = NULL;
        int nAudioDataLen = 0;
        if(_pAudioFile != NULL){
                ret = readFileToBuf(_pAudioFile, &pAudioData, &nAudioDataLen);
                if (ret != 0) {
                        DBG_LOG( "map data to buffer fail:%s", _pAudioFile);
                        return -1;
                }
        }

        char * pVideoData = NULL;
        int nVideoDataLen = 0;
        if(_pVideoFile != NULL){
                ret = readFileToBuf(_pVideoFile, &pVideoData, &nVideoDataLen);
                if (ret != 0) {
                        free(pAudioData);
                        DBG_LOG( "map data to buffer fail:%s", _pVideoFile);
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
                                cbRet = callback(pAudioData + audioOffset, 160, THIS_IS_AUDIO, nNextAudioTime-nSysTimeBase, id, callID);
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
                                        DBG_LOG( "send one video(%d) frame packet:%ld", type, end - sendp);
                                        cbRet = callback(sendp, end - sendp, THIS_IS_VIDEO, nNextVideoTime-nSysTimeBase, id, callID);
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
                        DBG_LOG( "sleeptime:%ld\n", nSleepTime);
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

static int receive_data_callback(void *pData, int nDataLen, int nFlag, int64_t timestamp, AccountID id, int callID)
{
        if (nFlag == THIS_IS_AUDIO) {
                DBG_LOG("send %d bytes audio data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                return SendPacket(id, callID, STREAM_AUDIO, pData, nDataLen, timestamp);
        } else {
                DBG_LOG("send %d bytes vidoe data to rtp with timestamp:%ld\n", nDataLen, timestamp);
                return SendPacket(id, callID, STREAM_VIDEO, pData, nDataLen, timestamp);
        }
        return 0;
}


void Mthread1(void* data)
{
    RegisterData *pData = (RegisterData*) data;
    Event * event= (Event*) malloc(sizeof(Event));
    ErrorID id;
    EventType type;
    while (1) {
                    id = PollEvent(pData->accountid, &type, &event, 0);
                    if (id != RET_OK) {
                           continue;
                    }
                    switch (type) {
                            case EVENT_CALL:
                            {
                                  CallEvent *pCallEvent = &(event->body.callEvent);
                                  DBG_LOG("Call status %d call id %d call account id %d\n", pCallEvent->status, pCallEvent->callID, pData->accountid);
                                  if (pCallEvent->status == CALL_STATUS_INCOMING) {
                                      DBG_LOG("AnswerCall ******************\n");
                                      if (pCallEvent->callID == 6) {
                                              RejectCall(pData->accountid, pCallEvent->callID);
                                      }
                                      else {
                                              AnswerCall(pData->accountid, pCallEvent->callID);
                                      }
                                      DBG_LOG("AnswerCall end *****************\n");
                                  }
                                  if (pCallEvent->status == CALL_STATUS_ESTABLISHED) {
                                        MediaInfo* info = (MediaInfo *)pCallEvent->context;
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d mediacount %d, type 1 %d type 2 %d\n",
                                                 pCallEvent->callID, pData->accountid, info->nCount, info->media[0].codecType, info->media[1].codecType);
                                        start_file_test(
                                            "/opt2/a.mulaw", "/opt2/v.h264",
                                            ////"/opt2/720p.mulaw", "/opt2/720p.h264",
                                            receive_data_callback, pData->accountid, pCallEvent->callID);
                                  }
                                  break;
                            }
                            case EVENT_DATA:
                            {
                                  DataEvent *pDataEvent = &(event->body.dataEvent);
                                  allcount += 1;
                                  //DBG_LOG("Data size %d call id %d call account id %d timestamp %lld \n", pDataEvent->size, pDataEvent->callID, pData->accountid, pDataEvent->pts);
                                  if (pData->timecount == 0) {
                                         pData->timecount = pDataEvent->pts;
                                  }
                                  else {

                                         if (allcount %10 == 0) {
                                         //        pData->misscount += pDataEvent->pts - pData->timecount;
                                                 DBG_ERROR("***miss %d*****size %d****error timestamp %ld last timestamp %ld callid %d count %d \n",
                                                pData->misscount, pDataEvent->size, pDataEvent->pts, pData->timecount, pDataEvent->callID, allcount);
                                         }
                                         pData->timecount = pDataEvent->pts;
                                  }
                                  break;
                            }
                            case EVENT_MESSAGE:
                            {
                                  MessageEvent *pMessage = &(event->body.messageEvent);
                                  DBG_LOG("Message %s status id %d account id %d\n", pMessage->message, pMessage->status, pData->accountid);
                                  break;
                            }
                    }
           }
}
int RegisterTestSuitCallback( TestSuit *this )
{
    RegisterTestCase *pTestCases = NULL;
    RegisterData *pData = NULL;
    Media media[2];
    media[1].streamType = STREAM_VIDEO;
    media[1].codecType = CODEC_H264;
    media[1].sampleRate = 90000;
    media[1].channels = 0;
    media[0].streamType = STREAM_AUDIO;
    media[0].codecType = CODEC_G711U;
    media[0].sampleRate = 8000;
    media[0].channels = 1;

    int i = 0;
    ErrorID sts = 0;

    if ( !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *) this->testCases;
    DBG_LOG("this->index = %d\n", this->index );
    pData = &pTestCases[0].data;

    if ( pData->init ) {
        DBG_LOG("InitSDK");
        sts = InitSDK( &media[0], 2 );
        if ( RET_OK != sts ) {
            DBG_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }
        pthread_t t_1;
        pthread_attr_t attr_1;
        pthread_attr_init(&attr_1);
        pthread_attr_setdetachstate(&attr_1, PTHREAD_CREATE_DETACHED);

    for (int count = 0; count < MAX_COUNT; ++count) {
            pData = &pTestCases[count].data;
            pData->timecount = 0;
            pData->count = 0;
            pData->misscount = 0;
            DBG_STR( pData->id );
            DBG_STR( pData->password );
            DBG_STR( pData->sigHost );
            DBG_LOG("Register in\n");
            pData->accountid = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost);
            DBG_LOG("Register out %x %x\n", pData->accountid, pTestCases->father.expact);
            pthread_create(&t_1, &attr_1, Mthread1, pData);
    }
    pData->callid = -1;
    ErrorID id;
    EventType type;

    //sleep(10);
    int count = 0;
    while (1) { sleep(100); }
}

int InitAllTestSuit()
{
    AddTestSuit( &gRegisterTestSuit );

    return 0;
}

int main()
{
    DBG_LOG("+++++ enter main...\n");
    TestSuitManagerInit();
    InitAllTestSuit();
    RunAllTestSuits();

    return 0;
}

