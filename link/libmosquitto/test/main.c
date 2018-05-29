#include "mosquitto.h"
#include <unistd.h>
#include <string.h>
#include <stdio.h>

void onMessage(IN const void* instance, IN const char* topic, IN const char* message, IN size_t length)
{
        fprintf(stderr, "topic %s message %s \n", topic,  message);
}

void onEvent(IN const void* instance, IN int id,  IN const char* reason)
{
        fprintf(stderr, "id %d, reason  %s \n", id, reason);
}

int main()
{
        struct MosquittoOptions options;
        MosquittoLibInit();
        options.pId = "test";
        options.bCleanSession = false;
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_NULL;
        options.primaryUserInfo.pHostname = "123.59.204.198";
        //strcpy(options.bindaddress, "172.17.0.2");
        options.secondaryUserInfo.pHostname = "172.17.0.4";
        //strcpy(options.secondBindaddress, "172.17.0.2`");
        options.primaryUserInfo.pUsername = "root";
        options.primaryUserInfo.pPassword = "root";
        options.secondaryUserInfo.pUsername = "test";
        options.secondaryUserInfo.pPassword = "111";
        options.secondaryUserInfo.nPort = 1883;
        options.primaryUserInfo.nPort = 1883;
        options.primaryUserInfo.pCafile = NULL;
        options.primaryUserInfo.pCertfile = NULL;
        options.primaryUserInfo.pKeyfile = NULL;
        options.secondaryUserInfo.pCafile = NULL;
        options.secondaryUserInfo.pCertfile = NULL;
        options.secondaryUserInfo.pKeyfile = NULL;
        options.nKeepalive = 10;
        options.nQos = 0;
        options.bRetain = false;
        options.callbacks.onMessage = &onMessage;
        options.callbacks.onEvent = &onEvent;
        void* instance = NULL;
        printf("try first \n");
        instance = MosquittoCreateInstance(&options);
        sleep(3);
        MosquittoSubscribe(instance, "sensor/room1/#");
        for (int i = 0 ; i < 10; ++i) {
            MosquittoPublish(instance, "sensor/room1/temperature", 10, "test1234456");
        }
        MosquittoDestroy(instance);
        printf("try second \n");
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_USER;
        options.primaryUserInfo.pHostname = "123.59.204.198";
        options.primaryUserInfo.pUsername = "root";
        options.primaryUserInfo.pPassword = "root";
        instance = MosquittoCreateInstance(&options);
                sleep(3);
        MosquittoSubscribe(instance, "sensor/room1/#");
        for (int i = 0 ; i < 10; ++i) {
              MosquittoPublish(instance, "sensor/room1/temperature", 10, "test1234456");
        }
        sleep(5); 
        MosquittoDestroy(instance);
        MosquittoLibCleanup();
        printf("try third \n");
        //memset(options.primaryUserInfo.hostname, 0, MAX_MOSQUITTO_USR_SIZE);
        //strcpy(options.primaryUserInfo.hostname, "172.17.0.8");
        options.primaryUserInfo.pCafile = "./test/ca.crt";
        options.primaryUserInfo.nPort = 8883;
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_USER | MOSQUITTO_AUTHENTICATION_ONEWAY_SSL;
        MosquittoLibInit();
        while(1) {
                //MosquittoLibInit();
                instance = MosquittoCreateInstance(&options);
                sleep(1);
                MosquittoSubscribe(instance, "sensor/room1/#");
                for (int i = 0; i < 10; ++ i) {
                        MosquittoPublish(instance, "sensor/room1/temperature", 10, "test1234456");
                        usleep(100000);
                }
                sleep(3);
                MosquittoDestroy(instance);
                //MosquittoLibCleanup();
        }
        MosquittoLibCleanup();
        return 1;
}
