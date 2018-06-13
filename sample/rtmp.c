
#include <stdlib.h>
#include <stdio.h>
#include <fcntl.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <unistd.h>
#include <sys/mman.h>
#include <string.h>
#include <signal.h>
#include <sys/time.h>

#include "rtmp_publish.h"
#include "curl_req.h"
#include "devsdk.h"
#include "rtmp.h"

#define QN_SUCCESS 0
#define QN_FAIL 1

#define SER_IP  "127.0.0.1"
#define SER_PORT 5557
#define QN_IPC_GET_URL  0x0001                 

#define NAL_SLICE 1             //非关键帧
#define NAL_SLICE_IDR  5        //关键帧
#define NAL_SPS  7              //SPS帧
#define NAL_PPS  8              //PPS帧

#define QU_LEN 128  //视频缓存队列大小
#define PTHREAD_STACK (20<< 20)


typedef int (*RTMPH264SEND_CALLBACK)(char *addrStream, int textLen, unsigned long timeStamp, int iskey);
typedef int (*RTMPAUDIOSEND_CALLBACK)(char *addrAAC, int textLen, double timestamp, unsigned int audioType);

typedef struct _pagAppContext
{
	RtmpPubContext *pRtmpc;

	int status;
	int is_ok;
	int video_state;
	int audio_state;
	pthread_mutex_t push_lock;
	RTMPH264SEND_CALLBACK RtmpH264Send;
	RTMPAUDIOSEND_CALLBACK RtmpAudioSend;
}ACONTEXT;

typedef struct
{
	int size;
	char *data;
} Nalu;

typedef struct{
	int len;
	char *data;
	unsigned long timestamp;
	int iskey;
} Qub;

typedef struct 
{
	Qub qub[QU_LEN];
	pthread_mutex_t qu_lock;
	int poPtr; 
	int puPtr;
} Qu;

int naluAlloc(Nalu *nalu, const char* data, int len);
void naluFree(Nalu *nalu);
int naluCopy(Nalu *nalu, const char* data, int len);
	
Nalu* parseNalu(const char* start, int len);
void destroyNalu(Nalu** nalu);


int qubAlloc(Qub *qub, const char *data, int len);
void qubFree(Qub *qub);
int initQu(Qu *qu);
int pushQu(Qu *qu, const Qub qub);

typedef struct _AdtsUint
{
	unsigned int size;   //数据大小
	unsigned char *data;  //数据
}Adts;

int adtsAlloc(Adts *adts, const char *data, int len);   //成员分配空间
void adtsFree(Adts *adts); 
Adts* parseAdts(char *start,int len);


int RTMPH264Send(char *addrStream, int textLen, unsigned long timeStamp, int iskey);
int RTMPAudioSend(char *addrAAC, int textLen, double timestamp, unsigned int audioType);


static int g_stop_all = 0;
ACONTEXT g_AContext;

Qu g_quV = {};   //视频缓存
Qu g_quA = {};

#define NUM_LINE 8
int g_fp = -1;    //测试文件句柄 

void print_pck(unsigned char *addr, int len)
{
	int i = 0;
	char value[6] = {0};
	char str[512] = {0};
	
	printf("========== memory <%p+%lu> ==========", addr, len);
 	for (; i < len; ++i) {
		   if (i % NUM_LINE == 0) {
				   printf("\n");
				   printf("<%p>:", addr + i - NUM_LINE);
		   }

		   printf("0x%02x\t", addr[i]);
   }
	printf("\n======== memory <%p+%lu> end ========\n", addr, len);
}

int qubAlloc(Qub *qub, const char *data, int len)
{
	if (NULL == qub || data == NULL)
	{
		return QN_FAIL;
	}

	qub->data = (char *)malloc(len);
	if (NULL == qub->data)
	{
		return QN_FAIL;
	}

	memcpy(qub->data, data, len);
	qub->len = len;

	return QN_SUCCESS;
}

