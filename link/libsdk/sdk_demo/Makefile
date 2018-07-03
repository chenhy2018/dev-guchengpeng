
CC=arm-linux-gnueabihf-gcc
ARCH:=mstar
LDFLAGS=-L./lib -L../output/lib/$(ARCH) -lua -lpthread -lm -lssl -lcrypto  -lcurl -lrtmp_sdk -lev -ldevsdk -ltools -lstdc++ -lfdk-aac
CFLAGS=-g -std=gnu99
TARGET_OBJS= main.o dbg.o stream.o h264_parse.o
OBJS_DIR=./objs
INCLUDE_PATH=-I./inc
TARGET=sdk-demo

all:$(TARGET)

$(TARGET_OBJS) : %.o : %.c
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(OBJS_DIR)/$@

OBJ_FILES = $(addprefix objs/,$(notdir $(TARGET_OBJS)))
vpath %.o $(OBJS_DIR)
$(TARGET):$(TARGET_OBJS)
	$(CC) -o $@ $(OBJ_FILES) $(LDFLAGS)
	cp sdk-demo /home/share/

clean:
	rm -rvf $(OBJS_DIR)/*.o sdk-demo

