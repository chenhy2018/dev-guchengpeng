// Last Update:2018-06-21 12:36:04
/**
 * @file unit_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
#include <pthread.h>
#include <sys/time.h>
#include <errno.h>
#include "unit_test.h"

TestSuitManager gTestSuitManager, *pTestSuitManager = &gTestSuitManager;

int AddTestSuit( TestSuit *pTestSuit )
{
    int i = 0;

    if ( !pTestSuit ) {
        return -1;
    }

    if ( pTestSuitManager->num >= TEST_SUIT_MAX ) {
        return -1;
    }

    pTestSuitManager->testSuits[pTestSuitManager->num++] = *pTestSuit;

    return 0;
}

int RunAllTestSuits()
{
    int i = 0, j=0;
    TestSuit *pTestSuit = NULL;
    int res = 0;
    TestCase *pTestCase = NULL;
    EventManger *pEventManager = &pTestSuitManager->eventManager;

    UT_VAL( pTestSuitManager->num );
    for ( i=0; i<pTestSuitManager->num; i++ ) {
        pTestSuit = &pTestSuitManager->testSuits[i];
        if ( !pTestSuit->enable ) {
            continue;
        }
        pTestSuit->pManager = pTestSuitManager;
        pTestSuit->index = 0;
        if ( pTestSuit->OnInit ) {
            pTestSuit->OnInit( pTestSuit, pTestSuitManager );
        }
        
        LOG("run the test suit : %s, %d test cases\n", pTestSuit->suitName, pTestSuit->total );
        if ( pTestSuit->GetTestCase ) {
            for ( j=0; j<pTestSuit->total; j++ ) {
                pTestSuit->index = j;
                pTestSuit->GetTestCase( pTestSuit, &pTestCase );
                if ( !pTestCase ) {
                    UT_ERROR("GetTestCase error\n");
                    return -1;
                }

                LOG("\n++++++ run test case [ ");
                LOG_BLUE("%s", pTestCase->caseName );
                LOG(" ]\n");
                if ( pTestSuit->TestCaseCb ) {
                    res = pTestSuit->TestCaseCb( pTestSuit );
                } else {
                    res = pTestCase->TestCaseCb( pTestSuit );
                }
                pTestCase->res = res;
                /*LOG("----- test case [ %s ] result ( %s ) \n", pTestCase->caseName, */
                    /*res == TEST_PASS ? "pass" : "fail" );*/
                pTestSuitManager->CancelThread( pTestSuit );
            }
        }
        pEventManager->DestroyAllEvent();
    }

    pTestSuitManager->Report();
    res = pthread_mutex_destroy( &pTestSuitManager->eventManager.mutex );
    if ( res == EINVAL ) {
        UT_ERROR("pthread_mutex_destroy() error, EINVAL \n");
        return ERROR_INVAL;
    } else if ( res == EBUSY ) {
        UT_ERROR("pthread_mutex_destroy() error, EBUSY \n");
        return ERROR_INVAL;
    } else {
    }

    return 0;
}

int TestSuitManagerInit()
{
    int i = 0;

    memset( pTestSuitManager, 0, sizeof(*pTestSuitManager) );
    pTestSuitManager->AddPrivateData = AddPrivateData;
    pTestSuitManager->startThread = startThread;
    pTestSuitManager->CancelThread = CancelThread;
    pTestSuitManager->Report = ResultReport;
    pTestSuitManager->eventManager.DestroyAllEvent = DestroyAllEvent;
    pTestSuitManager->eventManager.GetEventDataAddr = GetEventDataAddr;
    pTestSuitManager->eventManager.WaitForEvent = WaitForEvent;
    pTestSuitManager->eventManager.NotifyAllEvent = NotifyAllEvent;
    pthread_mutex_init( &pTestSuitManager->eventManager.mutex, NULL );


    return 0;
}


/*
 *
 *  structure :
 *      eventID0 -
 *               |
 *               \----- condition 0
 *               |
 *               \------ condition 1
 *               |
 *               \------ condition 2
 *               ...
 *      eventId1 -
 *              |
 *              \------ condition 8
 *              |
 *              \------ condition 9
 *              |
 *              \------ condition 10
 *              ...
 *               
 *
 * */
