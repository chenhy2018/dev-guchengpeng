ARCH:=x86
HOST:=
ifeq ($(ARCH), x86)
	CC=gcc
else ifeq ($(ARCH), mstar)
	CC=arm-linux-gnueabihf-gcc
	HOST=arm-linux-gnueabihf
else ifeq ($(ARCH), A12)
	CC=arm-none-linux-gnueabi-gcc
	HOST=arm-none-linux-gnueabi
else
	CC=gcc
endif

CURRENT_DIR:=$(shell pwd)
PROJECT_PATH=../../third_party
PJSIP_PATH=$(PROJECT_PATH)/pjproject-2.7.2
LDFLAGS=
INSTALL_DIR=$(CURRENT_DIR)/output/pjsip
PJ_CONFIG=--disable-sdl --disable-ffmpeg --disable-v4l2 \
		  --disable-openh264 --disable-libwebrtc --disable-sound \
		  --disable-libyuv \
		  --prefix=$(INSTALL_DIR)
INCLUDE_PATH=-I$(PJSIP_PATH)/pjlib/include \
			 -I$(PJSIP_PATH)/pjlib-util/include \
			 -I$(PJSIP_PATH)/pjsip/include \
			 -I$(PJSIP_PATH)/pjmedia/include \
			 -I$(PJSIP_PATH)/pjnath/include \
			 -I../libsip

SIP_PATH=../libsip
OUTPUT=output
SIP_OUT_PATH=$(OUTPUT)/libsip
LIBSIP=$(SIP_OUT_PATH)/sip.o
SDK_OBJS_PATH=$(OUTPUT)/libsdk
LIBSDK=queue.o sdk_interface.o \
       test.o dbg.o mock_api.o uaMgr.o callMgr.o mock_rtc.o
TARGET_LIBRARY=$(OUTPUT)/lib/$(ARCH)/libua.a
LIBPJ=$(OUTPUT)/ori/$(ARCH)
LIBMQTT=
LIBRTP=

all : $(TARGET_LIBRARY)

depclean:
	cd $(PJSIP_PATH); \
	find . -name "*.depend" -print -delete  

pjclean :
	rm -rvf $(OUTPUT)/pjsip
	rm -rvf $(LIBPJ)

pjsip $(LIBPJ) : #pjclean depclean
	@echo "compile pjsip"
	mkdir -p $(INSTALL_DIR)
	cd $(PJSIP_PATH); \
	make distclean && \
	make clean && \
	make realclean && \
	cd $(PJSIP_PATH); \
	if [ "$(HOST)" = "" ]; then \
		./configure $(PJ_CONFIG); \
	else \
		./configure --host=$(HOST) $(PJ_CONFIG); \
	fi \
	&& make dep \
	&& make clean \
	&& make \
	&& make install


$(LIBSIP):$(SIP_PATH)/sip.c
	mkdir -p $(SIP_OUT_PATH)
	$(CC) $(INCLUDE_PATH) -c $< -o $@

# set .o file search path
vpath %.o $(SDK_OBJS_PATH)
sdk $(TARGET_LIBRARY) : $(LIBPJ) $(LIBSIP) $(LIBSDK) $(LIBMQTT) $(LIBRTP)
	./merge-lib.sh $(ARCH)
	make -C test

$(LIBSDK) : %.o : %.c
	mkdir -p $(SDK_OBJS_PATH) 
	$(CC) $(INCLUDE_PATH) -c $< -o $(SDK_OBJS_PATH)/$@
libsdk : $(LIBSDK)

clean:
	rm -rvf output


