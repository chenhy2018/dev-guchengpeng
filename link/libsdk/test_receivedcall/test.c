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
#include "sdk_interface.h"
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
#define MAX_COUNT 5
#define HOST "123.59.204.198"
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", 0 },
        { "1010", "1010", HOST, HOST, HOST, 100, 1 }
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
                            case EVENT_MEDIA:
                            {
                                 MediaEvent *pMedia = &(event->body.mediaEvent);
                                 DBG_LOG("Callid %d ncount %d type 1 %d type 2 %d\n", pMedia->callID, pMedia->nCount, pMedia->media[0].codecType, pMedia->media[1].codecType);
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
    Event* event= (Event*) malloc(sizeof(Event));
    while (1) {
            for (int count = 0; count < MAX_COUNT; ++count) {
                    pData = &pTestCases[count].data;
                    id = PollEvent(pData->accountid, &type, &event, 1);
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
                                      AnswerCall(pData->accountid, pCallEvent->callID);
                                      DBG_LOG("AnswerCall end *****************\n");
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
                            case EVENT_MEDIA:
                            {
                                 MediaEvent *pMedia = &(event->body.mediaEvent);
                                 DBG_LOG("Callid %d ncount %d type 1 %d type 2 %d\n", pMedia->callID, pMedia->nCount, pMedia->media[0].codecType, pMedia->media[1].codecType);
                                 break;
                            }
                    }
           }
    }
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

