// Last Update:2018-09-04 12:28:41
/**
 * @file log2file.c
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-09-03
 */

#include <stdio.h>
#include <string.h>

static FILE *gFd = NULL;

int fileOpen( char *_pLogFile )
{
    gFd = fopen( _pLogFile, "w+" );
    if ( !gFd ) {
        return -1;
    }

    return 0;
}

int writeLog( char *log )
{
    fwrite( log, strlen(log), 1, gFd );
    fflush( gFd );

    return 0;
}

