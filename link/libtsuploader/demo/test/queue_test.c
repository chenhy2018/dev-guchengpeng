// Last Update:2018-09-12 18:48:12
/**
 * @file queue_test.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-12
 */

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include "../queue.h"

int tests_run = 0;
char gBuffer[1024] = { 0 };

#define mu_assert(test) do { \
    if ( !(test) ) { \
        memset( gBuffer, 0, sizeof(gBuffer) );\
        sprintf( gBuffer, "*** test fail, line : %d, "#test, __LINE__ ); \
        return gBuffer;\
    } \
} while (0)

#define mu_run_test(test) do { LOG("=====> run test "#test"\n"); char *message = test();  tests_run++; \
                                if (message) return message; } while (0)

#define ASSERT_EQUAL( a, b ) mu_assert( a == b )  
#define ASSERT_STR_EQUAL( a, b ) mu_assert( strcmp(a, b) == 0 )

#define LOG(args...) printf("[ - %03d - ] ", __LINE__);printf(args)
#define LOG_VAL( a ) LOG( #a" = %d\n", a )
#define LOG_STR( s ) LOG( #s" = %s\n", s)
#define ARR_SZ(arr) sizeof(arr)/sizeof(arr[0])

static char * basicFunctionTest()
{
    Queue *q = NewQueue();
    char *s = "this is a test\n";
    char buffer[1024] = { 0 };
    int size = 0;
    int ret = 0;

    ret = q->enqueue( q, s, strlen(s) );
    LOG("q->getSize(q) = %d\n", q->getSize(q) );
    ASSERT_EQUAL( 1, q->getSize( q ) );
    LOG("ret = %d\n", ret );
    ret = q->dequeue( q, buffer, &size );
    ASSERT_EQUAL( ret, 0 );
    LOG("ret = %d\n", ret );
    LOG("size = %d\n", size );
    ASSERT_EQUAL(  size, strlen(s) );
    ASSERT_STR_EQUAL( buffer, s );

    return 0;
}

static char * TwoItemsTest()
{
    Queue *q = NewQueue();
    char *s1 = "hello";
    char *s2 = "world";
    char *s3 = "test";
    int size = 0;
    int ret = 0;
    int queueSize = 0;
    char buffer[1024] = { 0 };

    ret = q->enqueue( q, s1, strlen(s1) );
    ASSERT_EQUAL( ret, 0 );
    queueSize = q->getSize( q );
    ASSERT_EQUAL( queueSize, 1 );

    ret = q->enqueue( q, s2, strlen(s2) );
    ASSERT_EQUAL( ret, 0 );
    queueSize = q->getSize( q );
    LOG_VAL( queueSize );
    ASSERT_EQUAL( queueSize, 2 );

    ret = q->enqueue( q, s3, strlen(s3) );
    ASSERT_EQUAL( ret, 0 );
    queueSize = q->getSize( q );
    ASSERT_EQUAL( queueSize, 3 );

    memset( buffer, 0, sizeof(buffer) );
    ret = q->dequeue( q, buffer, &size );
    ASSERT_EQUAL( ret, 0 );
    ASSERT_STR_EQUAL( buffer, s1 );

    memset( buffer, 0, sizeof(buffer) );
    ret = q->dequeue( q, buffer, &size );
    ASSERT_EQUAL( ret, 0 );
    ASSERT_STR_EQUAL( buffer, s2 );

    memset( buffer, 0, sizeof(buffer) );
    ret = q->dequeue( q, buffer, &size );
    ASSERT_EQUAL( ret, 0 );
    ASSERT_STR_EQUAL( buffer, s3 );

    return 0;
}

static char *multi_data_test()
{
    Queue *q = NewQueue();
    int i = 0, ret = 0;

    for ( i=0; i<1000; i++ ) {
        q->enqueue( q, &i, 4 );
    }

    ret = q->getSize( q );
    ASSERT_EQUAL( ret, 1000 );

    return 0;
}

static char *intTypeTest()
{
    int val = 1234;
    int out = 0;
    int size = 0;
    int ret = 0;
    Queue *q = NewQueue();

    ret = q->enqueue( q, &val, 4 );
    ASSERT_EQUAL( ret, 0 );
    ret = q->dequeue( q, &out, &size );
    ASSERT_EQUAL( out, val );

    return 0;
}

char *s = "test block";
void *Producer( void *param )
{
    Queue *q = (Queue *)param;
    int ret = 0;

    sleep(6);
    LOG("enqueue a data\n");
    ret = q->enqueue( q, s, strlen(s) );
    LOG_VAL( ret );

    return NULL;
}

static char *BlockTest()
{
    Queue *q = NewQueue();
    char buffer[256] = { 0 };
    int size = 0, ret = 0;
    pthread_t thread;

    pthread_create( &thread, NULL, Producer, (void *)q );

    size = q->getSize( q );
    ASSERT_EQUAL( size, 0 );
    ret = q->dequeue( q, buffer, &size );
    LOG_VAL( ret );
    ASSERT_EQUAL( ret, 0 );
    size = q->getSize( q );
    ASSERT_EQUAL( size, 0 );
    LOG_STR( buffer );
    ASSERT_STR_EQUAL( buffer, s );

    return 0;
}

char *arr[] = 
{
    "12345",
    "hello",
    "world",
    "test",
    "queue",
};

void *ProducerTask( void *param )
{
    Queue *q = (Queue *)param;
    int i = 0;

    for ( i=0; i<ARR_SZ(arr); i++ ) {
        q->enqueue( q, arr[i], strlen(arr[i]) );
        sleep(3);
    }

    return NULL;
}

static char *TwoThreadTest()
{
    int i = 0, ret = 0;
    pthread_t thread;
    Queue *q = NewQueue();
    char buffer[1024] = { 0 };
    int size = 0;

    pthread_create( &thread, NULL, ProducerTask, (void *)q );
    
    for ( i=0; i<ARR_SZ(arr); i++ ) {
        memset( buffer, 0, sizeof(buffer) );
        ret = q->dequeue( q, buffer, NULL );
        ASSERT_EQUAL( ret, 0 );
        LOG_STR( buffer );
        ASSERT_STR_EQUAL( buffer, arr[i] );
    }
    size = q->getSize(q); 
    ASSERT_EQUAL( size, 0 );

    return 0;
}

static char *all_tests()
{
    mu_run_test( basicFunctionTest );
    mu_run_test( TwoItemsTest );
    mu_run_test( multi_data_test );
    mu_run_test( intTypeTest );
    mu_run_test( BlockTest );
    mu_run_test( TwoThreadTest );

    return 0;
}


int main()
{
    char *result = all_tests();
     if (result != 0) {
         printf("%s\n", result);
     }
     else {
         printf("ALL TESTS PASSED\n");
     }
     printf("Tests run: %d\n", tests_run);
    return 0;
}

