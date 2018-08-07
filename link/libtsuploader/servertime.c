#include <fcntl.h>
#include <unistd.h>
#include <curl/curl.h>
#include <string.h>
#include "base.h"
#ifdef __APPLE__
#include <sys/types.h>
#include <sys/sysctl.h>
#else
#include <time.h>
#endif

#define USE_CLOCK 1

#ifndef __APPLE__
#ifdef USE_CLOCK
static struct timespec tmResolution;
#endif
#endif

int uptimefd = -1;

static int64_t getUptime()
{
#ifndef __APPLE__
    #ifdef USE_CLOCK
        struct timespec tp;
        clock_gettime(CLOCK_MONOTONIC, &tp);
        return (int64_t)(tp.tv_sec * 1000000000 + tp.tv_nsec / tmResolution.tv_nsec);
    #else
        char str[33];
        if(uptimefd < 0) {
                uptimefd = open("/proc/uptime", O_RDONLY);
        }
        int nReadLen = 0;
        if (uptimefd >= 0) {
                lseek(uptimefd, 0, SEEK_SET);
                nReadLen = read(uptimefd, str, sizeof(str));
        } else {
                return -1;
        }
        str[nReadLen - 1] = 0;
        char *pSpace = strchr(str, ' ');
        *pSpace = 0;
        return (int64_t)(atof(str) * 1000000000);
    #endif
#else
        struct timeval tm;
        size_t s = sizeof(tm);
        memset(&tm, 0, sizeof(tm));
        sysctlbyname("kern.boottime", &tm, &s, NULL, 0);
        return (int64_t)(tm.tv_sec * 1000000000 + tm.tv_usec * 1000);
#endif
}

static int64_t nServerTimestamp;
static int64_t nLocalupTimestamp;

struct ServerTime{
        char * pData;
        int nDataLen;
        int nCurlRet;
};

static size_t writeTime(void *pTimeStr, size_t size,  size_t nmemb,  void *pUserData) {
        struct ServerTime *pTime = (struct ServerTime *)pUserData;
        if (pTime->nDataLen < size * nmemb) {
                pTime->nCurlRet = -11;
                return 0;
        }
       
        memcpy(pTime->pData, pTimeStr, size * nmemb);
        return size * nmemb;
}

static int getTimeFromServer(int64_t *pStime) {
        CURL *curl;
        curl_global_init(CURL_GLOBAL_ALL);
        curl = curl_easy_init();
        
        curl_easy_setopt(curl, CURLOPT_URL, "http://127.0.0.1:12345/hello");//"http://39.107.247.14:8086/qiniu/upload/token");
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, writeTime);
        
        char timeStr[128] = {0};
        struct ServerTime stime;
        stime.pData = timeStr;
        stime.nDataLen = sizeof(timeStr);
        stime.nCurlRet = 0;
        
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, &stime);
        int ret =curl_easy_perform(curl);
        
        curl_easy_cleanup(curl);
        if(ret == 0) {
                *pStime = (int64_t)atoll(stime.pData);
        }
        return ret;
}

int64_t GetCurrentMillisecond()
{
        int64_t nUptime = getUptime();
        return (nUptime - nLocalupTimestamp)+nServerTimestamp;
}

int InitTime() {
#ifndef __APPLE__
    #ifdef USE_CLOCK
        clock_getres(CLOCK_MONOTONIC, &tmResolution);
    #endif
#endif
        int ret = 0;
        ret = getTimeFromServer(&nServerTimestamp);
        nLocalupTimestamp = getUptime();
        return ret;
}
