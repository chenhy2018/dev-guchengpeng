
all:queue-test

queue-test:queue_test.o queue.o
	gcc queue_test.o  queue.o -o queue-test

queue_test.o:queue_test.c
	gcc -c queue_test.c -o queue_test.o
queue.o:../queue.c
	gcc -c ../queue.c -o queue.o
clean:
	rm *.o queue-test
