#include "linking.h"
#include "input.hpp"
#define THIS_FILE "linking.cpp"
#include <QDebug>
#include <QDateTime>
#include <sstream>
#include <iostream>
#include <QTimer>

linking::linking()
{
    sipServerHost_ = "39.107.247.14";
    iceServerHost_ = "39.107.247.14";
    mqttServerHost_ = "39.107.247.14";
    accName_ = "1742";
    accPwd_ = "2RCuSQx8";
    //accName_ = "1744";
    //accPwd_ =  "QWtIppao";

    accountID_ = -1;
    state = CALL_STATUS_IDLE;
    callID_ = -1;
    isRegistered = false;
    quit_ = false;
    registerAccount();
}

linking::~linking() {
    qDebug()<<"~linking";

}

int linking::call()
{
    if (!isRegistered) {
        logerror("not registered");
        return -1;
    }
    loginfo("call:{}", callee_.c_str());
    logger_flush();
    switch(state) {
    case CALL_STATUS_REGISTER_FAIL:
    case CALL_STATUS_IDLE:
        logerror("call in wrong state:{}", state);
        return -2;
    default:
        break;
    }

    int callid = -1;
    ErrorID err;
    loginfo("account:{} call {}", accountID_, callee_.c_str());
    err = MakeCall(accountID_, callee_.c_str(), sipServerHost_.c_str(), &callid);

    if (RET_OK != err) {
        logerror("call error {}", err);
        return err;
    }

    callID_ = callid;
    return 0;
}

void linking::Stop()
{
    quit_ = true;
    if (state == CALL_STATUS_REGISTERED || state == CALL_STATUS_REGISTER_FAIL) {
        UnRegister(accountID_);
    }

    if (eventThread.joinable()) {
        eventThread.join();
    }
}

int linking::hangup()
{
    if (accountID_ > -1 && callID_ > -1) {
        qDebug()<<"HangupCall"<< accountID_ << callID_;
        ErrorID err = HangupCall( accountID_, callID_);
        if (err != RET_OK) {
            return err;
        }
        callID_ = -1;
        state = CALL_STATUS_HANGUP;
    }
    if(h264File.get() != nullptr && h264File->is_open()) {
        h264File->close();
        h264File.reset();
    }

    return 0;
}

int linking::registerAccount()
{
    loginfo("register:accname:{} accpwd:{} {}",
            accName_.c_str(), accPwd_.c_str(), sipServerHost_.c_str());
    int accid = Register(accName_.c_str(), accPwd_.c_str(), sipServerHost_.c_str(),
                         iceServerHost_.c_str(), mqttServerHost_.c_str());
    if (accid < 0) {
        logerror("Register fail:{}", accid);
        state = CALL_STATUS_REGISTERED;
        return CALL_STATUS_REGISTERED;
    }
    accountID_ = accid;
    eventThread = std::thread(linking::eventHandler, this);
    return 0;
}

std::shared_ptr<std::vector<uint8_t>> linking::PopAudioData()
{
    if (quit_)
        return nullptr;

    std::lock_guard<std::mutex> lock(aqMutex_);
    if (audioQ_.size() > 0){
        auto d = audioQ_.front();
        audioQ_.pop();
        logdebug("ice pop audio data:{}", d->size());
        return d;
    }
    return nullptr;
}

std::shared_ptr<std::vector<uint8_t>> linking::PopVideoData()
{
    if (quit_)
        return nullptr;

    std::lock_guard<std::mutex> lock(vqMutex_);
    if (videoQ_.size() > 0){
        auto d = videoQ_.front();
        videoQ_.pop();
        logdebug("ice pop video data:{}", d->size());
        return d;
    }
    return nullptr;
}

