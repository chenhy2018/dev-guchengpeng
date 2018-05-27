
CC=arm-linux-gnueabihf-gcc
PROJECT_PATH=../../third_party
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=-L $(PJSIP_PATH)/pjmedia/lib -L $(PJSIP_PATH)/pjsip/lib \
		-L $(PJSIP_PATH)/pjnath/lib -L $(PJSIP_PATH)/pjlib-util/lib \
		-L $(PJSIP_PATH)/pjlib/lib -L./lib \
		-lpjmedia-arm-unknown-linux-gnueabihf \
		-lpjmedia-codec-arm-unknown-linux-gnueabihf \
		-lpjmedia-audiodev-arm-unknown-linux-gnueabihf \
		-lpjsip-arm-unknown-linux-gnueabihf \
		-lpjsdp-arm-unknown-linux-gnueabihf \
		-lpj-arm-unknown-linux-gnueabihf \
		-lpjlib-util-arm-unknown-linux-gnueabihf \
		-lpjsua-arm-unknown-linux-gnueabihf \
		-lpjsip-ua-arm-unknown-linux-gnueabihf -lpthread -lm 

INCLUDE_PATH=-I$(PJSIP_PATH)/pjlib/include \
			 -I$(PJSIP_PATH)/pjlib-util/include \
			 -I$(PJSIP_PATH)/pjsip/include \
			 -I$(PJSIP_PATH)/pjmedia/include \
			 -I$(PJSIP_PATH)/pjnath/include 

OBJS=sip.o sdk_interface.o queue.o framework.o test.o

.c.o:
	$(CC) $(INCLUDE_PATH) -c $<


test:$(OBJS)
	$(CC) -o $@ $^ $(LDFLAGS)

clean:
	rm *.o 


