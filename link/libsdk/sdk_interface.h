// Last Update:2018-05-25 21:03:43
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
    STREAM_TYPE_NONE,
    STREAM_TYPE_VIDEO,
    STREAMD_TYPE_AUDIO,
} stream_type_e;

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
    RET_SUCESS,
    RET_FAIL,
} status_e;

typedef struct {
    stream_type_e type;
    int samplerate;
    int channels;
    int width;
    int height;
} stream_s;



int Register(const struct UA* ua, const char* id, const char* host, const char* password);
int MakeCall(const struct UA* ua, const char* id, const char* host);
int AnswerCall(const struct UA* ua, int callIndex);
int Reject(const struct UA* ua, int callIndex);
int HangupCall(const struct UA* ua, int callIndex);
int Report(struct UA* ua, const char* message, size_t length);
int SendPacket(const struct UA* ua, int callIndex, int streamIndex, const char* buffer, size_t size);
int PollEvents(const struct UA* ua, int* eventID, void* event);

#endif  /*SDK_INTERFACE_H*/
