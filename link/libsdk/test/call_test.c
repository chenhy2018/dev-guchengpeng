// Last Update:2018-06-21 12:41:25
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

MakeCallTestCase gMakeCallTestCases[] =
{
    {
        { "valid account1", CALL_STATUS_ESTABLISHED },
        { "1015", "123.59.204.198", 1, 10, 1 }
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
        { "invalid sip server", CALL_STATUS_TIMEOUT },
        { "0000", "123.59.204.198", 1, 10, 1 }
    },
};

TestSuit gMakeCallTestSuit =
{
    "MakeCall",
    MakeCallTestSuitCallback,
    NULL,
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

    for (;;) {
        UT_LOG("call PollEvent\n");
        ret = PollEvent( id, &type, &pEvent, 0 );
        UT_VAL( type );
        if ( type == EVENT_CALL ) {
            UT_LINE();
            pCallEvent = &pEvent->body.callEvent;
            if ( pManager->eventManager.NotifyAllEvent ) {
                UT_VAL( pCallEvent->status );
                pManager->eventManager.NotifyAllEvent( pCallEvent->status, NULL );
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

    UT_LINE();
    if ( !this ) {
        return -1;
    }

    pTestCases = (MakeCallTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    UT_STR( pData->id );

    sts = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }
    pTestCase->father.data = (void *)sts;
    UT_VAL( sts );
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

void *CalleeThread( void *arg )
{
    ErrorID sts = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    Media media;

    UT_LOG("CalleeThread() entry...\n");

    sts = Register( "1015", "1015", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        UT_ERROR("Register error, sts = %d\n", sts );
        return NULL;
    }

    for (;;) {
        sts = PollEvent( sts, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return NULL;
        }
        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("get event EVENT_CALL\n");
            pCallEvent = &pEvent->body.callEvent;
            char *callSts = DbgCallStatusGetStr( pCallEvent->status );
            UT_LOG("status : %s\n", callSts );
            UT_STR( pCallEvent->pFromAccount );
            break;
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            break;
        case EVENT_MESSAGE:
            UT_LOG("get event EVENT_MESSAGE\n");
            break;
        case EVENT_MEDIA:
            UT_LOG("get event EVENT_MEDIA\n");
            break;
        default:
            UT_LOG("unknow event, type = %d\n", type );
            break;
        }
    }

    return NULL;
}

