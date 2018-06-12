// Last Update:2018-06-12 19:00:55
/**
 * @file test.c
 * @brief 
 * @author
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
#include <unistd.h> 
#include "sdk_interface.h"
#include "dbg.h"
#include "unit_test.h"
#include "test.h"
#include "call_test.h"
#include "register_test.h"

int InitAllTestSuit()
{
    AddTestSuit( &gRegisterTestSuit );
    AddTestSuit( &gMakeCallTestSuit );

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

