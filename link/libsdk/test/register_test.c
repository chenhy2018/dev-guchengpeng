// Last Update:2018-06-20 19:26:26
/**
 * @file register_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-12
 */
#include "sdk_interface.h"
#include "dbg.h"
#include "unit_test.h"
#include <unistd.h>

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
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
        { "valid account1", CALL_STATUS_REGISTERED },
        { "1002", "1002", HOST, HOST, HOST, 10, 0 }
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
    }
#if 0
    {
        { "invalid sip register server", CALL_STATUS_REGISTER_FAIL },
        { "1003", "1003", INVALID_SERVER, INVALID_SERVER, INVALID_SERVER, 10, 0 }
    }
#endif
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
    Media media[2];
    ErrorID sts = 0;

    this->total = ARRSZ(gRegisterTestCases);
    this->index = 0;
    this->pManager = _pManager;

    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
    sts = InitSDK( media, 2 );
    if ( RET_OK != sts ) {
        UT_ERROR("sdk init error\n");
        return -1;
    }

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

int RegisterTestEventCallBack( TestCase *pTestCase, void *data )
{
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
    int event = 0;

    /*UT_LINE();*/
    if ( !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *) this->testCases;
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    if ( pData->init ) {
        sts = InitSDK( &media, 1 );
        if ( RET_OK != sts ) {
            UT_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }

    sts = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }
    pTestCase->father.data = (void *)sts;
    this->pManager->startThread( this );

    if ( pEventManager->WaitForEvent ) {
        ret = pEventManager->WaitForEvent( pTestCase->father.expact, pData->timeOut, RegisterTestEventCallBack, this, NULL );
        if ( ret == ERROR_TIMEOUT ) {
            UT_ERROR("ERROR_TIMEOUT\n");
            return TEST_FAIL;
        } else if ( ret == ERROR_INVAL ) {
            UT_ERROR("ERROR_INVAL\n");
            return TEST_FAIL;
        }
        /*UT_VAL( ret );*/
        //UnRegister(sts);
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

    /*UT_NOTICE("EventLoopThread enter ..., count = %d\n", count );*/
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }

    count++;

    pTestCases = ( RegisterTestCase *) pTestSuit->testCases;
    pTestCase = &pTestCases[pTestSuit->index];
    id = (AccountID)(long) pTestCase->father.data;

    while ( pTestCase->father.running ) {
        ret = PollEvent( id, &type, &pEvent, 0);
        if ( ret >= RET_MEM_ERROR ) {
            UT_ERROR("ret = %d\n", ret );
            return NULL;
        }
        if ( type == EVENT_CALL ) {
            DBG_LOG("EVENT_CALL\n");
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                if ( pManager->eventManager.NotifyAllEvent ) {
                    char * pStr = DbgCallStatusGetStr( pCallEvent->status );
                    UT_STR( pStr );
                    pManager->eventManager.NotifyAllEvent( pCallEvent->status, NULL );
                }
            }
        }
    }

    /*UT_NOTICE("thread %ld exit ...., count = %d\n", pTestSuit->tid, count );*/

    return NULL;
}

