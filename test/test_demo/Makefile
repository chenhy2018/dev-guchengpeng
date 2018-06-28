ARCH:=x86
LIBUA=../../output/lib/$(ARCH)/libua.a
LDFLAGS=-L../../output/lib/$(ARCH) -lua -lpthread -lm /usr/lib/x86_64-linux-gnu/libcrypto.so /usr/lib/x86_64-linux-gnu/libssl.so
CFLAGS=-g
TARGET_OBJS=test.o unit_test.o
TESTCASE_OBJS_DIR=./objs
INCLUDE_PATH=-I../
TARGET=test
CC=gcc

all:$(TARGET)


$(TARGET_OBJS) : %.o : %.c
	mkdir -p $(TESTCASE_OBJS_DIR) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(TESTCASE_OBJS_DIR)/$@

OBJ_FILES = $(addprefix objs/,$(notdir $(TARGET_OBJS)))
vpath %.o $(TESTCASE_OBJS_DIR)
$(TARGET):$(TARGET_OBJS) $(LIBUA)
	$(CC) -o $@ $(OBJ_FILES) $(LDFLAGS)

clean:
	rm -rvf $(TESTCASE_OBJS_DIR)/*.o test $(TESTCASE_OBJS_DIR)

