// Last Update:2018-06-07 18:33:09
/**
 * @file unit_test.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-05
 */

#ifndef UNIT_TEST_H
#define UNIT_TEST_H

#include <pthread.h>

#define TEST_FAIL 1
#define TEST_PASS 0
#define TEST_EQUAL( actual, expect ) ( actual == expect ? TEST_PASS : TEST_FAIL )
#define TEST_GT( actual, expect ) ( actual >= expect ? TEST_PASS : TEST_FAIL )

#define TEST_SUIT_MAX 256
#define LOG(args...) printf(args)
#define COND_MAX (256)
#define EVENT_WAIT_MAX (256)
#define THREAD_MAX (16)
#define ERROR_TIMEOUT (-2)
#define ERROR_INVAL (-3)
#define STS_OK (0)

typedef struct {
    char *caseName;
    int expact;
} TestCase;


typedef struct _TestSuit TestSuit;
typedef struct _TestSuitManager TestSuitManager; 
struct _TestSuit {
    char *suitName;
    int (*TestCaseCb)( TestSuit *this );
    int (*OnInit)( TestSuit *this, TestSuitManager *pManager );
    int (*GetTestCase) ( TestSuit *this, TestCase **testCase );
    void *testCases;
    int total;
    int failNum;
    int index;
    TestSuitManager *pManager;
};

typedef struct {
    int eventId;
    int condNum;
    pthread_cond_t condList[COND_MAX];
} EventWait;

typedef struct {
    int eventNum;
	pthread_mutex_t    mutex;
    EventWait eventWait[EVENT_WAIT_MAX];
    int (*WaitForEvent)( int event, int timeout );
} EventManger;

typedef void *(*ThreadFn)(void *);

typedef struct {
    pthread_t threadId;
    ThreadFn threadFn;
} ThreadInfo;

typedef struct {
    ThreadInfo threadList[THREAD_MAX];
    int num;
} ThreadManager;

struct _TestSuitManager{
    void *data;
    int num;
    TestSuit testSuits[TEST_SUIT_MAX];
    EventManger eventManager;
    ThreadManager threadManager;
    int (* NotifyAllEvent)( int _nEventId );
    int (*AddPrivateData)( void *data );
};

extern int AddTestSuit( TestSuit *pTestSuit );
extern int RunAllTestSuits();
extern int TestSuitManagerInit();
extern int NotifyAllEvent( int event );
extern int WaitForEvent( int _nEventId, int nTimeOut );
extern int ThreadRegister( ThreadFn threadFn );
extern int AddPrivateData( void *data );

#endif  /*UNIT_TEST_H*/