void qubFree(Qub *qub)
{
	if (NULL != qub)
	{
		if (NULL != qub->data)
		{
			free(qub->data);
			qub->data = NULL;
			qub->len = -1;
		}

	}
}

int initQu(Qu *qu)
{
	if (qu == NULL) {
		return QN_FAIL;
	}

	qu->poPtr = 0;
	qu->puPtr = 0;
	pthread_mutex_init(&qu->qu_lock, NULL);

	return QN_SUCCESS;
}

int pushQu(Qu *qu, const Qub qub)
{
	if (qu == NULL) {
			return QN_FAIL;
		}
		
	int ret = QN_FAIL;
	
	pthread_mutex_lock(&qu->qu_lock);
		
	//检查状态
	//printf("puPtr:%d poPtr:%d\n", qu->puPtr, qu->poPtr);
	if ((qu->puPtr + 1) % QU_LEN == qu->poPtr) {
		printf("queue is full.\n");
		goto PUSHQU_UNLOCK;
	}
	//插入
	ret = qubAlloc(&qu->qub[qu->puPtr], qub.data, qub.len);
	if (ret == QN_FAIL) {
		printf("push queue fail\n");
		goto PUSHQU_UNLOCK;
	}
	//
	qu->qub[qu->puPtr].timestamp = qub.timestamp;
	qu->qub[qu->puPtr].iskey = qub.iskey;
	//修改指针
	qu->puPtr = (qu->puPtr + 1) % QU_LEN;
	ret = QN_SUCCESS;
		
PUSHQU_UNLOCK: 
		pthread_mutex_unlock(&qu->qu_lock);
		return ret;

}


Qub* popQu(Qu *qu)
{
	if (qu == NULL) {
		return NULL;
	}

	Qub * ret = NULL;

	pthread_mutex_lock(&qu->qu_lock);
	//状态
	if (qu->poPtr == qu->puPtr)
	{
		//printf("[%d]%s >>empty queue\n", __LINE__, __func__);
		goto POPQU_UNLOCK;
	}
	//出
	Qub *qub = NULL;
	qub = (Qub*)malloc(sizeof(Qub));
	if (qubAlloc(qub, qu->qub[qu->poPtr].data, qu->qub[qu->poPtr].len) == QN_FAIL)
	{
		printf("qubAlloc failed.\n");
		goto POPQU_UNLOCK;
	}
	qub->len = qu->qub[qu->poPtr].len;
	//
	qub->timestamp = qu->qub[qu->poPtr].timestamp;
	qub->iskey = qu->qub[qu->poPtr].iskey;
	//指针
	qu->poPtr = (qu->poPtr + 1) % QU_LEN; 
	ret = qub;
	
POPQU_UNLOCK:	
	pthread_mutex_unlock(&qu->qu_lock);
	return ret;
}

void showQu(Qu *qu)
{
	int ptr = qu->poPtr;
	char *tmp = NULL;
	int num = 0;
	
	printf("poPtr:%d   puPtr:%d\n", ptr, qu->puPtr);
	while (ptr != qu->puPtr) {
		num = qu->qub[ptr].len;
		tmp = (char *)malloc(qu->qub[ptr].len + 1);
		memcpy(tmp, qu->qub[ptr].data, qu->qub[ptr].len);
		tmp[num] = '\0';
		printf("%s\t", tmp);
		ptr = (ptr +1) % QU_LEN;
	}
	printf("\n");
}

//rtmp发送线程
static void *pop_pthread(void *arg)
{
	Qub *body = NULL;

	while (1) {
		body = popQu(&g_quV);
		if (NULL == body) {
			continue;
		}
		
		RTMPH264Send(body->data, body->len, body->timestamp, body->iskey);
		qubFree(body);
	}

}

int naluAlloc(Nalu *nalu, const char* data, int len)
{
	if (NULL == nalu || NULL == data)
	{
		return QN_FAIL;
	}

	nalu->data = (char *)malloc(len);
	if (NULL == nalu->data)
	{
		return QN_FAIL;
	}
	
	memcpy(nalu->data, data, len);
	nalu->size = len;
	
	return QN_SUCCESS;
}

