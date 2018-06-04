// Last Update:2018-06-04 14:11:11
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
    CALL_STATUS_TIMEOUT,
    CALL_STATUS_ESTABLISHED,
    CALL_STATUS_RING,
    CALL_STATUS_REJECT,
    CALL_STATUS_HANGUP,
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
    RET_REGISTER_TIMEOUT = -4,
    RET_TIMEOUT_FROM_SERVER = -5,
    RET_USER_UNAUTHORIZED = -6,
} ErrorID;

typedef enum {
    STREAM_NONE,
    STREAM_AUDIO,
    STREAM_VIDEO,
    STREAM_DATA
} Stream;

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
    char *pFromAccount;
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
        MessageEvent messageEvent;
    } body;
} Event;

typedef struct {
    Stream streamType;
    Codec codecType;
    int sampleRate;
    int channels;
} Media;

// library initial, application only need to call one time
ErrorID InitSDK(  Media* mediaConfigs, int size);
ErrorID UninitSDK();
// register a account
// @return if return value > 0, it is the account id, if < 0, it is the ErrorID
AccountID Register(const char* id, const char* password, const char* sigHost,
                   const char* mediaHost, const char* imHost, int deReg, int timeOut );
ErrorID UnRegister( AccountID id );
// make a call, user need to save call id
ErrorID MakeCall(AccountID accountID, const char* id, const char* host, OUT int* callID);
ErrorID AnswerCall( AccountID id, int nCallId );
ErrorID RejectCall( AccountID id, int nCallId );
// hangup a call
ErrorID HangupCall( AccountID id, int nCallId );
// send a packet
ErrorID SendPacket(AccountID id, int callID, Stream streamID, const char* buffer, int size);
// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID PollEvent(AccountID id, EventType* type, Event* event, int timeOut );

// mqtt report
ErrorID Report(AccountID id, const char* topic, const char* message, int length);
ErrorID RegisterTopic(AccountID id, const char* topic);
ErrorID UnregisterTopic(AccountID id, const char* topic);

#endif  /*SDK_INTERFACE_H*/

