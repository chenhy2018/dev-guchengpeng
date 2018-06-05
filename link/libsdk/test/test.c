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

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))

int RegisterTestSuitCallback( TestSuit *this );
int RegisterTestSuitInit( TestSuit *this );
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

#define HOST "123.59.204.198"
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "normal", RET_OK },
        { "1006", "1006", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", RET_ACCOUNT_NOT_EXIST },
        { "1006", "1007", HOST, HOST, HOST, 100, 0 }
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


int RegisterTestSuitInit( TestSuit *this )
{
    this->total = ARRSZ(gRegisterTestCases);
    this->index = 0;

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
    Media media;
    int i = 0;
    ErrorID sts = 0;

    if ( !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *) this->testCases;
    DBG_LOG("this->index = %d\n", this->index );
    pData = &pTestCases[this->index].data;

    if ( pData->init ) {
        sts = InitSDK( &media, 1 );
        if ( RET_OK != sts ) {
            DBG_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }

    DBG_STR( pData->id );
    DBG_STR( pData->password );
    DBG_STR( pData->sigHost );
    sts = Register( pData->id, pData->password, pData->sigHost, pData->mediaHost, pData->imHost, pData->timeOut );
    return TEST_EQUAL( sts, pTestCases->father.expact );
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
    InitAllTestSuit();
    RunAllTestSuits();

    return 0;
}

