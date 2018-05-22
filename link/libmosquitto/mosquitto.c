#include "mosquitto.h"
#include <stdio.h>
#include <pthread.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <unistd.h>

#define STATUS_IDLE 0
#define STATUS_CONNECTING 1
#define STATUS_CONNACK_RECVD 2
#define STATUS_WAITING 3
#define STATUS_DISCONNECTING 4
#define STATUS_PUBLISHING 5
#define STATUS_PUBLISHED 6

#define MinQueueSize 50

typedef struct Node
{
        char topic[MAX_MOSQUITTO_TOPIC_SIZE];
        struct Node *pNext;
}Node;

bool insertNode(Node* pHead, char* val) {
        int i = 0;
        Node* p = pHead;
        while(NULL != p->pNext)
        {
                i++;
                p = p->pNext;
        }
        Node* pNew = malloc(sizeof(Node));
        strcpy(pNew->topic, val);
        p->pNext = pNew;
        pNew->pNext = NULL;
        return true;
}


bool deleteNode(Node* PHead, char * pval)
{
        int i = 0;
        Node* p = PHead;
        while(p->pNext != NULL){
               if (strcmp(p->pNext->topic, pval) == 0) {
                       Node* temp = p->pNext;
                       p->pNext = temp->pNext;
                       free(temp);
                       return true;
               }
        }
        printf("Can't find \n");
        return false;
}

struct MosquittoInstance
{
        struct mosquitto *mosq;
        struct MosquittoOptions options;
        int status;
        bool connected;
        Node pSubsribeList;
};

void onLogCallback(struct mosquitto* _pMosq, void* _pObj, int level, const char* _pStr)
{
        printf("%s\n", _pStr);
}

void onMessageCallback(struct mosquitto* _pMosq, void* _pObj, const struct mosquitto_message* _pMessage)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        if (pInstance->options.callbacks.onMessage) {
                pInstance->options.callbacks.onMessage(_pObj, _pMessage->payload, _pMessage->payloadlen);
        }
}

void onConnectCallback(struct mosquitto* _pMosq, void* _pObj, int result)
{
        int rc = MOSQ_ERR_SUCCESS;
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        pInstance->status = STATUS_WAITING;
        pInstance->connected = true;
        fprintf(stderr, " on_connect_callback \n ");
        if (result) {
                fprintf(stderr, "%s\n", mosquitto_connack_string(result));
        }
        else {
                Node* p = pInstance->pSubsribeList.pNext;
                while (p) {
                        mosquitto_subscribe(pInstance->mosq, NULL, p->topic, pInstance->options.nQos);
                        p = p->pNext;
                }
        }
}


void onDisconnectCallback(struct mosquitto* _pMosq, void* _pObj, int rc)
{

        fprintf(stderr, " on_disconnect_callback \n ");
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        pInstance->connected = false;
        pInstance->status = STATUS_IDLE;
}

void onSubscribeCallback(struct mosquitto* _pMosq, void* pObj, int mid, int qos_count, const int* pGranted_qos)
{       
        fprintf(stderr, "Subscribed (mid: %d): %d", mid, pGranted_qos[0]);
}

void onUnsubscribeCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, "Unsubscribed (mid: %d)", mid);
}

void onPublishCallback(struct mosquitto* _pMosq, void* _pObj, int mid)
{
        fprintf(stderr, " my_publish_callback \n ");
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pObj);
        int last_mid_sent = mid;
        pInstance->status = STATUS_PUBLISHED;
}


