// Last Update:2018-06-08 20:13:25
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
#include "test.h"
#include <unistd.h> 
#include <stdlib.h>
#include <pthread.h>

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define HOST "39.107.247.14"
#define INVALID_SERVER "192.168.1.239"

#define MAX_COUNT 9
#define CHENGPENG 1
//#define MENGKE 1
int RegisterTestSuitCallback( TestSuit *this );
int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
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
    int sendflag;
    int64_t timecount;
} RegisterData;

typedef struct {
    TestCase father;
    RegisterData data;
} RegisterTestCase;

#ifdef CHENGPENG
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2041", "LDb1wHu9", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2042", "qIwCXfUp", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2043", "IZYkxp5O", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2044", "WZx6T3Er", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2045", "vdK3TsK0", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2046", "I1gj5Wef", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2047", "rx7HOnDT", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "2048", "VidAZJIM", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2049", "1NR6D190", HOST, HOST, HOST, 100, 1 }
    },
};
#endif

#ifdef MENGKE
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_REGISTERED, UA1_EventLoopThread },
        { "1711", "aUSEOnOy", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, UA2_EventLoopThread },
        { "1712", "Q0EEBOEc", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTER_FAIL, UA3_EventLoopThread },
        { "1713", "IeFxv0sP", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTER_FAIL, UA4_EventLoopThread },
        { "1714", "9ZykZwsJ", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTER_FAIL, UA5_EventLoopThread },
        { "1715", "AzkaVAo0", INVALID_SERVER, INVALID_SERVER, INVALID_SERVER, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1716", "5XPM9DUv", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1717", "W9DaI77R", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1718", "4JoChsXl", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1719", "NVWrpASp", HOST, HOST, HOST, 100, 1 }
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
    1
};

int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    this->total = ARRSZ(gRegisterTestCases);
    this->index = 0;
    this->pManager = _pManager;

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
    ErrorID id = 0;
    EventType type;
    int timecount = 0;
    int sendflag = 0;
    char data1[100] = {1};
    DBG_LOG("send pack *****%d call id %d\n", pData->accountid, pData->callid);
    while (1) {     
                    if (sendflag) {
                            if (timecount > 500) {
                                    DBG_LOG("hangcall ******************\n");
                                    HangupCall(pData->accountid, pData->callid);
                                    sendflag = 0;
                                    timecount = 0;
                                    continue;
                            }
                            id = SendPacket(pData->accountid, pData->callid, STREAM_AUDIO, data1, 100, timecount);
                            if (id > 0) DBG_LOG("send pack *****%d\n", id);
                            timecount += 1;
                            usleep(10000);
                            continue;
                    }
                    id = PollEvent(pData->accountid, &type, &event, 10);
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
                                  if (pCallEvent->status == CALL_STATUS_ERROR || pCallEvent->status == CALL_STATUS_HANGUP) {
                                        DBG_LOG("makecall **************************** reason %d\n", pCallEvent->reasonCode);
                                        do {
                                                sleep(1);
                                                id = MakeCall(pData->accountid, "2040", HOST, &pData->callid);
                                        } while (id != RET_OK);
                                  }
                                  if (pCallEvent->status == CALL_STATUS_ESTABLISHED) {
                                        MediaInfo* info = (MediaInfo *)pCallEvent->context;
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d mediacount %d, type 1 %d type 2 %d\n",
                                                 pCallEvent->callID, pData->accountid, info->nCount, info->media[0].codecType, info->media[1].codecType);
                                        sendflag = 1;
                                  }

                                  break;
                            }
                            case EVENT_DATA:
                            {     
                                  DataEvent *pDataEvent = &(event->body.dataEvent);
                                  //DBG_LOG("Data size %d call id %d call account id %d timestamp %lld \n", pDataEvent->size, pDataEvent->callID, pData->accountid, pDataEvent->pts);
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
    int all_count = 0;
    pTestCases = (RegisterTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pData = &pTestCases[this->index].data;

    if ( pData->init ) {
        UT_LOG("InitSDK");
        sts = InitSDK( &media[0], 2 );
        if ( RET_OK != sts ) {
            UT_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }
    setPjLogLevel(2);
    UT_STR( pData->id );
    for (int count = 0; count < MAX_COUNT; ++ count) {
            pData = &pTestCases[count].data;
            UT_STR(pData->password);
            UT_STR(pData->sigHost);
            UT_LOG("Register in\n");
            pData->accountid = Register(pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost);
            UT_LOG("Register out %x %x\n", pData->accountid, pTestCases->father.expact);
            int nCallId1 = -1;
    }
    sleep(10);

        pthread_t t_1;
        pthread_attr_t attr_1;
        pthread_attr_init(&attr_1);
        pthread_attr_setdetachstate(&attr_1, PTHREAD_CREATE_DETACHED);

    Event* event = (Event*) malloc(sizeof(Event));
    ErrorID id;
    for (int count = 0; count < MAX_COUNT; ++ count) {
           pData = &pTestCases[count].data;
           pData->sendflag = 0;
           pData->timecount = 0;
           UT_LOG("MakeCall in accountid %d\n", pData->accountid);
           pData->callid = 0;
           id = MakeCall(pData->accountid, "2040", HOST, &pData->callid);
           if (RET_OK != id) {
                    fprintf(stderr, "call error %d \n", id);
                     continue;
           }
           UT_LOG("MakeCall in callidid %d\n", pData->callid);
           pthread_create(&t_1, &attr_1, Mthread1, pData);
    }
    while(1) { sleep(10); }
}

int InitAllTestSuit()
{
    AddTestSuit( &gRegisterTestSuit );
    return 0;
}

int main()
{
    UT_LOG("+++++ enter main...\n");
    TestSuitManagerInit();
    InitAllTestSuit();
    RunAllTestSuits();
    return 0;
}

