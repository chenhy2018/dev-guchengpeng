// Last Update:2018-09-27 11:19:57
/**
 * @file ipc_test.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-08-27
 */

#ifndef IPC_TEST_H
#define IPC_TEST_H

#define IPC_TRACE_TIMESTAMP 1
#define TIMESTAMP_REPORT_INTERVAL 20
#define TOKEN_RETRY_COUNT 1000
#define G711_TIMESTAMP_INTERVAL 40
#define FRAME_DATA_LEN 1024
#define STREAM_CACHE_SIZE 75

typedef enum {
    TYPE_VIDEO,
    TYPE_AUDIO,
} StreamType;

typedef enum {
    UPDATE_FROM_SOCKET,
    UPDATE_FROM_FILE,
};

typedef struct {
    unsigned char logOutput;
    unsigned char logVerbose;
    unsigned char logPrintTime;
    int timeStampPrintInterval;
    unsigned char heartBeatInterval;
    char *logFile;
    int tokenUploadInterval;
    int tokenRetryCount;
    char *bucketName;
    char *ak;
    char *sk;
    unsigned char movingDetection;
    int configUpdateInterval;
    unsigned char multiChannel;
    unsigned char openCache;
    int cacheSize;
    unsigned char updateFrom;
} Config;

typedef struct {
    char *data;
    unsigned char isKey;
    double timeStamp;
    int len;
} Frame;

extern Config *GetConfig();
extern int GetKodoInitSts();
extern int GetMovingDetectSts();
extern int GetAudioType();
extern void SetOutputType( int output );
extern int GetOutputType();
extern void SetMovingDetection( int enable );
extern int GetMovingDetection();
extern void SetUpdateFrom( int updateFrom );
extern void SetCache( int enable );

#endif  /*IPC_TEST_H*/
