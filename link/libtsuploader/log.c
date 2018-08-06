#include "log.h"

int nLogLevel = LOG_LEVEL_INFO;

void SetLogLevelToTrace()
{
        nLogLevel = LOG_LEVEL_TRACE;
}

void SetLogLevelToDebug()
{
        nLogLevel = LOG_LEVEL_DEBUG;
}

void SetLogLevelToInfo()
{
        nLogLevel = LOG_LEVEL_INFO;
}

void SetLogLevelToWarn()
{
        nLogLevel = LOG_LEVEL_WARN;
}

void SetLogLevelToError()
{
        nLogLevel = LOG_LEVEL_ERROR;
}