void linking::PushAudioData(uint8_t * ptr, int size)
{
    if (quit_)
        return;
    logdebug("ice push audio data:{}", size);
    std::lock_guard<std::mutex> lock(aqMutex_);
    if (receiveFirstAudio == false) {
        emit onFirstAudio(QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz"));
        receiveFirstAudio = true;
    }
    audioQ_.push(std::make_shared<std::vector<uint8_t>>(ptr, ptr+size));
}

void linking::PushVideoData(uint8_t * ptr, int size)
{
    if (quit_)
        return;
    logdebug("ice push video data:{}", size);
    std::lock_guard<std::mutex> lock(vqMutex_);

    if(h264File.get() == nullptr) {
        std::stringstream streamname;
        streamname << "call"<<(void *)this<<".h264";
        std::string name;
        streamname >> name;
        h264File = std::make_shared<std::fstream>();
        h264File->open(name, std::ios::out | std::ios::binary);
    }
    if(h264File.get() != nullptr && h264File->is_open()) {
        h264File->write((char *)ptr, size);
    }

    if (receiveFirstVideo == false) {
        emit onFirstVideo(QDateTime::currentDateTime().toString("yyyy-MM-dd hh:mm:ss.zzz"));
        receiveFirstVideo = true;
    }
    videoQ_.push(std::make_shared<std::vector<uint8_t>>(ptr, ptr+size));
}

void linking::eventHandler(void *opaque){
    linking * obj = (linking *)opaque;

    EventType type;
    Event* event = (Event*) malloc(sizeof(Event));
    ErrorID id;
    while(!obj->quit_){
        id = PollEvent(obj->accountID_, &type, &event, 0);

        if (id != RET_OK) {
            continue;
        }
        switch (type) {
        case EVENT_CALL:
        {
            CallEvent *pCallEvent = &(event->body.callEvent);
            loginfo("Call status {} call id {} call account id {}",
                    pCallEvent->status, pCallEvent->callID, obj->accountID_);
            if (pCallEvent->status == CALL_STATUS_REGISTERED) {
                    loginfo("CALL_STATUS_REGISTERED");
                    obj->isRegistered = true;
                    obj->state = CALL_STATUS_REGISTERED;
                    if (obj->registerOkEmited == false) {
                        qDebug()<<"sip state ok";
                        obj->registerOkEmited = true;
                        emit obj->registerSuccess();
                    }
            }

            break;
        }
        case EVENT_DATA:
        {
            DataEvent *pDataEvent = &(event->body.dataEvent);
            loginfo("Data size {} call id {} call account id {}",
                    pDataEvent->size, pDataEvent->callID, obj->accountID_);
            qDebug()<<"--------->"<<pDataEvent->size;
            switch(pDataEvent->codec){
            case CODEC_H264:
                obj->PushVideoData((uint8_t*)pDataEvent->data, pDataEvent->size);
                break;
            case CODEC_G711A:
            case CODEC_G711U:
                obj->PushAudioData((uint8_t*)pDataEvent->data, pDataEvent->size);
                break;
            }
            break;
        }
        case EVENT_MESSAGE:
        {
            MessageEvent *pMessage = &(event->body.messageEvent);
            loginfo("Message {} status id {} account id {}",
                    pMessage->message, pMessage->status, obj->accountID_);
            break;
        }
        case EVENT_MEDIA:
        {
            MediaEvent *pMedia = &(event->body.mediaEvent);
            loginfo("Callid {} ncount {} type 1 {} type 2 {} account id {}",
                   pMedia->callID, pMedia->nCount, pMedia->media[0].codecType,
                    pMedia->media[1].codecType, obj->accountID_);
            obj->callID_ = pMedia->callID;
            obj->sendFlag_ = true;
            break;
        }
        }
    }
}

void linking::Reset()
{
    sendFlag_ = false;
    std::lock_guard<std::mutex> lock1(aqMutex_);
    std::lock_guard<std::mutex> lock2(vqMutex_);
    receiveFirstAudio = false;
    receiveFirstVideo = false;

    while(videoQ_.size() > 0) {
        videoQ_.pop();
    }
    while(audioQ_.size() > 0) {
        audioQ_.pop();
    }
}



