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

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define HOST "123.59.204.198"
#define INVALID_SERVER "192,.168.1.239"

int RegisterTestSuitCallback( TestSuit *this );
int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
int RegisterTestSuitCallback2( TestSuit *this );
void *UA1_EventLoopThread( void *arg );
void *UA2_EventLoopThread( void *arg );
void *UA3_EventLoopThread( void *arg );
void *UA4_EventLoopThread( void *arg );
void *UA5_EventLoopThread( void *arg );

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
        { "valid_account1", CALL_STATUS_REGISTERED, UA1_EventLoopThread },
        { "1011", "1011", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, UA2_EventLoopThread },
        { "1003", "1003", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTER_FAIL, UA3_EventLoopThread },
        { "1003", "1004", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTER_FAIL, UA4_EventLoopThread },
        { "0000", "0000", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTER_FAIL, UA5_EventLoopThread },
        { "1003", "1003", INVALID_SERVER, INVALID_SERVER, INVALID_SERVER, 10, 0 }
    }
};

TestSuit gRegisterTestSuit =
{
    "Register",
    RegisterTestSuitCallback,
    RegisterTestSuitInit,
    RegisterTestSuitGetTestCase,
    (void*)&gRegisterTestCases,
    1
};

TestSuit gRegisterTestSuit2 =
{
    "Register2",
    RegisterTestSuitCallback2,
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

    UT_STR( pData->id );
    UT_STR( pData->password );
    UT_STR( pData->sigHost );
    UT_LOG("Register in\n");
    sts = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost);
    UT_LOG("Register out %x %x\n", sts, pTestCases->father.expact);
    TEST_GT( sts, pTestCases->father.expact );
    int nCallId1 = -1;
    ErrorID id;
    //sleep(10);
    int count = 0;
    while (count != 10) {
            UT_LOG("MakeCall in\n");
            id = MakeCall(sts, pData->id, "<sip:1010@123.59.204.198;transport=tcp>", &nCallId1);
            if (RET_OK != id) {
                    fprintf(stderr, "call error %d \n", id);
                    sleep(1);
                    ++ count;
                    continue;
            }
            sleep(1500000);
            UT_LOG("HangupCall in\n");
            int ret = HangupCall(sts, nCallId1);
            TEST_EQUAL(ret, RET_OK);
            UT_LOG("HangupCall out\n");
            sleep(1);
            ++ count;
    }
    UnRegister(sts);
}

int RegisterTestSuitCallback2( TestSuit *this )
{
    RegisterTestCase *pTestCases = NULL;
    RegisterData *pData = NULL;
    RegisterTestCase *pTestCase = NULL;
    Media media;
    int i = 0;
    int ret = 0;
    static ErrorID sts = 0;
    EventManger *pEventManager = &this->pManager->eventManager;

    UT_LINE();
    if ( !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    if ( pData->init ) {
        sts = InitSDK( &media, 1 );
        if ( RET_OK != sts ) {
            UT_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }

    UT_STR( pData->id );
    UT_STR( pData->password );
    UT_STR( pData->sigHost );

    sts = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }
    UT_VAL( sts );
    this->pManager->AddPrivateData( &sts );
    if ( pTestCase->father.threadEntry )
        this->pManager->ThreadRegister( pTestCase->father.threadEntry );

    if ( pEventManager->WaitForEvent ) {
        UT_VAL( pTestCase->father.expact );
        ret = pEventManager->WaitForEvent( pTestCase->father.expact, pData->timeOut );
        if ( ret == ERROR_TIMEOUT ) {
            UT_ERROR("ERROR_TIMEOUT\n");
            return TEST_FAIL;
        } else if ( ret == ERROR_INVAL ) {
            UT_ERROR("ERROR_INVAL\n");
            return TEST_FAIL;
        }
        UT_VAL( ret );
        return TEST_PASS;
    }

    return TEST_FAIL;
}

void *UA1_EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

void *UA2_EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

void *UA3_EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

void *UA4_EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

void *UA5_EventLoopThread( void *arg )
{
    TestSuitManager *pManager = (TestSuitManager *)arg;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    if ( !pManager->data ) {
        UT_ERROR("check data error\n");
        return NULL;
    }

    id = *(AccountID *)pManager->data;
    UT_VAL( id );


    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->NotifyAllEvent( pCallEvent->status );
            }
        }
    }

    return NULL;
}

int InitAllTestSuit()
{
    AddTestSuit( &gRegisterTestSuit );
    //AddTestSuit( &gRegisterTestSuit2 );

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