void naluFree(Nalu *nalu)
{
	if (NULL != nalu)
	{
		if (NULL != nalu->data)
		{
			free(nalu->data);
			nalu->data = NULL;
			nalu->size = -1;
		}
		
	}
}

int naluCopy(Nalu *nalu, const char* data, int len)
{
	if (nalu == NULL || data == NULL)
	{
		return QN_FAIL;
	}
	if (nalu->size < len)
	{
		return QN_FAIL;
	}

	memcpy(nalu->data, data, len);

	return QN_SUCCESS;
}

Nalu* parseNalu(const    char* start, int len)
{
	if (len <= 0) {
		printf("<parseNalu:%d> len < o\n", __LINE__, len);
		return NULL;
	}
	if (NULL == start)
	{	printf("<parseNalu:%d> len < o\n", __LINE__, len);
		return NULL;
	}

	const char* pStart = start;
	const char *pEnd = NULL;
	Nalu *nalu = NULL;

	//////
	while (pStart < start + len)
	{
		// 判断临界点
		if (pStart >= start + len - 4) {
			pStart = start + len;
			break;
		}

		if (pStart[0] == 0x00 
			&& pStart[1] == 0x00
			&& pStart[2] == 0x00
			&& pStart[3] == 0x01) {

			pStart = pStart + 4;
			pEnd = pStart;
			while (pEnd < start + len )
			{
				if (pEnd[0] == 0x00
					&& pEnd[1]== 0x00
					&& pEnd[2] == 0x00
					&& pEnd[3] == 0x01)
				{
					goto NEXT;
				}
				else 
				{
					pEnd = pEnd + 1;
				}
			}
			if (pEnd >= start + len)
			{
				break;
			}
		} else {
				pStart = pStart + 1;
		}
	}
	
	//////
NEXT:
	if (pStart == start + len) {
		printf("<parseNalu:%d> pStart:%p start+len:%p\n", __LINE__, pStart, start + len);
		return NULL;
	}

	nalu = malloc(sizeof(Nalu));
	if (nalu == NULL) {
		printf("line :%d\n", __LINE__);
		return NULL;
	}
	naluAlloc(nalu, pStart, (int)(pEnd - pStart));

	return nalu;
}

void destroyNalu(Nalu** nalu)
{
	naluFree(*nalu);
	free(*nalu);
	*nalu = NULL;
}

//data包括adts头
int adtsAlloc(Adts *adts, const char *data, int len)
{
	if (adts == NULL || data == NULL) {
		printf("null pointer\n");
		return QN_FAIL;
	}

	adts->data = (char *)malloc(len);
	if (adts->data == NULL) {
		printf("<adtsAlloc>:%d get memory for adts->data failed.\n", __LINE__);
		return QN_FAIL;
	}
	memcpy(adts->data, data, len);
	adts->size = len;

	return QN_SUCCESS;
}

void adtsFree(Adts *adts)
{
	if (adts != NULL) {
		if (adts->data != NULL) {
			free(adts->data);
			adts->data = NULL;
			adts->size = -1;
		}
	}
}