int NotifyAllEvent( int _nEventId, void *data )
{
    int i = 0;
    int j = 0;
    EventManger *pEventManager = &pTestSuitManager->eventManager;
    CondWait *pCondWait = NULL;

    /*UT_VAL( pEventManager->eventNum );*/

    pthread_mutex_lock( &pEventManager->mutex );
    for ( i=0; i<pEventManager->eventNum; i++ ) {
        EventWait *pEventWait = &pEventManager->eventWait[i];
        /*UT_VAL(_nEventId);*/
        /*UT_VAL(pEventWait->eventId);*/
        if ( _nEventId == pEventWait->eventId ) {
            /*UT_VAL( pEventWait->condNum );*/
            for ( j=0; j<pEventWait->condNum; j++ ) {
                pCondWait = &pEventWait->waitList[j];
                pthread_cond_signal( &pCondWait->cond );
                if ( pCondWait->eventCallBack ) {
                    pCondWait->eventCallBack( pCondWait->pTestCase, data );
                }
            }
        }
    }
    pthread_mutex_unlock( &pEventManager->mutex );

    return 0;
}

int AddEventWait( EventManger *pEventManager, EventWait *pEventWait,
                  EventCallBack eventCallBack, int nTimeOut, TestSuit *pTestSuit, void *data )
{
    struct timeval now;
    struct timespec after;
    int ret = 0;
    TestCase *pTestCase = NULL;

    if ( !pTestSuit ) {
        return -1;
    }

    ret = pthread_cond_init( &pEventWait->waitList[pEventWait->condNum].cond, NULL );
    if ( ret == EINVAL ) {
        UT_ERROR("pthread_cond_init error, EINVAL\n");
        return ERROR_INVAL;
    } else if ( ret == ENOMEM ) {
        UT_ERROR("pthread_cond_init error, ENOMEM\n");
        return ERROR_INVAL;
    }  else if ( ret == EAGAIN ) {
        UT_ERROR("pthread_cond_init error, EAGAIN\n");
        return ERROR_INVAL;
    }

    gettimeofday(&now, NULL);
    after.tv_sec = now.tv_sec + nTimeOut;
    /*after.tv_nsec = now.tv_usec * 1000 + 10 * 1000 * 1000;*/
    after.tv_nsec = 1000;
    pTestSuit->GetTestCase( pTestSuit, &pTestCase );
    if ( !pTestCase ) {
        UT_ERROR("GetTestCase() error\n");
        return -1;
    }
    pEventWait->waitList[pEventWait->condNum].eventCallBack = eventCallBack;
    pEventWait->waitList[pEventWait->condNum].pTestCase = pTestCase;
    UT_VAL( pEventWait->condNum );
    pEventWait->waitList[pEventWait->condNum].data = data;
    pTestCase->pCondWait = &pEventWait->waitList[pEventWait->condNum];
    int nReason = pthread_cond_timedwait( &pEventWait->waitList[pEventWait->condNum++].cond,
                                          &pEventManager->mutex, &after );
    if (nReason == ETIMEDOUT) {
        UT_ERROR("pthread_cond_timedwait time out\n");
        return ERROR_TIMEOUT;
    }
    if ( nReason == EINVAL ) {
        UT_NOTICE("pEventWait->eventId = %d\n", pEventWait->eventId );
        UT_NOTICE("pEventWait->condNum = %d\n", pEventWait->condNum );
        UT_ERROR("pthread_cond_timedwait TINVAL param error, nTimeOut = %d\n", nTimeOut );
        return ERROR_INVAL;
    }

    return STS_OK;
}

int WaitForEvent( int _nEventId, int nTimeOut, EventCallBack eventCallBack,
                  TestSuit *pTestSuit, void *data )
{
    int i = 0;
    int j = 0;
    int found = 0;
    int ret = 0;
    EventManger *pEventManager = &pTestSuitManager->eventManager;
    EventWait *pEventWait = NULL;

    /*UT_VAL(_nEventId);*/
    /*UT_VAL( pEventManager->eventNum );*/
    pthread_mutex_lock( &pEventManager->mutex );
    for ( i=0; i<pEventManager->eventNum; i++ ) {
        pEventWait= &pEventManager->eventWait[i];
        if ( _nEventId == pEventWait->eventId ) {
            ret = AddEventWait( pEventManager, pEventWait, eventCallBack, nTimeOut, pTestSuit, data );
            /*UT_VAL( ret );*/
            found = 1;
        }
    }

    if ( !found ) {
        pEventWait = &pEventManager->eventWait[pEventManager->eventNum];
        pEventWait->eventId = _nEventId;
        pEventManager->eventNum++;
        ret = AddEventWait( pEventManager, pEventWait, eventCallBack, nTimeOut, pTestSuit, data );
        UT_VAL( ret );
    }
    pthread_mutex_unlock( &pEventManager->mutex );

    return ret;
} 

