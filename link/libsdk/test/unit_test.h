// Last Update:2018-06-20 19:25:47
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
#include <stdio.h>

#define TEST_FAIL 1
#define TEST_PASS 0
#define TEST_EQUAL( actual, expect ) ( actual == expect ? TEST_PASS : TEST_FAIL )
#define TEST_GT( actual, expect ) ( actual >= expect ? TEST_PASS : TEST_FAIL )

#define NONE                 "\e[0m"
#define BLACK                "\e[0;30m"
#define L_BLACK              "\e[1;30m"
#define RED                  "\e[0;31m"
#define L_RED                "\e[1;31m"
#define GREEN                "\e[0;32m"
#define L_GREEN              "\e[1;32m"
#define BROWN                "\e[0;33m"
#define YELLOW               "\e[1;33m"
#define BLUE                 "\e[0;34m"
#define L_BLUE               "\e[1;34m"
#define PURPLE               "\e[0;35m"
#define L_PURPLE             "\e[1;35m"
#define CYAN                 "\e[0;36m"
#define L_CYAN               "\e[1;36m"
#define GRAY                 "\e[0;37m"
#define WHITE                "\e[1;37m"

#define BOLD                 "\e[1m"
#define UNDERLINE            "\e[4m"
#define BLINK                "\e[5m"
#define REVERSE              "\e[7m"
#define HIDE                 "\e[8m"
#define CLEAR                "\e[2J"
#define CLRLINE              "\r\e[K" //or "\e[1K\r"

#define TEST_SUIT_MAX 256
#define LOG(args...) printf(args)
#define COND_MAX (256)
#define EVENT_WAIT_MAX (256)
#define THREAD_MAX (16)
#define ERROR_TIMEOUT (-2)
#define ERROR_INVAL (-3)
#define STS_OK (0)
#undef DBG_LOG
#undef DBG_BASIC
#undef DBG_ERROR
#define DBG_BASIC() printf("[ %s %s +%d ] ", __FILE__, __FUNCTION__, __LINE__)
#define DBG_LOG(args...) DBG_BASIC();printf(" ");printf(args)
#define DBG_ERROR DBG_LOG
#define UT_LOG(args...) printf("[ UNIT TEST ] ");DBG_LOG(args)
//#define UT_ERROR(args...) printf("[ UNIT TEST] ");DBG_ERROR(args)
#define UT_VAL(v) printf("[ UNIT TEST ] ");DBG_LOG(#v" = %d\n", v)
#define UT_STR(s) printf("[ UNIT TEST ]");DBG_LOG(#s" = %s\n", s)
#define UT_LINE() printf("[ UNIT TEST ]");DBG_LOG("======================\n")
#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define TEST_CASE_RESULT_MAX 512
#define UT_NOTICE(args...) printf( RED"[ %s %s +%d ] "NONE, __FILE__, __FUNCTION__, __LINE__);printf(args)
#define LOG_RED(args...) printf( RED"" );printf(args);printf(""NONE)
#define LOG_GREEN(args...) printf( GREEN"" );printf(args);printf(""NONE)
#define LOG_BLUE(args...) printf( BLUE"" );printf(args);printf(""NONE)
#define UT_ERROR(args...) LOG_RED("[ UNIT TEST] "); LOG_RED("[ %s %s +%d ] ", __FILE__, __FUNCTION__, __LINE__);LOG_RED(args)

typedef struct _CondWait CondWait;
typedef struct _TestSuit TestSuit;
typedef void *(*ThreadFn)(void *);
typedef struct {
    char *caseName;
    int expact;
    ThreadFn threadEntry;
    int (*TestCaseCb)( TestSuit *this );
    pthread_t tid;
    void *data;
    unsigned char running;
    int res;
    CondWait *pCondWait;
} TestCase;

typedef int (*EventCallBack)( TestCase *pTestCase, void *data );
typedef struct _TestSuitManager TestSuitManager; 
struct _TestSuit {
    char *suitName;
    int (*TestCaseCb)( TestSuit *this );
    int (*OnInit)( TestSuit *this, TestSuitManager *pManager );
    int (*GetTestCase) ( TestSuit *this, TestCase **testCase );
    void *testCases;
    unsigned char enable;
    ThreadFn threadEntry;
    int total;
    int failNum;
    int index;
    TestSuitManager *pManager;
    pthread_t tid;
};

struct _CondWait {
    pthread_cond_t cond;
    TestCase *pTestCase;
    EventCallBack eventCallBack;
    void *data;
};

typedef struct {
    int eventId;
    int condNum;
    CondWait waitList[COND_MAX];
} EventWait;

typedef struct {
    int eventNum;
	pthread_mutex_t mutex;
    EventWait eventWait[EVENT_WAIT_MAX];
    int (*WaitForEvent)( int _nEventId, int nTimeOut, EventCallBack eventCallback,
                  TestSuit *pTestSuit, void *data );
    int (* NotifyAllEvent)( int _nEventId, void *data );
    int (*DestroyAllEvent)();
    int (*GetEventDataAddr)( TestSuit *pTestSuit, void **data );
} EventManger;

typedef struct {
    char *pTestCaseName;
    int res;
} TestCaseResult;

struct _TestSuitManager{
    void *data;
    int num;
    TestSuit testSuits[TEST_SUIT_MAX];
    EventManger eventManager;
    int (*AddPrivateData)( void *data );
    int (*startThread)( TestSuit *_pTestSuit );
    int (*CancelThread)( TestSuit *pTestSuit );
    int (*Report)();
};

extern int AddTestSuit( TestSuit *pTestSuit );
extern int RunAllTestSuits();
extern int TestSuitManagerInit();
extern int NotifyAllEvent( int event, void *data );
extern int AddPrivateData( void *data );
extern int startThread( TestSuit *_pTestSuit );
extern int CancelThread( TestSuit *_pTestSuit );
extern int ResultReport();
extern int DestroyAllEvent();
extern int GetEventDataAddr( TestSuit *pTestSuit, void **data );
extern int WaitForEvent( int _nEventId, int nTimeOut,
                         EventCallBack eventCallback, TestSuit *pTestSuit, void *data );

#endif  /*UNIT_TEST_H*/
