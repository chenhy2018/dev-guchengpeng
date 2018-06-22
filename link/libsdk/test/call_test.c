// Last Update:2018-06-21 17:16:08
/**
 * @file call_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-11
 */

#include "dbg.h"
#include "unit_test.h"
#include "sdk_interface.h"

typedef struct {
    char *id;
    char *host;
    unsigned char init;
    int timeOut;
    unsigned char media;
} CallData;

typedef struct {
    TestCase father;
    CallData data;
} MakeCallTestCase;


void *MakeCallEventLoopThread( void *arg );
int MakeCallTestSuitCallback( TestSuit *this );
int MakeCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
void *CalleeThread( void *arg );
int MakeCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager );

MakeCallTestCase gMakeCallTestCases[] =
{
    {
        { "valid account1", CALL_STATUS_ESTABLISHED },
        { "1003", "123.59.204.198", 1, 10, 1 }
    },
    {
        { "valid account2", CALL_STATUS_ESTABLISHED },
        { "1002", "123.59.204.198", 1, 10, 1 }
    },
    {
        { "invalid account1", CALL_STATUS_TIMEOUT },
        { "0000", "123.59.204.198", 1, 10, 1 }
    },
    {
        { "invalid account2", CALL_STATUS_TIMEOUT },
        { "abcd", "123.59.204.198", 1, 10, 1 }
    },
    {
        { "invalid sip server", CALL_STATUS_TIMEOUT },
        { "0000", "123.59.204.198", 1, 10, 1 }
    },
    {
        { "invalid sip server2", CALL_STATUS_TIMEOUT },
        { "0000", "www.google.com", 1, 10, 1 }
    },
    {
        { "illegal_account_id", CALL_STATUS_TIMEOUT },
        { NULL, "123.59.204.198", 1, 10, 1 }
    },
    {
        { "illegal_server_address", CALL_STATUS_TIMEOUT },
        { "1002", NULL, 1, 10, 1 }
    },
    {
        { "illegal_account_id_and_server_address", CALL_STATUS_TIMEOUT },
        { NULL, NULL, 1, 10, 1 }
    },
};

TestSuit gMakeCallTestSuit =
{
    "MakeCall()",
    MakeCallTestSuitCallback,
    MakeCallTestSuitInit,
    MakeCallTestSuitGetTestCase,
    (void*)&gMakeCallTestCases,
    1,
    MakeCallEventLoopThread,
    ARRSZ(gMakeCallTestCases) 
};

void *MakeCallEventLoopThread( void *arg )
{
    TestSuit *pTestSuit = (TestSuit *)arg;
    TestSuitManager *pManager = pTestSuit->pManager;
    ErrorID ret = 0;
    EventType type = 0;
    AccountID id = 0;
    Event event, *pEvent;
    CallEvent *pCallEvent;
    EventManger *pEventManager = &pManager->eventManager;
    MakeCallTestCase *pTestCase = NULL;
    MakeCallTestCase *pTestCases = NULL;

    UT_LOG("EventLoopThread enter ...\n");
    if ( !pManager ) {
        UT_ERROR("check param error\n");
        return NULL;
    }
    pTestCases = ( MakeCallTestCase *) pTestSuit->testCases;
    pTestCase = &pTestCases[pTestSuit->index];
    id = (AccountID)(long) pTestCase->father.data;

    while ( pTestCase->father.running ) {
        ret = PollEvent( id, &type, &pEvent, 1000 );
        if ( ret >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, ret = %d\n", ret );
            return NULL;
        } else if ( ret == RET_RETRY ) {
            continue;
        }

        if ( type == EVENT_CALL ) {
            UT_LOG("get event EVENT_CALL\n");
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                UT_VAL( pCallEvent->callID );
                if ( pManager->eventManager.NotifyAllEvent ) {
                    char *str = DbgCallStatusGetStr( pCallEvent->status );
                    UT_STR( str );
                    pManager->eventManager.NotifyAllEvent( pCallEvent->status, NULL );
                }
            }
        }
    }

    return NULL;
}

int MakeCallTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    MakeCallTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (MakeCallTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}

