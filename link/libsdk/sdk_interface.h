// Last Update:2018-05-31 18:34:13
/**
 * @file sdk_interface.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */

#ifndef SDK_INTERFACE_H
#define SDK_INTERFACE_H

typedef enum {
    CALL_STATUS_INCOMING,
    CALL_STATUS_RING_TIMEOUT,
    CALL_STATUS_ESTABLISHED,
    CALL_STATUS_RING,
    CALL_STATUS_REJECT,
    CALL_STATUS_DISCONNECTED,
} CallStatus;

typedef enum {
    EVENT_CALL,
    EVENT_DATA,
    EVENT_MESSAGE,
} EventType;

typedef enum {
    RET_OK,
    RET_FAIL,
    RET_RETRY,
    RET_MEM_ERROR = -1,
    RET_PARAM_ERROR = -2,
    RET_ACCOUNT_NOT_EXIST = -3,
} ErrorID;

typedef enum {
    STREAM_NONE,
    STREAM_AUDIO,
    STREAM_VIDEO,
    STREAM_DATA
} StreamType;

typedef enum {
    CODEC_NONE,
    CODEC_H264,
    CODEC_H265,
    CODEC_G711A,
    CODEC_G711U,
    CODEC_SPEEX,
    CODEC_AAC
} Codec;

typedef int AccountID;

#undef IN
#define IN const

#undef OUT
#define OUT

#define URL_LEN_MAX (128)
#define STREAM_PACKET_LEN (256)
#define AUDIO_CODEC_MAX 16
#define VIDEO_CODEC_MAX 16

typedef struct {
    int callID;
    void *data;
    int size;
    int pts;
    Codec codec;
    Stream stream;
} DataEvent;

typedef struct {
    int callID;
    CallStatus status;
    char From[URL_LEN_MAX];
} CallEvent;

typedef struct {
    char *message;
    char *topic;
    int qos;
} MessageEvent;

typedef struct {
    int eventID;
    int time;
    union {
        CallEvent callEvent;
        DataEvent dataEvent;
        Message messageEvent;
    } body;
} Event;

// library initial, application only need to call one time
Status LibraryInit();
// register a account
AccountID Register( IN char* id, IN char* host, IN char* password, int nDeReg );
ErrorID UnRegister( AccountID id );
// make a call, user need to save call id
ErrorID MakeCall( AccountID id, IN char* pDestUri, OUT int *pCallId );
ErrorID AnswerCall( AccountID id, int nCallId );
ErrorID RejectCall( AccountID id, int nCallId );
// hangup a call
ErrorID HangupCall( AccountID id, int nCallId );
// send a packet
ErrorID SendPacket( AccountID id , int nCallId, int streamIndex, IN char* buffer, int size);
// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID PollEvents( AccountID id, OUT int* eventID, OUT EventData* data, int nTimeOut );
ErrorID AddCodec( AccountID id, Codec codecs[], int size, int samplerate, int channels);

// mqtt report
int Report( int fd, const char* message, size_t length );

#endif  /*SDK_INTERFACE_H*/

