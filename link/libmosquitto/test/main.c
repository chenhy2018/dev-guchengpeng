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
        strcpy(options.id, "test");
        options.bCleanSession = false;
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_NULL;
        strcpy(options.primaryUserInfo.hostname, "123.59.204.198");
        //strcpy(options.bindaddress, "172.17.0.2");
        strcpy(options.secondaryUserInfo.hostname, "172.17.0.4");
        //strcpy(options.secondBindaddress, "172.17.0.2`");
        strcpy(options.primaryUserInfo.username, "root");
        strcpy(options.primaryUserInfo.password, "root");
        strcpy(options.secondaryUserInfo.username, "test");
        strcpy(options.secondaryUserInfo.password, "111");
        options.secondaryUserInfo.nPort = 1883;
        options.primaryUserInfo.nPort = 1883;
        options.nKeepalive = 10;
        options.nQos = 0;
        options.bRetain = false;
        options.callbacks.onMessage = &onMessage;
        options.callbacks.onEvent = &onEvent;
        void* instance = NULL;
        printf("try first \n");
        instance = MosquittoCreateInstance(&options);
        sleep(3);
        MosquittoSubscribe(instance, NULL, "sensor/room1/#");
        for (int i = 0 ; i < 10; ++i) {
            MosquittoPublish(instance, NULL, "sensor/room1/temperature", 10, "test1234456");
            usleep(1000000);
        }
        MosquittoDestroy(instance);
        printf("try second \n");
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_USER;
        strcpy(options.primaryUserInfo.hostname, "123.59.204.198");
        strcpy(options.primaryUserInfo.username, "root");
        strcpy(options.primaryUserInfo.password, "root");
        instance = MosquittoCreateInstance(&options);
                sleep(3);
        MosquittoSubscribe(instance, NULL, "sensor/room1/#");
        for (int i = 0 ; i < 10; ++i) {
            MosquittoPublish(instance, NULL, "sensor/room1/temperature", 10, "test1234456");
            usleep(1000000);
        } 
        MosquittoDestroy(instance);
        printf("try third \n");
        //memset(options.primaryUserInfo.hostname, 0, MAX_MOSQUITTO_USR_SIZE);
        //strcpy(options.primaryUserInfo.hostname, "172.17.0.8");
        strcpy(options.primaryUserInfo.cafile, "./test/ca.crt");
        options.primaryUserInfo.nPort = 8883;
        options.primaryUserInfo.nAuthenicatinMode = MOSQUITTO_AUTHENTICATION_USER | MOSQUITTO_AUTHENTICATION_ONEWAY_SSL;
        instance = MosquittoCreateInstance(&options);
        sleep(3);
        MosquittoSubscribe(instance, NULL, "sensor/room1/#");
        for (int i = 0 ; i < 10; ++i) {
            MosquittoPublish(instance, NULL, "sensor/room1/temperature", 10, "test1234456");
            usleep(1000000);
        }
        MosquittoDestroy(instance);
        MosquittoLibCleanup();
        return 1;
}
