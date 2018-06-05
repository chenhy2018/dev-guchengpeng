

LDFLAGS=-L../output/lib/x86 -lua -lpthread -lm
CFLAGS=-g
TARGET_OBJS=test.o unit_test.o
TESTCASE_OBJS_DIR=../output/test/objs
INCLUDE_PATH=-I../
TARGET=test
CC=gcc

all:$(TARGET)

$(TARGET_OBJS) : %.o : %.c
	mkdir -p $(TESTCASE_OBJS_DIR) 
	$(CC) $(CFLAGS) $(INCLUDE_PATH) -c $< -o $(TESTCASE_OBJS_DIR)/$@

vpath %.o $(TESTCASE_OBJS_DIR)
$(TARGET):$(TARGET_OBJS)
	$(CC) -o $@ $^ $(LDFLAGS)

clean:
	rm -rvf $(TESTCASE_OBJS_DIR)/*.o test

