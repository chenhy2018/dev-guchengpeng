
ARCH:=x86
HOST:=
ifeq ($(ARCH), x86)
	CC=gcc
else ifeq ($(ARCH), mstar)
	CC=arm-none-linux-gnueabihf-gcc
	HOST=arm-none-linux-gnueabihf
else ifeq ($(ARCH), A12)
	CC=arm-none-linux-gnueabi-gcc
	HOST=arm-none-linux-gnueabi
else
	CC=gcc
endif

LIBS=libs
PROJECT_PATH=../../third_party
TARGET_LIBRARY=$(LIBS)/$(ARCH)/libua.a
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=
PJ_CONFIG=--disable-sdl --disable-ffmpeg --disable-v4l2 \
		  --disable-openh264 --disable-libwebrtc --disable-sound
INCLUDE_PATH=-I$(PJSIP_PATH)/pjlib/include \
			 -I$(PJSIP_PATH)/pjlib-util/include \
			 -I$(PJSIP_PATH)/pjsip/include \
			 -I$(PJSIP_PATH)/pjmedia/include \
			 -I$(PJSIP_PATH)/pjnath/include 

OBJS=sip.o sdk_interface.o queue.o framework.o test.o dbg.o mock_api.o

pjlib :
	cd $(PJSIP_PATH); \
	if [ "$(HOST)" = "" ]; then \
		./configure $(PJ_CONFIG); \
	else \
		./configure --host=$(HOST) $(PJ_CONFIG); \
	fi \
	&& make dep \
	&& make 

$(TARGET_LIBRARY) : pjlib

.c.o:
	$(CC) $(INCLUDE_PATH) -c $(HOST)


test:$(OBJS)
	$(CC) -o $@ $^ $(LDFLAGS)

clean:
	rm *.o 


