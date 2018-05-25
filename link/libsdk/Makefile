
CC=arm-linux-gnueabihf-gcc
PROJECT_PATH=../third_party
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=-L $(PJSIP_PATH)/pjmedia/lib -L $(PJSIP_PATH)/pjsip/lib -L $(PJSIP_PATH)/pjnath/lib -L $(PJSIP_PATH)/pjlib-util/lib \
		-L $(PJSIP_PATH)/pjlib/lib -L./lib \
		-lpjmedia-arm-unknown-linux-gnueabihf -lpjmedia-codec-arm-unknown-linux-gnueabihf \
		-lpjmedia-audiodev-arm-unknown-linux-gnueabihf -lpjsip-arm-unknown-linux-gnueabihf \
		-lpjsdp-arm-unknown-linux-gnueabihf -lpj-arm-unknown-linux-gnueabihf \
		-lpjlib-util-arm-unknown-linux-gnueabihf -lpjsua-arm-unknown-linux-gnueabihf \
		-lpjsip-ua-arm-unknown-linux-gnueabihf -lpthread -lm \
		-lcurl -lrtmp_sdk -lev -ldevsdk -ltools -lstdc++ -lrtmp -lfdk-aac
INCLUDE_PATH=-I../third_party/pjproject-2.7.2/pjlib/include -I../third_party/pjproject-2.7.2/pjlib-util/include -I../third_party/pjproject-2.7.2/pjsip/include -I../third_party/pjproject-2.7.2/pjmedia/include -I../third_party/pjproject-2.7.2/pjnath/include \

OBJS=sip.o main.o libev_app.o curl_req.o rtmp.o

.c.o:
	$(CC) $(INCLUDE_PATH) -c $<


sip-demo:$(OBJS)
	$(CC) -o $@ $^ $(LDFLAGS)

clean:
	rm *.o sip-demo


