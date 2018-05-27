// Last Update:2018-05-27 16:47:14
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
    EVENT_TYPE_NONE,
    EVENT_TYPE_INCOMING_CALL,
// session has been established, could start to transport media stream
    EVENT_TYPE_SESSION_ESTABLISHED,
// session has been established, should stop media stream transport
    EVENT_TYPE_SESSION_FINISHED,
//  received media(audio/video) packet
    EVENT_TYPE_RECEIVE_PACKET,
    EVENT_TYPE_ERROR,
} event_type_e;

typedef enum {
    RET_SUCCESS,
    RET_FAIL,
} status_e;

typedef enum {
    STREAM_TPE_NONE,
    STREAM_TYPE_AUDIO,
    STREAM_TYPE_VIDEO,
} StreamType_e;

#define URL_LEN_MAX (128)
#define STREAM_PACKET_LEN (256)

typedef struct {
    StreamType_e type;
    int samplerate;
    int channels;
    int width;
    int height;
} stream_s;

typedef struct {
    StreamType_e streamType;
    unsigned char packet[STREAM_PACKET_LEN];
} StreamPaket_s;

typedef struct {
    int fd;
    int nAccountId;
    int nCallId;
    union {
        char From[URL_LEN_MAX];
        StreamPaket_s stream;
    } body;
} event_s;

int CreateUA();
int UA_Destroy();
int Register( const char* id, const char* host, const char* password, const int _bDeReg);
int MakeCall( int fd, int _nNid, const char* _pDestUri, const stream_s * _pStream );
int AnswerCall( int fd, int callId );
int Reject( int fd, int callIndex);
int HangupCall( int fd, int _nCallId );
int Report( int fd, const char* message, size_t length );
int SendPacket( int fd , int callIndex, int streamIndex, const char* buffer, size_t size);
int PollEvents( int* eventID, void* event, int nTimeOut );

#endif  /*SDK_INTERFACE_H*/