Adts* parseAdts(char *start,int len)
{
	int ret = QN_FAIL;
	
	if (start == NULL) {
		printf("<parseAdts>:%d null pointer\n", __LINE__);
		return NULL;
	}

	int headLen = 0;              //aac头长度
	const char *pStart = start;   //遍历首指针
	const char *pEnd = NULL;      //遍历结尾指针
	
	while (pStart < start + len) {

		if (pStart + 2 >= start + len) {
			break;
		}

		//一个adts开始
		if (pStart[0] == 0xff
			&& (pStart[1] & 0xf0) == 0xf0) {
			//printf("<parseAdts>:%d find start\n", __LINE__);	
			headLen = (pStart[1] & 0x1) == 1 ? 7 : 9; 
			break;
		} else {
			
			pStart++;
		}
	}

ADTS_NEXT:
	if (pStart >= start + len -2) {
		//printf("<parseAdts>:%d pStart >= start + len\n", __LINE__);
		return NULL;
	}
	
	int frameLen = 0;         
	int aacFrameLen = 0;     //aac帧长度

	//取adts header的aac_frame_length,即aac帧长度
	aacFrameLen |= (pStart[3] & 0x3) << 11;
	aacFrameLen |= pStart[4] << 3;
	aacFrameLen |= (pStart[5] & 0xe0) >> 5;
	frameLen = (int)(pEnd - pStart);
	//分配adts空间
	Adts *adts = (Adts *)malloc(sizeof(Adts));
	if (adts == NULL) {
		printf("<parseAdts>:%d malloc failed.\n", __LINE__);
		return NULL;
	}
	ret = adtsAlloc(adts, start, aacFrameLen);
	if (ret == QN_FAIL) {

		free(adts);
		adts = NULL;
		return NULL;
	}
	//printf("<parseAdts>:%d aacFrameLen:%d\n", __LINE__, aacFrameLen);
	return adts;
}

int checkAudioEnable()
{
	AudioConfig audioConfig;

	dev_sdk_get_AudioConfig(&audioConfig);

	if( audioConfig.audioEncode.enable != 1)
		return 0;
	else
		return 1 ;
}

int video_callback(char *frame, int len, int iskey, double timestamp, unsigned long frame_index,	unsigned long keyframe_index, void *pcontext)
{
	static unsigned int nLastTimeStamp = 0;
	int *pStreamNo = (int*)pcontext;
	int streamno = *pStreamNo;
	int nDiff = (unsigned int)timestamp - nLastTimeStamp;	

	nLastTimeStamp = (unsigned int)timestamp;

#if 1
	if (g_AContext.RtmpH264Send == NULL)
	{
		printf("<video_callback>: g_AContext.RtmpH264Send is NULL.\n");
		return QN_FAIL;
	}
	g_AContext.RtmpH264Send(frame, len, timestamp, iskey);
#endif
	
#if 0
	Qub qub = {};
	qubAlloc(&qub, frame, len);
	qub.timestamp = timestamp;
	qub.iskey = iskey;
	if (pushQu(&g_quV, qub) == QN_FAIL) {
		printf("push video data\n");
		qubFree(&qub);
		return QN_FAIL;
	}
	qubFree(&qub);
#endif
	return QN_SUCCESS;
}


/**
 * timestamp :
 * frame_index :
 * pcontext :上下文
 */
int audio_callback(char *frame, int len, double timestamp, unsigned long frame_index, void *pcontext)
{
	g_AContext.RtmpAudioSend(frame, len, timestamp, RTMP_PUB_AUDIO_AAC);

	//write(g_fp, frame, len);
	return 0;
}

void aj_Init()
{
	static int context = 1;
	int s32Ret = 0;

	s32Ret = dev_sdk_init(DEV_SDK_PROCESS_APP);
	dev_sdk_start_video(0, 0, video_callback, &context);
	if (checkAudioEnable())
	{
		dev_sdk_start_audio_play(AUDIO_TYPE_AAC);
		dev_sdk_start_audio(0, 1, audio_callback, NULL);
	}

	return;
}
int aj_unInit()
{
	dev_sdk_stop_video(0, 1);
	dev_sdk_stop_audio(0, 1);
	dev_sdk_stop_audio_play();
	dev_sdk_release();
}