int MakeCallTestSuitCallback( TestSuit *this )
{
    MakeCallTestCase *pTestCases = NULL;
    CallData *pData = NULL;
    MakeCallTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    EventManger *pEventManager = &this->pManager->eventManager;

    if ( !this ) {
        return -1;
    }

    pTestCases = (MakeCallTestCase *) this->testCases;
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    sts = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        char *str = DbgSdkRetGetStr( sts );
        UT_ERROR("str = %s\n", str );
        return TEST_FAIL;
    }
    pTestCase->father.data = (void *)sts;
    this->pManager->startThread( this );

    if ( pEventManager->WaitForEvent ) {
        UT_VAL( pTestCase->father.expact );
        ret = pEventManager->WaitForEvent( CALL_STATUS_REGISTERED, 10, NULL, this, NULL );
        if ( ret == ERROR_TIMEOUT ) {
            UT_ERROR("ERROR_TIMEOUT\n");
            return TEST_FAIL;
        } else if ( ret == ERROR_INVAL ) {
            UT_ERROR("ERROR_INVAL\n");
            return TEST_FAIL;
        } else {
            UT_VAL( ret );
            sts = MakeCall( sts,  pData->id, pData->host, &callId );
            UT_VAL( sts );
            if ( sts != RET_OK ) {
                UT_ERROR( "make call error, sts = %d\n", sts );
                return TEST_FAIL;
            }
            UT_VAL( callId );
            ret = pEventManager->WaitForEvent( pTestCase->father.expact, pData->timeOut, NULL, this, NULL  );
            if ( ret != STS_OK ) {
                if ( ret == ERROR_TIMEOUT ) {
                    UT_ERROR(" MakeCall fail, ret = ERROR_TIMEOUT\n" );
                } else if ( ret == ERROR_INVAL ) {
                    UT_ERROR(" MakeCall fail, ret = ERROR_INVAL\n" );
                } else {
                    UT_ERROR(" MakeCall fail, ret = unknow errorf\n" );
                }
                return TEST_FAIL;
            } else {
                return TEST_PASS;
            }
        }
    }

    return TEST_FAIL;
}

int MakeCallTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    int ret = 0;
    pthread_t tid = 0;
    Media media[2];
    ErrorID sts = 0;

    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
    sts = InitSDK( media, 2 );
    if ( RET_OK != sts && RET_SDK_ALREADY_INITED != sts ) {
        UT_ERROR("sdk init error\n");
        return -1;
    }
    
    ret = pthread_create( &tid, NULL, CalleeThread, this );
    if ( ret != 0 ) {
        UT_ERROR("pthread_create() error, ret = %d\n", ret );
        return -1;
    }
    return 0;
}

void *CalleeThread( void *arg )
{
    ErrorID sts = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    Media media;
    AccountID accountID = 0;
    TestSuit *pTestSuit = (TestSuit *)arg;
    MakeCallTestCase *pTestCases = NULL;
    MakeCallTestCase *pTestCase = NULL;

    UT_LOG("CalleeThread() entry...\n");

    pTestCases = (MakeCallTestCase *) pTestSuit->testCases;
    pTestCase = &pTestCases[pTestSuit->index];

    sts = Register( "1005", "1005", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        UT_ERROR("[ callee ] Register error, sts = %d\n", sts );
        return NULL;
    }
    accountID = (AccountID)sts;

    while( pTestCase->father.running ) {
        sts = PollEvent( accountID, &type, &pEvent, 1000 );
        if ( sts >= RET_MEM_ERROR ) {
            char *str = DbgSdkRetGetStr( sts );
            UT_ERROR("ret = %s\n", str );
            return NULL;
        }
        if ( sts == RET_RETRY ) {
            continue;
        }
        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("[ callee ] get event EVENT_CALL\n");
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                char *callSts = DbgCallStatusGetStr( pCallEvent->status );
                UT_LOG("[ callee ] status : %s\n", callSts );
                UT_VAL( pCallEvent->callID );
                if ( pCallEvent->status == CALL_STATUS_INCOMING ) {
                    if ( pCallEvent->pFromAccount ) {
                        UT_STR( pCallEvent->pFromAccount );
                    }
                    ErrorID ret = AnswerCall( accountID, pCallEvent->callID );
                    if ( ret >= RET_MEM_ERROR ) {
                        UT_ERROR("AnswerCall() error, ret = %s\n", DbgSdkRetGetStr(ret) );
                        return NULL;
                    }
                }
            }
            break;
        case EVENT_DATA:
            UT_LOG("[ callee ] get event EVENT_DATA\n");
            break;
        case EVENT_MESSAGE:
            UT_LOG("[ callee ] get event EVENT_MESSAGE\n");
            break;
        case EVENT_MEDIA:
            UT_LOG("[ callee ] get event EVENT_MEDIA\n");
            break;
        default:
            UT_LOG("[ callee ] unknow event, type = %d\n", type );
            break;
        }
    }

    return NULL;
}

