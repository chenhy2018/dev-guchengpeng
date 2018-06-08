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
#include "test.h"
#include <unistd.h> 

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define HOST "123.59.204.198"

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
} RegisterData;

typedef struct {
    TestCase father;
    RegisterData data;
} RegisterTestCase;

RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", 0 },
        { "1011", "1011", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1011", "1007", HOST, HOST, HOST, 100, 0 }
    }
};

TestSuit gRegisterTestSuit =
{
    "Register",
    RegisterTestSuitCallback,
    RegisterTestSuitInit,
    RegisterTestSuitGetTestCase,
    (void*)&gRegisterTestCases,
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

int RegisterTestSuitCallback( TestSuit *this )
{
    RegisterTestCase *pTestCases = NULL;
    RegisterData *pData = NULL;
    Media media[2];
    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    int i = 0;
    ErrorID sts = 0;

    if ( !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *) this->testCases;
    DBG_LOG("this->index = %d\n", this->index );
    pData = &pTestCases[this->index].data;

    if ( pData->init ) {
        DBG_LOG("InitSDK");
        sts = InitSDK( &media[0], 2 );
        if ( RET_OK != sts ) {
            DBG_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }

    DBG_STR( pData->id );
    DBG_STR( pData->password );
    DBG_STR( pData->sigHost );
    DBG_LOG("Register in\n");
    sts = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost);
    DBG_LOG("Register out %x %x\n", sts, pTestCases->father.expact);
    TEST_GT( sts, pTestCases->father.expact );
    int nCallId1 = -1;
    ErrorID id;
    //sleep(10);
    int count = 0;
    while (count != 10) {
            DBG_LOG("MakeCall in\n");
            id = MakeCall(sts, pData->id, "<sip:1004@123.59.204.198>", &nCallId1);
            if (RET_OK != id) {
                    fprintf(stderr, "call error %d \n", id);
                    sleep(1);
                    ++ count;
                    continue;
            }
            sleep(150000);
            DBG_LOG("HangupCall in\n");
            int ret = HangupCall(sts, nCallId1);
            TEST_EQUAL(ret, RET_OK);
            DBG_LOG("HangupCall out\n");
            sleep(1);
            ++ count;
    }
    UnRegister(sts);
}

void *EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    DBG_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        DBG_ERROR("check param error\n");
        return NULL;
    }

    if ( pEventManager->WaitForEvent ) {
        int ret = pEventManager->WaitForEvent( WAIT_FOR_ACCOUNTID, 5 );
        if ( STS_OK != ret ) {
            DBG_ERROR("wait for event WAIT_FOR_ACCOUNTID error, ret = %d\n", ret );
            return NULL;
        } else {
            DBG_LOG("get WAIT_FOR_ACCOUNTID event ok\n");
        }
    } else {
        DBG_ERROR("WaitForEvent error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        DBG_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    DBG_VAL( id );


    for (;;) {
        DBG_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &event, 0 );
        DBG_VAL( type );
        if ( type == EVENT_CALL ) {
            DBG_LINE();
            pCallEvent = &event.body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                DBG_LINE();
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
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
    ThreadRegister( EventLoopThread );
    InitAllTestSuit();
    RunAllTestSuits();

    return 0;
}