int rtmp_init(ACONTEXT *pAContext)
{	
#if 0
	g_fp = open ("audioTest.aac", O_RDWR | O_APPEND);
	if (g_fp < 0)
	{
		printf("<naluAlloc>: open audioTest.g711a failed.\n");
		return QN_FAIL;
	}
#endif

	g_AContext.video_state = FALSE;
	g_AContext.audio_state = FALSE;
	g_AContext.is_ok = FALSE;
	
	int ret = QN_FAIL;
	char RtmpUrl[256] = {};

	//while (0 != GetUrlRtmp(RtmpUrl));
	memset(RtmpUrl, 0, 256);
	memcpy(RtmpUrl, "rtmp://pili-publish.caster.test.cloudvdn.com/caster-test/test1", 256);
	printf("rtmp url :%s\n", RtmpUrl);

	pthread_mutex_init(&pAContext->push_lock, NULL);

	//pAContext->pRtmpc = RtmpPubNew(RtmpUrl, 10, RTMP_PUB_AUDIO_NONE, RTMP_PUB_AUDIO_NONE, RTMP_PUB_TIMESTAMP_ABSOLUTE);
	pAContext->pRtmpc = RtmpPubNew(RtmpUrl, 10, RTMP_PUB_AUDIO_AAC, RTMP_PUB_AUDIO_AAC, RTMP_PUB_TIMESTAMP_ABSOLUTE);
	if (NULL == pAContext->pRtmpc)
	{
		printf("rtmp_init line:%d.\n", __LINE__);
		return -1;
	}

	ret = RtmpPubInit(pAContext->pRtmpc);
	if (QN_SUCCESS != ret)
	{
		RtmpPubDel((RtmpPubContext *)pAContext->pRtmpc);
		pAContext->pRtmpc = NULL;
		return ret;
	}
	ret = RtmpPubConnect(pAContext->pRtmpc);
	if (QN_SUCCESS != ret)
	{
		RtmpPubDel((RtmpPubContext *)pAContext->pRtmpc);
		pAContext->pRtmpc = NULL;
		return ret;
	}

	pAContext->RtmpH264Send = (RTMPH264SEND_CALLBACK)RTMPH264Send;
	pAContext->RtmpAudioSend = (RTMPAUDIOSEND_CALLBACK)RTMPAudioSend;

#if 0
	//发送线程
	pthread_attr_t attr;
	pthread_t popPid = {0};
	ret = pthread_attr_init(&attr);
	if (ret != 0) {
		printf("pthread_attr_init failed.\n");
		return -1;
	}
	ret = pthread_attr_setstacksize(&attr, PTHREAD_STACK);
	if (ret != 0) {
		printf("pthread_attr_setstacksize failed.\n");
		return -1;
	}

	if (pthread_create(&popPid, &attr, pop_pthread, NULL)) {
		printf("pthread_create push_pthread failed.\n");
		return -1;
	}
#endif

	return QN_SUCCESS;
}
void rtmp_uninit()
{
	RtmpPubDel((RtmpPubContext *)g_AContext.pRtmpc);
    aj_unInit();
	return;
}
//解析adts固定头部
int getAdtsFHL(const Adts *adts, int *headLen)
{
	int ret = QN_FAIL;

	if (adts == NULL || headLen == NULL) {
		return ret;
	}

	if ((adts->data[1] & 0x1) == 1) {
		*headLen = 7;
	} else {
		*headLen = 9;
	}
	
	return QN_SUCCESS;
}

void putC(unsigned char *buf, int len)
{
	int i = 0;

	printf("===============================================\n");
	printf("<%p>:", &buf[i]);
	for (i = 0; i < len; i++)
	{
		if (i % 16 == 0 && i != 0)
		{
			printf("\n");
			printf("<%p>:", &buf[i]);
		}
		printf("%x\t", buf[i]);
	}
	printf("===============================================\n");
	printf("\n");
	return;
}
void getNals(const char *stream, int len, int *bufSize)
{
	int uIdx = 0;
	int num = 0;
	Nalu *nalU = NULL;
	
	while (uIdx < len)
	{
		printf("uIdx :%d   len:%d\n", uIdx, len);
		//putC(stream + uIdx, len - uIdx);
		printf("=================================================\n");
		nalU = parseNalu(stream + uIdx, len - uIdx);
		if (NULL == nalU)
		{
			return;
		}
		uIdx += nalU->size + 4;
		destroyNalu(&nalU);
		
	}
	*bufSize = uIdx;
	
	return;
}

