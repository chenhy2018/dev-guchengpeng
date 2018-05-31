// Last Update:2018-05-31 15:50:49
/**
 * @file dbg.c
 * @brief 
 * @author
 * @version 0.1.00
 * @date 2018-05-31
 */

#include "sdk_interface.h"
#include "list.h"
#include "sdk_local.h"

#if SDK_DBG
void DumpUAList()
{
    UA *pUA = NULL;

    list_for_each_entry( pUA, &UaList.list, list ){
        if ( pUA ) {
            printf("[%02d] account id : \n", pUA->id );
        }
    }
    
}

void DbgBacktrace()
{
    void *array[20];
    size_t size;
    char **strings;
    size_t i;

    size = backtrace (array, 20);
    strings = backtrace_symbols (array, size);

    DBG_LOG ("Obtained %zd stack frames.\n", size);

    for (i = 0; i < size; i++)
        printf ("%s\n", strings[i]);

    free (strings);
}
#endif
