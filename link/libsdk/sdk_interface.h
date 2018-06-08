// Last Update:2018-06-08 17:36:15
/**
 * @file sdk_interface.h
 * @brief 
 * @author 
 * @version 0.1.00
 * @date 2018-05-25
 */

#ifndef SDK_INTERFACE_H
#define SDK_INTERFACE_H

#include <stdint.h>

typedef enum {
    CALL_STATUS_IDLE,
    CALL_STATUS_REGISTERED,
    CALL_STATUS_REGISTER_FAIL,
    CALL_STATUS_INCOMING,
    CALL_STATUS_TIMEOUT,
    CALL_STATUS_ESTABLISHED,
    CALL_STATUS_RING,
    CALL_STATUS_REJECT,
    CALL_STATUS_HANGUP,
    CALL_STATUS_ERROR
} CallStatus;

typedef enum {
    EVENT_CALL,
    EVENT_DATA,
    EVENT_MESSAGE,
} EventType;

typedef enum {
    RET_OK,
    RET_FAIL,
    RET_REGISTERING,
    RET_RETRY,
    RET_MEM_ERROR = 1001,
    RET_PARAM_ERROR,
    RET_INIT_ERROR,
    RET_ACCOUNT_NOT_EXIST,
    RET_CALL_NOT_EXIST,
    RET_REGISTER_TIMEOUT,
    RET_TIMEOUT_FROM_SERVER,
    RET_USER_UNAUTHORIZED,
    RET_CALL_INVAILD_CONNECTION,
    RET_CALL_INVAILD_SDP,
    RET_CALL_INVAILD_OPERATING
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
    CODEC_G711A,
    CODEC_G711U,
} Codec;

typedef int AccountID;

#undef IN
#define IN

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
    int status;
} MessageEvent;

typedef struct {
    int eventID;
    int time;
    EventType type;
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
                   const char* mediaHost, const char* imHost);
ErrorID UnRegister( AccountID id );
// make a call, user need to save call id
ErrorID MakeCall(AccountID accountID, const char* id, const char* host, OUT int* callID);
ErrorID AnswerCall( AccountID id, int nCallId );
ErrorID RejectCall( AccountID id, int nCallId );
// hangup a call
ErrorID HangupCall( AccountID id, int nCallId );
// send a packet
ErrorID SendPacket(AccountID id, int callID, Stream streamID, const uint8_t* buffer, int size, int64_t nTimestamp);
// poll a event
// if the EventData have video or audio data
// the user shuould copy the the packet data as soon as possible
ErrorID PollEvent(AccountID id, EventType* type, Event** event, int timeOut );

// mqtt report
ErrorID Report(AccountID id, const char* message, int length);

#endif  /*SDK_INTERFACE_H*/

