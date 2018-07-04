// Last Update:2018-07-03 17:17:04
/**
 * @file common.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-29
 */

#ifndef COMMON_H
#define COMMON_H

#include "dbg.h"

#if DBG_STREAM_WRITE_FILE
#undef DBG_ERROR
#define DBG_ERROR(...)
#endif

typedef unsigned char u_int8_t;
typedef unsigned int u_int32_t;

#define CHECK_SDK_RETURN( function, ret ) \
    if ( ret >= RET_MEM_ERROR ) {\
        DBG_ERROR(#function"() error,"#ret"=%d\n", ret); \
        return -1; \
    }
#define ARRSZ(arr) sizeof(arr)/sizeof(arr[0])

#endif  /*COMMON_H*/
