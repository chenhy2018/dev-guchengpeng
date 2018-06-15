// Last Update:2018-06-15 16:17:11
/**
 * @file register_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-12
 */
#include "sdk_interface.h"
#include "unit_test.h"

#define HOST "123.59.204.198"
#define INVALID_SERVER "192.168.1.239"

int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
int RegisterTestSuitCallback( TestSuit *this );
void *RegisterEventLoopThread( void *arg );

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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid account2", CALL_STATUS_REGISTERED },
        { "1003", "1003", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid account1", CALL_STATUS_REGISTER_FAIL },
        { "1003", "1004", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid account2", CALL_STATUS_REGISTER_FAIL },
        { "0000", "0000", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid sip register server", CALL_STATUS_REGISTER_FAIL },
        { "1003", "1003", INVALID_SERVER, INVALID_SERVER, INVALID_SERVER, 10, 0 }
    }
};

TestSuit gRegisterTestSuit =
{
    "Register()",
    RegisterTestSuitCallback,
    RegisterTestSuitInit,
    RegisterTestSuitGetTestCase,
    (void*)&gRegisterTestCases,
    1,
    RegisterEventLoopThread
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
    RegisterData *pData = NULL;
    RegisterTestCase *pTestCase = NULL;
    RegisterTestCase *pTestCases = NULL;
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
    pTestCase->father.data = (void *)sts;
    if ( pTestCase->father.threadEntry )
        this->pManager->startThread( this, pTestCase->father.threadEntry );
    else
        this->pManager->startThread( this, this->threadEntry );

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

void *RegisterEventLoopThread( void *arg )
{
    TestSuit *pTestSuit = ( TestSuit *) arg;
    TestSuitManager *pManager = pTestSuit->pManager;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;
    RegisterTestCase *pTestCase = NULL;
    RegisterTestCase *pTestCases = NULL;
    static int count = 0;

    UT_NOTICE("EventLoopThread enter ..., count = %d\n", count );
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    count++;

    pTestCases = ( RegisterTestCase *) pTestSuit->testCases;
    pTestCase = &pTestCases[pTestSuit->index];
    id = (AccountID)(long) pTestCase->father.data;

    UT_VAL( id );

    while ( pTestCase->father.running ) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 10000 );
        UT_VAL( ret );
        if ( ret >= RET_MEM_ERROR ) {
            UT_ERROR("ret = %d\n", ret );
            return NULL;
        }
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                if ( pManager->NotifyAllEvent ) {
                    UT_VAL( pCallEvent->status );
                    pManager->NotifyAllEvent( pCallEvent->status );
                }
            }
        }
    }

    UT_NOTICE("thread %ld exit ....\n", pTestSuit->tid );

    return NULL;
}

