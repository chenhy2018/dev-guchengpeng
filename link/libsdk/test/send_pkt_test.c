// Last Update:2018-06-21 12:40:24
/**
 * @file send_pkt_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-14
 */

#include <string.h>
#include "dbg.h"
#include "unit_test.h"
#include "sdk_interface.h"
#include "send_pkt_test.h"

typedef struct {
    char *pData;
    unsigned char video;
    unsigned char audio;
    unsigned char data;
} SendPacketData;

typedef struct {
    TestCase father;
    SendPacketData data;
} SendPacketTestCase;

typedef struct {
    char *id;
    char *expact;
    int callId;
    int res;
} Account;

static Account gAccountList[MAX_CALLS] =
{
    { "1001", "12345678" },
    { "1002", "abcdefgh" },
    { "1003", "11223344" },
    { "1004", "98765432" },
    { "1005", "hijklmno" },
    { "1006", "opqrstuw" },
    { "1007", "vwxyz123" },
    { "1008", "aabbccdd" },
    { "1009", "eeffgghh" },
    { "1010", "iijjkkll" },
    { "1011", "mmnnoopp" },
    { "1012", "qqrrsstt" },
    { "1013", "uuvvwwxx" },
    { "1014", "55667788" },
    { "1015", "99887766" },
    { "1016", "55443322" }
};


int SendPacketTestSuitCallback( TestSuit *this );
int SendPacketTestSuitGetTestCase( TestSuit *this, TestCase **testCase );
int SendPacketTestCaseMultiCallback( TestSuit *this );
void * SendPacketMultiCallTestEventLoop( void *arg );

SendPacketTestCase gSendPacketTestCases[] =
{
    {
        { "normal video packet", RET_OK },
        { "1234567890abc", 1 }
    },
    {
        { "normal audio packet", RET_OK },
        { "1234567890abc", 0, 1 }
    },
    {
        { "normal data packet", RET_OK },
        { "1234567890abc", 0, 1 }
    },
    {
        { "multi call", RET_OK, SendPacketMultiCallTestEventLoop, SendPacketTestCaseMultiCallback },
        { "1234567890abc", 0, 1 }
    },
};

TestSuit gSendPacketTestSuit = 
{
    "SendPacket()",
    SendPacketTestSuitCallback,
    NULL,
    SendPacketTestSuitGetTestCase,
    (void*)&gSendPacketTestCases,
    0,
    NULL,
    ARRSZ(gSendPacketTestCases)
};

