#include "mosquitto.h"
#include <unistd.h>
#include <string.h>
#include <stdio.h>

void onMessage(IN const void* instance, IN const char* message, IN size_t length)
{
  fprintf(stderr, "message %s", message);
}

int main()
{
        struct MosquittoOptions options;
        strcpy(options.id, "test");
        options.bCleanSession = false;
        strcpy(options.primaryUserInfo.hostname, "172.17.0.2");
        //strcpy(options.bindaddress, "172.17.0.2");
        strcpy(options.secondaryUserInfo.hostname, "172.17.0.4");
        //strcpy(options.secondBindaddress, "172.17.0.2`");
        strcpy(options.primaryUserInfo.username, "testaaa");
        strcpy(options.primaryUserInfo.password, "111");
        strcpy(options.secondaryUserInfo.username, "test");
        strcpy(options.secondaryUserInfo.password, "111");
        options.secondaryUserInfo.nPort = 1883;
        options.primaryUserInfo.nPort = 1883;
        options.nKeepalive = 10;
        options.nQos = 0;
        options.bRetain = false;
        options.callbacks.onMessage = &onMessage;
        void* instance = MosquittoCreateInstance(&options);
        sleep(3);
        MosquittoSubscribe(instance, NULL, "sensor/room1/#");
        do {
            MosquittoPublish(instance, NULL, "sensor/room1/temperature", 10, "test1234456");
            usleep(1000000);
        } while(1);
        return 1;
}
