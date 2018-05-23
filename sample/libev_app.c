#include <stdio.h>
#include <stdlib.h>
#include <netinet/in.h>
#include <strings.h>
#include <pthread.h>

#include "ev.h"
#include "ev_app.h"

#define QN_EV_FAIL -1
#define QN_EV_SUCCESS 0

#define LISTENCTLPORT 8333
#define BUFFER_SIZE 1024

#define EV_PRT(fmt, args...) \
{\
	printf(fmt, ##args);\
}

#define QN_CMD(x) *(unsigned int *)x
#define QN_CTRL_LOCK_OPEN  0x0001
#define QN_CTRL_LED_OPEN   0x0002
/**
 * 检测结果接收线程
 */
void *ProcessMsg(void *args)
{
	unsigned int cmd = 0;

	if (NULL != args)
	{
		switch ( QN_CMD(args) )
		{
			case QN_CTRL_LOCK_OPEN :
				{
					printf("========open lock========\n");
				}
				break;

			case QN_CTRL_LED_OPEN :
				{
				}
				break;

			default :
				{
				}
				break;
		}
		return (void*)QN_EV_SUCCESS;
	}
	else 
	{
		return (void *)QN_EV_FAIL;
	}
}
void ev_ReadVal(struct ev_loop *loop, struct ev_io *watcher, int revents)
{
	char cmd_buf[BUFFER_SIZE] = {};
	size_t readN;
	pthread_t *processThr;
	pthread_attr_t thrAttr;

	if (EV_ERROR & revents)
	{
		EV_PRT("error event in read\n");
		return ;
	}

	readN = recv(watcher->fd, cmd_buf,BUFFER_SIZE, 0);
	if (readN < 0)
	{
		EV_PRT("recv error.\n");
		return;
	}
	if (readN == 0)
	{
		EV_PRT("distconnected...\n");
		return;
	}
	bzero(cmd_buf, readN);
	
	//创建一个派发线程，做一次解析，派发
	pthread_attr_init(&thrAttr);
	pthread_attr_setdetachstate(&thrAttr, PTHREAD_CREATE_JOINABLE);
	if (0!= pthread_create(processThr, &thrAttr, ProcessMsg, (void *)cmd_buf))
	{
		EV_PRT("create process data thread failed.\n");
		return;
	}
	pthread_attr_destroy(&thrAttr);

	return;
}
void ev_ProcessAccept(struct ev_loop *loop, struct ev_io *watcher, int revents)
{
	struct sockaddr_in adrClient;
	socklen_t clientLen = sizeof(adrClient);
	int sdClient;

	struct ev_io wClient;
	if (EV_ERROR & revents)
	{
		EV_PRT("error event in accept\n");
		return;
	}
	sdClient = accept(watcher->fd, (struct sockaddr *)&adrClient, &clientLen);
	if (sdClient < 0)
	{
		EV_PRT("accept error\n");
		return;
	}
	ev_io_init(&wClient, ev_ReadVal, sdClient, EV_READ);
	ev_io_start(loop, &wClient);
}

int ev_ListenCtlSock()
{
	int sd;
	struct sockaddr_in addr;

	if ((sd = socket(AF_INET, SOCK_STREAM, 0) < 0))
	{
		EV_PRT("socket failed.\n");
		return QN_EV_SUCCESS;
	}
	bzero(&addr, sizeof(addr));
	addr.sin_family = AF_INET;
	addr.sin_port = htons(LISTENCTLPORT);
	addr.sin_addr.s_addr = INADDR_ANY;
	if (bind(sd, (struct sockaddr*)&addr, sizeof(addr)) != 0) 
	{
		EV_PRT("bind error\n");
		return QN_EV_FAIL;
	}
	if (listen(sd, 2) < 0)
	{
		EV_PRT("listen error.\n");
		return QN_EV_FAIL;
	}

	struct ev_loop *loop;
	struct ev_io sWatcher;
	//struct ev_timer tWatcher;

	loop = ev_default_loop(0);
	ev_io_init(&sWatcher, ev_ProcessAccept, sd, EV_READ);
	ev_io_start(loop, &sWatcher);

	while (1)
	{
		//ev_loop(loop, 0);
		ev_run(loop, 0);
	}

	return QN_EV_SUCCESS;
}