int SendPacketTestSuitCallback( TestSuit *this )
{
    SendPacketTestCase *pTestCases = NULL;
    SendPacketData *pData = NULL;
    SendPacketTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    AccountID id = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    DataEvent *pDataEvent = NULL;
    unsigned char established = 0;
    unsigned char retry = 0;

    UT_LINE();
    if ( !this ) {
        return TEST_FAIL;
    }

    pTestCases = (SendPacketTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    id = Register( "1003", "1003", "123.59.204.198", "123.59.204.198", "123.59.204.198" );
    if ( sts >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }

    UT_VAL( id );
    for (;;) {
        sts = PollEvent( id, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return TEST_FAIL;
        }

        if ( RET_RETRY == sts ) {
            if ( established ) {
                retry++;
                if ( retry == 5 ) {
                    UT_ERROR("waiting data time out\n");
                    return TEST_FAIL;
                }
            } else {
                retry ++;
                if ( retry == 5 ) {
                    UT_ERROR("waiting call established time out\n");
                    return TEST_FAIL;
                }
            }
        }
        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("get event EVENT_CALL\n");
            UT_LOG("pEvent = %p\n", pEvent );
            if ( pEvent ) {
                pCallEvent = &pEvent->body.callEvent;
                UT_VAL( pCallEvent->status );
                char *callSts = DbgCallStatusGetStr( pCallEvent->status );
                UT_LOG("status : %s\n", callSts );
                char * pFromAccount = (char*) pCallEvent->context;
                if ( pFromAccount )
                    UT_STR( pFromAccount );
                if ( CALL_STATUS_INCOMING == pCallEvent->status ) {
                    UT_VAL( pCallEvent->callID );
                    ret = AnswerCall( id, pCallEvent->callID );
                    if ( ret != RET_OK ) {
                        return TEST_FAIL;
                    }
                } else if ( CALL_STATUS_ESTABLISHED == pCallEvent->status ) {
                    established = 1;
                    retry = 0;
                    UT_LOG("CALL_STATUS_ESTABLISHED\n");
                } else {
                    UT_ERROR("wrong event, pCallEvent->status = %d\n", pCallEvent->status );
                    return TEST_FAIL;
                }
            }
            break;
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            if ( pEvent ) {
                pDataEvent = &pEvent->body.dataEvent;
                UT_VAL( pDataEvent->stream );
                UT_VAL( pDataEvent->size );
                switch ( pDataEvent->stream ) {
                case STREAM_VIDEO:
                    if ( pData->video ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                case STREAM_AUDIO:
                    if ( pData->audio ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                case STREAM_DATA:
                    if ( pData->data ) {
                        if ( memcmp( pData->pData, pDataEvent->data, pDataEvent->size) == 0 ) {
                            return TEST_PASS;
                        } else {
                            UT_STR( (char *) pDataEvent->data );
                            return TEST_FAIL;
                        }
                    } else {
                        UT_ERROR("fail\n");
                        return TEST_FAIL;
                    }
                    break;
                default:
                    UT_ERROR("unknow stream type, stream = %d\n", pDataEvent->stream );
                    break;
                }
            }
            break;
        case EVENT_MESSAGE:
            UT_LOG("get event EVENT_MESSAGE\n");
            break;
        default:
            UT_LOG("unknow event, type = %d\n", type );
            break;
        }
    }

    return TEST_FAIL;
}

int SendPacketTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    SendPacketTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (SendPacketTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}


int SendPacketTestCaseMultiCallback( TestSuit *this )
{
    SendPacketTestCase *pTestCases = NULL;
    SendPacketData *pData = NULL;
    SendPacketTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0;
    static ErrorID sts = 0;
    AccountID id = 0;
    EventManger *pEventManager = &this->pManager->eventManager;
    int event = 0;

    if ( !this ) {
        return TEST_FAIL;
    }

    pTestCases = (SendPacketTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pTestCase = &pTestCases[this->index];
    pData = &pTestCase->data;

    id = Register( "1003", "1003", HOST, HOST, HOST );
    if ( id >= RET_MEM_ERROR ) {
        DBG_ERROR("sts = %d\n", sts );
        return TEST_FAIL;
    }

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
            for ( i=0; i<MAX_CALLS; i++ ) {
                sts = MakeCall( sts,  gAccountList[i].id, HOST, &gAccountList[i].callId );
                if ( sts != RET_OK ) {
                    UT_ERROR( "make call error, sts = %d\n", sts );
                    return TEST_FAIL;
                }
            }

            if ( pEventManager->WaitForEvent ) {
                ret = pEventManager->WaitForEvent( EVENT_16_CHANNEL_OK, 20, NULL, this, (void *)&event );
                if ( ret == ERROR_TIMEOUT ) {
                    UT_ERROR("ERROR_TIMEOUT\n");
                    return TEST_FAIL;
                } else if ( ret == ERROR_INVAL ) {
                    UT_ERROR("ERROR_INVAL\n");
                    return TEST_FAIL;
                } else {
                    if ( event == EVENT_16_CHANNEL_CHECK_OK ) {
                        return TEST_PASS;
                    } else {
                        return TEST_FAIL;
                    }
                }
            }
        }
    }

    return TEST_FAIL;
}

int CallEventHnadle( AccountID nAccountId, Event *pEvent )
{            
    CallEvent *pCallEvent = NULL;
    int ret = 0;

    if ( pEvent ) {
        pCallEvent = &pEvent->body.callEvent;
        UT_VAL( pCallEvent->status );
        char *callSts = DbgCallStatusGetStr( pCallEvent->status );
        UT_LOG("status : %s\n", callSts );
        if ( CALL_STATUS_INCOMING == pCallEvent->status ) {
            char* pFromAccount = (char*)pCallEvent->context;
            if ( pFromAccount )
                UT_STR( pFromAccount );
            UT_VAL( pCallEvent->callID );
            ret = AnswerCall( nAccountId, pCallEvent->callID );
            if ( ret != RET_OK ) {
                return -1;
            }
        } else if ( CALL_STATUS_ESTABLISHED == pCallEvent->status ) {
            UT_LOG("CALL_STATUS_ESTABLISHED\n");
            UT_VAL( pCallEvent->callID );
        } else {
            UT_ERROR("wrong event, pCallEvent->status = %d\n", pCallEvent->status );
            return -1;
        }
    }

    return 0;
}

int CheckDataEvent( int _nCallId, char *data, int size )
{
    int i;

    for ( i=0; i<ARRSZ(gAccountList); i++ ) {
        if ( _nCallId == gAccountList[i].callId ) {
            if ( memcmp( data, gAccountList[i].expact, size ) == 0 ) {
                gAccountList[i].res = 1;
                return 0;
            } else {
                return DATA_CHECK_ERROR;
            }
        }
    }

    return CALL_NOT_FOUND;
}

int CheckAllAccountStatus()
{
    int i;

    for ( i=0; i<ARRSZ(gAccountList); i++ ) {
        if ( !gAccountList[i].res ) {
            return EVENT_NOT_COMPLETE;
        }
    }

    return 0;
}

int DataEventHandle( Event *pEvent )
{
    DataEvent *pDataEvent = NULL;

    if ( pEvent ) {
        pDataEvent = &pEvent->body.dataEvent;
        UT_VAL( pDataEvent->stream );
        UT_VAL( pDataEvent->size );
        switch ( pDataEvent->stream ) {
        case STREAM_VIDEO:
            UT_LOG("STREAM_VIDEO\n");
            if ( 0 != CheckDataEvent( pDataEvent->callID, pDataEvent->data, pDataEvent->size) ) {
                return -1;
            }
            break;
        case STREAM_AUDIO:
            UT_LOG("STREAM_AUDIO\n");
            break;
        case STREAM_DATA:
            UT_LOG("STREAM_DATA\n");
            break;
        default:
            UT_ERROR("unknow stream type, stream = %d\n", pDataEvent->stream );
            break;
        }
    }

    return 0;
}

void * SendPacketMultiCallTestEventLoop( void *arg )
{
    SendPacketTestCase *pTestCases = NULL;
    SendPacketData *pData = NULL;
    SendPacketTestCase *pTestCase = NULL;
    int i = 0;
    int ret = 0, callId = 0;
    static ErrorID sts = 0;
    AccountID id = 0;
    EventType type = 0;
    Event *pEvent = NULL;
    CallEvent *pCallEvent = NULL;
    DataEvent *pDataEvent = NULL;
    unsigned char established = 0;
    unsigned char retry = 0;
    TestSuit *pTestSuit = ( TestSuit *) arg;
    EventManger *pEventManager = &pTestSuit->pManager->eventManager;
    int *pVal = NULL;

    UT_LINE();
    if ( !arg ) {
        return NULL;
    }

    pTestCases = (SendPacketTestCase *) pTestSuit->testCases;
    UT_LOG("this->index = %d\n", pTestSuit->index );
    pTestCase = &pTestCases[pTestSuit->index];
    pData = &pTestCase->data;

    UT_VAL( id );
    for (;;) {
        sts = PollEvent( id, &type, &pEvent, 5 );
        if ( sts >= RET_MEM_ERROR ) {
            UT_ERROR("PollEvent error, sts = %d\n", sts );
            return NULL;
        }

        if ( RET_RETRY == sts ) {
            if ( 0 != CheckAllAccountStatus() ) {
                UT_VAL( retry );
                retry++;
                if ( retry == 20 ) {
                    UT_ERROR("waiting data time out\n");
                    pEventManager->GetEventDataAddr( pTestSuit, (void **)&pVal );
                    if ( pVal ) {
                        *pVal = EVENT_16_CHANNEL_CHECK_FAIL;
                    }
                    pEventManager->NotifyAllEvent( EVENT_16_CHANNEL_OK, pVal );
                    return NULL;
                }
            }
        }

        UT_VAL( type );
        switch( type ) {
        case EVENT_CALL:
            UT_LOG("get event EVENT_CALL\n");
            CallEventHnadle( id, pEvent ); 
            break;
        case EVENT_DATA:
            UT_LOG("get event EVENT_DATA\n");
            if ( 0 != DataEventHandle( pEvent ) ) {
                if ( pVal ) {
                    *pVal = EVENT_16_CHANNEL_CHECK_FAIL;
                }
                pEventManager->NotifyAllEvent( EVENT_16_CHANNEL_OK, pVal );
                return NULL;
            }
            if ( 0 == CheckAllAccountStatus() ) {
                pEventManager->GetEventDataAddr( pTestSuit, (void **)&pVal );
                if ( pVal ) {
                    *pVal = EVENT_16_CHANNEL_CHECK_OK;
                }
                pEventManager->NotifyAllEvent( EVENT_16_CHANNEL_OK, pVal );
            }
            break;
        case EVENT_MESSAGE:
            UT_LOG("get event EVENT_MESSAGE\n");
            break;
        default:
            UT_LOG("unknow event, type = %d\n", type );
            break;
        }
    }

    return NULL;
}