static void MosquittoInstanceInit(struct MosquittoInstance* _pInstance, const struct MosquittoOptions* _pOption)
{
        /* copy options */
        memcpy(&_pInstance->options, _pOption, sizeof(struct MosquittoOptions));
        _pInstance->options.id[MAX_MOSQUITTO_ID_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.username[MAX_MOSQUITTO_USR_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.password[MAX_MOSQUITTO_PWD_SIZE - 1] = 0;
        _pInstance->options.primaryUserInfo.hostname[MAX_MOSQUITTO_HOST_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.username[MAX_MOSQUITTO_USR_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.password[MAX_MOSQUITTO_PWD_SIZE - 1] = 0;
        _pInstance->options.secondaryUserInfo.hostname[MAX_MOSQUITTO_HOST_SIZE - 1] = 0;

        _pInstance->mosq = NULL;
        _pInstance->connected = false;
        _pInstance->status = STATUS_IDLE;
        _pInstance->pSubsribeList.pNext = NULL;
}

void * Mosquittothread(void* _pData)
{
        int rc;
        mosquitto_lib_init();
        
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pData);

        pInstance->mosq = mosquitto_new(pInstance->options.id, true, pInstance);
        if (!pInstance->mosq) {
                switch(errno) {
                        case ENOMEM:
                                fprintf(stderr, "Error: Out of memory.\n");
                                break;
                        case EINVAL:
                                fprintf(stderr, "Error: Invalid id.\n");
                                break;
                }
                mosquitto_lib_cleanup();
                return NULL;
        }
        mosquitto_threaded_set(pInstance->mosq, true);
        //mosquitto_log_callback_set(pInstance->mosq, onLogCallback);
        mosquitto_connect_callback_set(pInstance->mosq, onConnectCallback);
        mosquitto_disconnect_callback_set(pInstance->mosq, onDisconnectCallback);
        mosquitto_publish_callback_set(pInstance->mosq, onPublishCallback);
        mosquitto_message_callback_set(pInstance->mosq, onMessageCallback);
        mosquitto_subscribe_callback_set(pInstance->mosq, onSubscribeCallback);
        mosquitto_unsubscribe_callback_set(pInstance->mosq, onUnsubscribeCallback);
        
        do {
                 if (!pInstance->connected && pInstance->status == STATUS_IDLE) {
                         fprintf(stderr, "connecting \n");
                         pInstance->status = STATUS_CONNECTING;
                         mosquitto_username_pw_set(pInstance->mosq, pInstance->options.primaryUserInfo.username, pInstance->options.primaryUserInfo.password);
                         rc = mosquitto_connect(pInstance->mosq, pInstance->options.primaryUserInfo.hostname, pInstance->options.primaryUserInfo.nPort, pInstance->options.nKeepalive);
                         if (rc) {
                                 sleep(1);
                                 
                                 fprintf(stderr, "Unable to connect (%s). try to reconnect to secondary server. \n", mosquitto_strerror(rc));
                                 mosquitto_username_pw_set(pInstance->mosq, pInstance->options.secondaryUserInfo.username, pInstance->options.secondaryUserInfo.password);
                                 rc = mosquitto_connect(pInstance->mosq, pInstance->options.secondaryUserInfo.hostname, pInstance->options.secondaryUserInfo.nPort, pInstance->options.nKeepalive);
                                 if (rc) {
                                         fprintf(stderr, "Unable to connect Secondary server  %s \n", mosquitto_strerror(rc) );
                                         pInstance->status = STATUS_IDLE;
                                         // TODO add error callback.
                                         sleep(1);
                                 }
                                 else {
                                         pInstance->status = STATUS_CONNECTING;
                                 }
                         }
                 }
                 rc = mosquitto_loop(pInstance->mosq, -1, 1);
        } while (pInstance->status != STATUS_DISCONNECTING);
        if (pInstance->connected) {
                mosquitto_disconnect(pInstance->mosq);
        }
        mosquitto_destroy(pInstance->mosq);
        mosquitto_lib_cleanup();
        if (rc) {
                fprintf(stderr, "Error: %s\n", mosquitto_strerror(rc));
        }
        return NULL;
}

void* MosquittoCreateInstance(IN const struct MosquittoOptions* pOption)
{
        /* allocate one mosquitto instance struct */
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)malloc(sizeof(struct MosquittoInstance));
        if (pInstance == NULL) {
                return NULL;
        }
        
        MosquittoInstanceInit(pInstance, pOption);
        pthread_t t;
        pthread_create(&t, NULL, Mosquittothread, pInstance);
        return pInstance;
}

void MosquittoDestroy(IN const void* _pInstance)
{	
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        pInstance->status = STATUS_DISCONNECTING;
}

static int MosquittoErrorStatusChange(int nStatus)
{
        switch (nStatus) {
                case MOSQ_ERR_INVAL:
                        return MOSQUITTO_ERR_INVAL;
                case MOSQ_ERR_NOMEM:
                        return MOSQUITTO_ERR_NOMEM;
                case MOSQ_ERR_NO_CONN:
                        return MOSQUITTO_ERR_NO_CONN;
                case MOSQ_ERR_PROTOCOL:
                        return MOSQUITTO_ERR_PROTOCOL;
                case MOSQ_ERR_PAYLOAD_SIZE:
                        return MOSQUITTO_ERR_PAYLOAD_SIZE;
                case MOSQ_ERR_SUCCESS:
                        return MOSQUITTO_ERR_SUCCESS;
                default:
                        return MOSQUITTO_ERR_OTHERS;
        }
        return MOSQUITTO_ERR_OTHERS;
}

int MosquittoPublish(IN const void* _pInstance, IN int* _pMid, IN char* _pTopic, IN int _nPayloadlen, IN const void* _pPayload)
{
       struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
       int rc = mosquitto_publish(pInstance->mosq, NULL, _pTopic, _nPayloadlen, _pPayload, pInstance->options.nQos, pInstance->options.bRetain);
       if (rc) {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input. Does your topic contain '+' or '#'?\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to publish message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN:
                               fprintf(stderr, "Error: Client not connected when trying to publish.\n");
                               break;
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
                       case MOSQ_ERR_PAYLOAD_SIZE:
                               fprintf(stderr, "Error: Message payload is too large.\n");
                               break;
               }
       }
       return rc;
}

int MosquittoSubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pTopic)
{
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        int rc = mosquitto_subscribe(pInstance->mosq, _pMid, _pTopic, pInstance->options.nQos);
        fprintf(stderr, "mos sub %d", rc);
        if (!rc) {
                insertNode(&pInstance->pSubsribeList, _pTopic);
        }
        else {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input.\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to subscribe message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN:
                               fprintf(stderr, "Error: Client not connected when trying to subscribe.\n");
                               break;
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
                       case MOSQ_ERR_PAYLOAD_SIZE:
                               fprintf(stderr, "Error: Message payload is too large.\n");
                               break;
               }
        }
        return rc;
}

int MosquittoUnsubscribe(IN const void* _pInstance, OUT int* _pMid, IN char* _pSub)
{
        struct MosquittoInstance* pInstance = (struct MosquittoInstance*)(_pInstance);
        int rc = mosquitto_unsubscribe(pInstance->mosq, _pMid, _pSub);
        fprintf(stderr, "mos sub %d", rc);
        if (!rc) {
               deleteNode(&pInstance->pSubsribeList, _pSub);
        }
        else {
               switch (rc) {
                       case MOSQ_ERR_INVAL:
                               fprintf(stderr, "Error: Invalid input.\n");
                               break;
                       case MOSQ_ERR_NOMEM:
                               fprintf(stderr, "Error: Out of memory when trying to unsubscribe message.\n");
                               break;
                       case MOSQ_ERR_NO_CONN: 
                               fprintf(stderr, "Error: Client not connected when trying to unsubscribe.\n");
                               break; 
                       case MOSQ_ERR_PROTOCOL:
                               fprintf(stderr, "Error: Protocol error when communicating with broker.\n");
                               break;
               }
        }
        return rc;
}
