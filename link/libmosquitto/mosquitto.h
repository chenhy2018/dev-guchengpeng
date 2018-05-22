#ifndef __MOSQUITTO__
#define __MOSQUITTO__

#include <mosquitto.h>
#include <stdbool.h>
#include <stddef.h>

#define IN
#define OUT

#define MAX_MOSQUITTO_ID_SIZE 50
#define MAX_MOSQUITTO_HOST_SIZE  128
#define MAX_MOSQUITTO_USR_SIZE   64
#define MAX_MOSQUITTO_PWD_SIZE   64
#define MAX_MOSQUITTO_TOPIC_SIZE   64
#define MAX_MOSQUITTO_MID_SIZE   10

enum {
       MOSQUITTO_ERR_SUCCESS = 3000,
       MOSQUITTO_ERR_INVAL = 3001,
       MOSQUITTO_ERR_NOMEM,
       MOSQUITTO_ERR_NO_CONN,
       MOSQUITTO_ERR_PROTOCOL,
       MOSQUITTO_ERR_PAYLOAD_SIZE,
       MOSQUITTO_ERR_OTHERS
};

typedef struct MosquittoOptions MosquittoOptions;

struct MosquittoCallback
{
        void (*onMessage)(IN const void* instance, IN const char* message, IN size_t length);
        void (*onError)(IN const void* instance, IN int errorCode, const char* reason);
};

struct MosquittoUserInfo
{
        char username[MAX_MOSQUITTO_USR_SIZE];
        char password[MAX_MOSQUITTO_PWD_SIZE];
        char hostname[MAX_MOSQUITTO_HOST_SIZE];
        int nPort;
        //char bindaddress[MAX_MOSQUITTO_HOST_SIZE]; //not used in current time.
};

struct MosquittoOptions
{
        char id[MAX_MOSQUITTO_ID_SIZE];
        bool bCleanSession;
        struct MosquittoUserInfo primaryUserInfo;
        struct MosquittoUserInfo secondaryUserInfo;
        int nKeepalive;
        struct MosquittoCallback callbacks; // A user pointer that will be passed as an argument to any callbacks that are specified.
        int nQos;
        bool bRetain;
};

/* step 1 : create mosquitto instance */
extern void* MosquittoCreateInstance(IN const struct MosquittoOptions* pOption);

extern void MosquittoDestroy(IN const void* pInstance);

/* step 2 : mosquitto pub/sub */

extern int MosquittoPublish(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload);

extern int MosquittoSubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic);

extern int MosquittoUnsubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* pSub);

#endif
