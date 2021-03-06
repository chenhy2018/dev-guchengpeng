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
#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif

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

#define CHENGPENG 1
//#define MENGKE 1

#ifdef CHENGPENG
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", 0 },
        { "2040", "vGrMmyIs", HOST, HOST, HOST, 100, 1 }
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
#endif

#ifdef MENGKE
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", 0 },
        { "1710", "KR3Cw6m4", HOST, HOST, HOST, 100, 1 }
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
#endif

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
                                      if (pCallEvent->callID < 0) {
                                              DBG_LOG("RejectCall ******************\n");
                                              RejectCall(pData->accountid, pCallEvent->callID);
                                      }
                                      else {
                                              AnswerCall(pData->accountid, pCallEvent->callID);
                                      }
                                      DBG_LOG("AnswerCall end *****************\n");
                                  }
                                  if (pCallEvent->status == CALL_STATUS_ESTABLISHED) {
#ifdef WITH_P2P
                                        MediaInfo* info = (MediaInfo *)pCallEvent->context;
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d mediacount %d, type 1 %d type 2 %d\n",
                                                 pCallEvent->callID, pData->accountid, info->nCount, info->media[0].codecType, info->media[1].codecType);
#else
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d \n", pCallEvent->callID, pData->accountid);
#endif
                                  }
                                  break;
                            }
#ifdef WITH_P2P
                            case EVENT_DATA:
                            {
                                  DataEvent *pDataEvent = &(event->body.dataEvent);
                                  allcount += 1;
                                  if (pData->timecount == 0) {
                                         pData->timecount = pDataEvent->pts;
                                  }
                                  else {

                                         if (allcount % 10 == 0) {
                                                 DBG_LOG("*******size %d** timestamp %ld last timestamp %ld callid %d count %d \n",
                                                      pDataEvent->size, pDataEvent->pts, pData->timecount, pDataEvent->callID, allcount);
                                         }
                                         pData->timecount = pDataEvent->pts;
                                  }
                                  break;
                            }
#endif
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
    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;

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
    setPjLogLevel(2);
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
#ifdef WITH_P2P
            pData->accountid = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost);
#else
            pData->accountid = Register( pData->id, pData->password, pData->sigHost, pData->imHost);
#endif
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