int RTMPH264Send(char *addrStream, int textLen, unsigned long timeStamp, int iskey)
{
	int s32Ret = QN_FAIL;
	long long int presentationTime = (long long int)timeStamp;
    
	Nalu *nalU = NULL;
	int uIdx = 0;
	RtmpPubContext *pRtmpc = g_AContext.pRtmpc;

	if (NULL == addrStream || pRtmpc == NULL)
	{
		printf("<RTMPH264Send>: pointer is NULL.\n");
		return s32Ret;
	}

	if (g_AContext.status != RTMP_START)
	{
		return QN_SUCCESS;
	}
	
	if (iskey && g_AContext.video_state == FALSE)
	{
		// 设置视频时间基
		RtmpPubSetVideoTimebase(pRtmpc, presentationTime);

		// 解析sps
		nalU = parseNalu(addrStream, textLen);
		if (NULL == nalU) 
		{
			return QN_FAIL;
		}
		RtmpPubSetSps(pRtmpc, nalU->data, nalU->size); 
		uIdx += nalU->size + 4;
		destroyNalu(&nalU);

		// 解析pps
		nalU = parseNalu(addrStream + uIdx, textLen - uIdx );
		if (NULL == nalU)
		{
			return QN_FAIL;
		}
		RtmpPubSetPps(pRtmpc, nalU->data, nalU->size); 
		uIdx += nalU->size +4;
		destroyNalu(&nalU);

		g_AContext.video_state = TRUE;
		g_AContext.is_ok = TRUE;
	}


	// 解析数据包 annexB => NALU
	const int maxBufSize = 1024 * 1024 * 10;
	char *buf = (char *)malloc(maxBufSize);
	if (buf == NULL) {
		printf("<RTMPH264Send>: malloc failed for buf");
		free(buf);
		return s32Ret;
	}
	int bufSize = 0;
	while (uIdx < textLen) {
		// 获取一个nalu
		nalU = parseNalu(addrStream + uIdx, textLen - uIdx);
		if (nalU == NULL) {
			break;
		}
		
		// payload长度不能大于buf的堆内存大小
		if (bufSize + nalU->size > maxBufSize) {
			printf("<RTMPH264Send>: malloc failed for buf");
			destroyNalu(&nalU);
			free(buf);
			return s32Ret;
		}

		// 只拷贝keyframe和interframe
		if (((nalU->data[0] & 0x0f) == 0x01) || ((nalU->data[0] & 0x0f) == 0x05)) {
			// nalu payload
			memcpy(&buf[bufSize], nalU->data, nalU->size);
			bufSize += nalU->size;
		}
	
		// 移动源buffer指针
		uIdx += nalU->size + 4;
		destroyNalu(&nalU);
	}

	// 不是合法的视频帧，直接返回
	if (bufSize == 0) {
		goto QUIT_RTMPH264SEND;
	}

	#if 1
	pthread_mutex_lock(&g_AContext.push_lock);
	//printf("iskey :%d  timestamp:%ld\n", iskey, timeStamp);
	if (iskey)
	{
		s32Ret = RtmpPubSendVideoKeyframe(pRtmpc, buf, bufSize, presentationTime);
		if (s32Ret != QN_SUCCESS)
		{
			printf("Send video key frame fail.\n");
		}
	}
	else
	{
		RtmpPubSendVideoInterframe(pRtmpc, buf, bufSize, presentationTime);
	}
	pthread_mutex_unlock(&g_AContext.push_lock);
    #endif
	
QUIT_RTMPH264SEND:
	free(buf);
	return QN_SUCCESS;	
}