int startThread( TestSuit *_pTestSuit )
{
    int ret = 0;
    TestCase *pTestCase = NULL;

    if ( !_pTestSuit ) {
        return -1;
    }

    _pTestSuit->GetTestCase( _pTestSuit, &pTestCase );
    if ( pTestCase ) {
        if ( pTestCase->threadEntry ) {
            ret = pthread_create( &pTestCase->tid, NULL,
                                  pTestCase->threadEntry, (void *)_pTestSuit );
        } else {
            if ( _pTestSuit->threadEntry ) {
                ret = pthread_create( &pTestCase->tid, NULL,
                                      _pTestSuit->threadEntry, (void *)_pTestSuit );
            }
        }
        if ( 0 != ret ) {
            UT_ERROR("create thread error, ret = %d\n", ret );
            return -1;
        }
        pTestCase->running = 1;
    }
    return 0;
}

int AddPrivateData( void *data )
{
    pTestSuitManager->data = data;
}


int CancelThread( TestSuit *_pTestSuit )
{
    int ret = 0;
    TestCase *pTestCase = NULL;

    _pTestSuit->GetTestCase( _pTestSuit, &pTestCase );
    if ( pTestCase ) {
        /*UT_NOTICE("cancel thread\n");*/
        pTestCase->running = 0;
    }

    return ret;
}

int ResultReport()
{
    int i = 0, j=0;
    TestSuit *pTestSuit = NULL;
    TestCase *pTestCase = NULL;

    LOG("total %d test suits\n", pTestSuitManager->num );
    for ( i=0; i<pTestSuitManager->num; i++ ) {
        pTestSuit = &pTestSuitManager->testSuits[i];
        if ( !pTestSuit->enable )
            continue;
        LOG("[============] test suit : %s, %d test(s)\n", pTestSuit->suitName, pTestSuit->total );
        for ( j=0; j<pTestSuit->total; j++ ) {
            pTestSuit->index = j;
            pTestSuit->GetTestCase( pTestSuit, &pTestCase );
            if ( pTestCase ) {
                /*LOG("*** test case ( %s ) result [ ", pTestCase->caseName );*/
                if ( pTestCase->res == TEST_PASS ) {
                    LOG("[    ");
                    LOG_GREEN("PASS");
                    LOG("    ]");
                    LOG(" %s\n", pTestCase->caseName );
                } else {
                    LOG("[    ");
                    LOG_RED("FAIL");
                    LOG("    ]");
                    LOG(" %s\n", pTestCase->caseName );
                }
            }
        }
    }
    return 0;
}

int DestroyAllEvent()
{
    EventManger *pEventManager = &pTestSuitManager->eventManager;
    int i = 0, j = 0;

    for ( i=0; i<pEventManager->eventNum; i++ ) {
        EventWait *pEventWait = &pEventManager->eventWait[i];
        for ( j=0; j<pEventWait->condNum; j++ ) {
            int ret = pthread_cond_destroy( &pEventWait->waitList[j].cond );
            if ( ret == EINVAL ) {
                UT_ERROR("pthread_cond_destroy() EINVAL\n");
                return ERROR_INVAL;
            } else if ( ret == EBUSY ) {
                UT_ERROR("pthread_cond_destroy() EBUSY\n");
                return ERROR_INVAL;
            } else {
                /*UT_LOG("destroy, eventId = %d, cond = %d OK\n", pEventWait->eventId, j );*/
            }
        }
        pEventWait->condNum = 0;
    }
    pEventManager->eventNum = 0;

    return 0;
}

int GetEventDataAddr( TestSuit *pTestSuit, void **data )
{
    TestCase *pTestCase = NULL;

    if ( pTestSuit ) {
        pTestSuit->GetTestCase( pTestSuit, &pTestCase );
        if ( pTestCase && pTestCase->pCondWait ) {
            *data = pTestCase->pCondWait->data;
            return 0;
        }
    }

    return -1;
}