int RTMPAudioSend(char *addrAAC, int textLen, double timestamp, unsigned int audioType)
{
	//从编码器获取音频编码信息
	long long int presentationTime = (long long int)timestamp;
	RtmpPubContext *pRtmpc = g_AContext.pRtmpc;
	static unsigned char audioSpecCfg[] = {0x14, 0x10};   ///{0x14, 0x08};
	int ret = QN_FAIL;

	if (g_AContext.is_ok == FALSE) {
		//printf("<RTMPAudioSend>: wait for video metadata send\n");
		usleep(5000);
		return QN_SUCCESS;
	}

	if (g_AContext.audio_state == FALSE)
	{                                   
		RtmpPubSetAudioTimebase(pRtmpc, presentationTime);
		//识别音频类型
		if (audioType == AUDIO_TYPE_AAC)
		{
			//设置AAC配置
			RtmpPubSetAac(pRtmpc, audioSpecCfg, sizeof(audioSpecCfg));
		}
		
		g_AContext.audio_state = TRUE;
	}
	
	// ADTS Frame = ADTS Header ：AAC ES
	int uIdx = 0;
	int bufSize = 0;
	int headLen = 0;
	Adts *adts = NULL;
	char *buf = NULL;

	buf = (char *)malloc(textLen);
	if (buf == NULL) {
		printf("malloc buffer failed.\n");
		return ret;
	}

	while (uIdx < textLen) {
		//解析一个adts 
		adts = parseAdts(addrAAC + uIdx, textLen - uIdx);
		if (adts == NULL) {
			break;
		}
		//adts header 为7bytes
		if (QN_FAIL == getAdtsFHL(adts, &headLen)) {
			break;
		}
		memcpy(&buf[bufSize], &adts->data[headLen], adts->size - headLen);
		bufSize += adts->size - headLen;
		uIdx = uIdx + adts->size;
	}
	//
	#if 1
	pthread_mutex_lock(&g_AContext.push_lock);
	{
		RtmpPubSendAudioFrame(pRtmpc, buf, bufSize, presentationTime);
	}
	pthread_mutex_unlock(&g_AContext.push_lock);
	#endif
	free(buf);
	buf = NULL;
	
	return QN_SUCCESS;
}
//初始化信号处理函数
//ctrl+c时进行清理
void sig_exit(int sig)
{
	//清理aj sdk
	aj_unInit();
	//清理rtmp
	rtmp_uninit();
	g_stop_all = 1;
}
void init_signals(void)
{
	struct sigaction sa = {};

	sa.sa_flags = 0;
	sigemptyset(&sa.sa_mask);
	sigaddset(&sa.sa_mask, SIGTERM);
	sigaddset(&sa.sa_mask, SIGINT);

	sa.sa_handler = sig_exit;
	sigaction(SIGTERM, &sa, NULL);
	sa.sa_handler = sig_exit;
	sigaction(SIGINT, &sa, NULL);

	signal(SIGPIPE, SIG_IGN);
	
	return;
}

void RTMPStat(int _status)
{
    int ret = 0;
	g_AContext.status = _status;

	if (_status == RTMP_START) {
		g_AContext.is_ok = FALSE;
		g_AContext.video_state = FALSE;
		g_AContext.audio_state = FALSE;
        ret = rtmp_init(&g_AContext);
        if (0 != ret)
        {
            printf("Open Rtmp Client fail.\n");
        }
        aj_Init();
	} else {
        rtmp_uninit(&g_AContext);
    }
}

void Rtmp_Init()
{
	int  ret = QN_FAIL;

	//初始化信号处理
	init_signals();	

#if 0
	//初始化缓冲队列
	if (initQu(&g_quV) == QN_FAIL) {
		printf("init queue failed.\n");
		return QN_FAIL;
	}
#endif

    /*ret = rtmp_init(&g_AContext);*/
    /*if (0 != ret)*/
    /*{*/
        /*printf("Open Rtmp Client fail.\n");*/
        /*return -1;*/
    /*}*/
	
	/*aj_Init();*/
}

