#include "tsmuxuploader.h"
#include "base.h"
#include <libavformat/avformat.h>
#include <unistd.h>

#define FF_OUT_LEN 4096
#define QUEUE_INIT_LEN 150

#define TK_STREAM_TYPE_AUDIO 1
#define TK_STREAM_TYPE_VIDEO 2

typedef struct _FFTsMuxContext{
        AsyncInterface asyncWait;
        TsUploader *pTsUploader_;
        AVFormatContext *pFmtCtx_;
        int nOutVideoindex_;
        int nOutAudioindex_;
        int64_t nPrevAudioTimestamp;
        int64_t nPrevVideoTimestamp;
}FFTsMuxContext;

typedef struct _FFTsMuxUploader{
        TsMuxUploader tsMuxUploader_;
        pthread_mutex_t muxUploaderMutex_;
        char token_[256];
        char ak_[64];
        char sk_[64];
        char bucketName_[256];
        int deleteAfterDays_;
        char callback_[512];
        FFTsMuxContext *pTsMuxCtx;
        
        int64_t nLastVideoTimestamp;
        int64_t nLastUploadVideoTimestamp; //initial to -1
        int nKeyFrameCount;
        int nFrameCount;
        int nSegmentId;
        AvArg avArg;
        UploadState ffMuxSatte;
}FFTsMuxUploader;

static void pushRecycle(FFTsMuxUploader *_pFFTsMuxUploader)
{
        if (_pFFTsMuxUploader) {
                
                pthread_mutex_lock(&_pFFTsMuxUploader->muxUploaderMutex_);
                if (_pFFTsMuxUploader->pTsMuxCtx) {
                        av_write_trailer(_pFFTsMuxUploader->pTsMuxCtx->pFmtCtx_);
                        logerror("push to mgr:%p", _pFFTsMuxUploader->pTsMuxCtx);
                        PushFunction(_pFFTsMuxUploader->pTsMuxCtx);
                        _pFFTsMuxUploader->pTsMuxCtx = NULL;
                }
                
                pthread_mutex_unlock(&_pFFTsMuxUploader->muxUploaderMutex_);
        }
        return;
}

static int ffWriteTsPacketToMem(void *opaque, uint8_t *buf, int buf_size)
{
        FFTsMuxContext *pTsMuxCtx = (FFTsMuxContext *)opaque;
        
        int ret = pTsMuxCtx->pTsUploader_->Push(pTsMuxCtx->pTsUploader_, (char *)buf, buf_size);
        if (ret < 0){
                logdebug("write ts to queue fail:%d", ret);
        } else {
                logtrace("write_packet: should write:len:%d  actual:%d\n", buf_size, ret);
        }
        return ret;
}

static int push(FFTsMuxUploader *pFFTsMuxUploader, char * _pData, int _nDataLen, int64_t _nTimestamp, int _nFlag){
        AVPacket pkt;
        av_init_packet(&pkt);
        pkt.data = (uint8_t *)_pData;
        pkt.size = _nDataLen;
        
        //logtrace("push thread id:%d\n", (int)pthread_self());
        pthread_mutex_lock(&pFFTsMuxUploader->muxUploaderMutex_);
        
        FFTsMuxContext *pTsMuxCtx = NULL;
        int count = 0;

        count = 1;
        pTsMuxCtx = pFFTsMuxUploader->pTsMuxCtx;
        while(pTsMuxCtx == NULL && count) {
                pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
                usleep(3*1000);
                pthread_mutex_lock(&pFFTsMuxUploader->muxUploaderMutex_);
                pTsMuxCtx = pFFTsMuxUploader->pTsMuxCtx;
                count--;
        }
        if (pTsMuxCtx == NULL) {
                pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
                logwarn("upload context is NULL");
                return 0;
        }
        if (pTsMuxCtx->pTsUploader_->GetUploaderState(pTsMuxCtx->pTsUploader_) == TK_UPLOAD_FAIL) {
                pFFTsMuxUploader->ffMuxSatte = TK_UPLOAD_FAIL;
                pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
                return 0;
        }

        int ret = 0;
        if (pTsMuxCtx != NULL) {
                if (_nFlag == TK_STREAM_TYPE_AUDIO){
                        //fprintf(stderr, "audio frame: len:%d pts:%lld\n", _nDataLen, _nTimestamp);
                        if (pTsMuxCtx->nPrevAudioTimestamp != 0 && _nTimestamp - pTsMuxCtx->nPrevAudioTimestamp <= 0) {
                                pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
                                logwarn("audio pts not monotonically: prev:%lld now:%lld", pTsMuxCtx->nPrevAudioTimestamp, _nTimestamp);
                                return 0;
                        }
                        pkt.pts = _nTimestamp * 90;
                        pkt.stream_index = pTsMuxCtx->nOutAudioindex_;
                        pkt.dts = pkt.pts;
                        pTsMuxCtx->nPrevAudioTimestamp = _nTimestamp;
                }else{
                        //fprintf(stderr, "video frame: len:%d pts:%lld\n", _nDataLen, _nTimestamp);
                        if (pTsMuxCtx->nPrevVideoTimestamp != 0 && _nTimestamp - pTsMuxCtx->nPrevVideoTimestamp <= 0) {
                                pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
                                logwarn("video pts not monotonically: prev:%lld now:%lld", pTsMuxCtx->nPrevVideoTimestamp, _nTimestamp);
                                return 0;
                        }
                        pkt.pts = _nTimestamp * 90;
                        pkt.stream_index = pTsMuxCtx->nOutVideoindex_;
                        pkt.dts = pkt.pts;
                        pTsMuxCtx->nPrevVideoTimestamp = _nTimestamp;
                }
                
                if ((ret = av_interleaved_write_frame(pTsMuxCtx->pFmtCtx_, &pkt)) < 0) {
                        logerror("Error muxing packet");
                } else {
                        pTsMuxCtx->pTsUploader_->RecordTimestamp(pTsMuxCtx->pTsUploader_, _nTimestamp);
                }
        } else {
                logwarn("upload context is NULL");
        }
        pthread_mutex_unlock(&pFFTsMuxUploader->muxUploaderMutex_);
        return ret;
}

static int PushVideo(TsMuxUploader *_pTsMuxUploader, char * _pData, int _nDataLen, int64_t _nTimestamp, int nIsKeyFrame, int _nIsSegStart)
{
        FFTsMuxUploader *pFFTsMuxUploader = (FFTsMuxUploader *)_pTsMuxUploader;

        int ret = 0;
        if (pFFTsMuxUploader->nKeyFrameCount == 0 && !nIsKeyFrame) {
                logwarn("first video frame not IDR. drop this frame\n");
                return 0;
        }
        if (pFFTsMuxUploader->nLastUploadVideoTimestamp == -1) {
                pFFTsMuxUploader->nLastUploadVideoTimestamp = _nTimestamp;
        }
        // if start new uploader, start from keyframe
        if (nIsKeyFrame) {
                if( (pFFTsMuxUploader->nKeyFrameCount >= 2 && (_nTimestamp - pFFTsMuxUploader->nLastUploadVideoTimestamp) > 4980)
                   //at least 2 keyframe and aoubt last 5 second
                   || (_nIsSegStart && pFFTsMuxUploader->nFrameCount != 0)// new segment is specified
                   ||  pFFTsMuxUploader->ffMuxSatte == TK_UPLOAD_FAIL){   // upload fail
                        pFFTsMuxUploader->nKeyFrameCount = 0;
                        pFFTsMuxUploader->nFrameCount = 0;
                        pFFTsMuxUploader->nLastUploadVideoTimestamp = _nTimestamp;
                        pFFTsMuxUploader->ffMuxSatte = TK_UPLOAD_INIT;
                        pushRecycle(pFFTsMuxUploader);
                        if (_nIsSegStart) {
                                pFFTsMuxUploader->nSegmentId = (int64_t)time(NULL);
                        }
                        ret = TsMuxUploaderStart(_pTsMuxUploader);
                        if (ret != 0) {
                                return ret;
                        }
                }
                pFFTsMuxUploader->nKeyFrameCount++;
        }

        pFFTsMuxUploader->nLastVideoTimestamp = _nTimestamp;
        
        ret = push(pFFTsMuxUploader, _pData, _nDataLen, _nTimestamp, TK_STREAM_TYPE_VIDEO);
        if (ret == 0){
                pFFTsMuxUploader->nFrameCount++;
        }
        return ret;
}

static int PushAudio(TsMuxUploader *_pTsMuxUploader, char * _pData, int _nDataLen, int64_t _nTimestamp)
{
        FFTsMuxUploader *pFFTsMuxUploader = (FFTsMuxUploader *)_pTsMuxUploader;
        int ret = push(pFFTsMuxUploader, _pData, _nDataLen, _nTimestamp, TK_STREAM_TYPE_AUDIO);
        if (ret == 0){
                pFFTsMuxUploader->nFrameCount++;
        }
        return ret;
}

static int waitToCompleUploadAndDestroyTsMuxContext(void *_pOpaque)
{
        FFTsMuxContext *pTsMuxCtx = (FFTsMuxContext*)_pOpaque;
        
        if (pTsMuxCtx) {
                if (pTsMuxCtx->pFmtCtx_) {
                        if (pTsMuxCtx->pFmtCtx_ && !(pTsMuxCtx->pFmtCtx_->oformat->flags & AVFMT_NOFILE))
                                avio_close(pTsMuxCtx->pFmtCtx_->pb);
                        avformat_free_context(pTsMuxCtx->pFmtCtx_);
                }
                pTsMuxCtx->pTsUploader_->UploadStop(pTsMuxCtx->pTsUploader_);

                UploaderStatInfo statInfo = {0};
                pTsMuxCtx->pTsUploader_->GetStatInfo(pTsMuxCtx->pTsUploader_, &statInfo);
                logdebug("uploader push:%d pop:%d remainItemCount:%d", statInfo.nPushDataBytes_,
                         statInfo.nPopDataBytes_, statInfo.nLen_);
                DestroyUploader(&pTsMuxCtx->pTsUploader_);
                free(pTsMuxCtx);
        }
        
        return 0;
}

static int newFFTsMuxContext(FFTsMuxContext ** _pTsMuxCtx, AvArg *_pAvArg)
{
        FFTsMuxContext * pTsMuxCtx = (FFTsMuxContext *)malloc(sizeof(FFTsMuxContext));
        if (pTsMuxCtx == NULL) {
                return TK_NO_MEMORY;
        }
        memset(pTsMuxCtx, 0, sizeof(FFTsMuxContext));
        
        int ret = NewUploader(&pTsMuxCtx->pTsUploader_, TSQ_FIX_LENGTH, FF_OUT_LEN, QUEUE_INIT_LEN);
        if (ret != 0) {
                free(pTsMuxCtx);
                return ret;
        }
        
        //Output
        ret = avformat_alloc_output_context2(&pTsMuxCtx->pFmtCtx_, NULL, "mpegts", NULL);
        if (ret < 0) {
                logerror("Could not create output context\n");
                goto end;
        }
        AVOutputFormat *pOutFmt = pTsMuxCtx->pFmtCtx_->oformat;
        uint8_t *pOutBuffer = (unsigned char*)av_malloc(4096);
        AVIOContext *avio_out = avio_alloc_context(pOutBuffer, 4096, 1, pTsMuxCtx, NULL, ffWriteTsPacketToMem, NULL);
        pTsMuxCtx->pFmtCtx_->pb = avio_out;
        pTsMuxCtx->pFmtCtx_->flags = AVFMT_FLAG_CUSTOM_IO;
        pOutFmt->flags |= AVFMT_NOFILE;
        pOutFmt->flags |= AVFMT_NODIMENSIONS;
        //ofmt->video_codec //是否指定为ifmt_ctx_v的视频的codec_type.同理音频也一样
        //测试下来即使video_codec和ifmt_ctx_v的视频的codec_type不一样也是没有问题的
        
        //add video
        AVStream *pOutStream = avformat_new_stream(pTsMuxCtx->pFmtCtx_, NULL);
        if (!pOutStream) {
                logerror("Failed allocating output stream\n");
                ret = AVERROR_UNKNOWN;
                goto end;
        }
        pOutStream->time_base.num = 1;
        pOutStream->time_base.den = 90000;
        pTsMuxCtx->nOutVideoindex_ = pOutStream->index;
        pOutStream->codecpar->codec_tag = 0;
        pOutStream->codecpar->codec_type = AVMEDIA_TYPE_VIDEO;
        if (_pAvArg->nVideoFormat == TK_VIDEO_H264)
                pOutStream->codecpar->codec_id = AV_CODEC_ID_H264;
        else
                pOutStream->codecpar->codec_id = AV_CODEC_ID_H265;
        //end add video
        
        //add audio
        pOutStream = avformat_new_stream(pTsMuxCtx->pFmtCtx_, NULL);
        if (!pOutStream) {
                logerror("Failed allocating output stream\n");
                ret = AVERROR_UNKNOWN;
                goto end;
        }
        pOutStream->time_base.num = 1;
        pOutStream->time_base.den = 90000;
        pTsMuxCtx->nOutAudioindex_ = pOutStream->index;
        pOutStream->codecpar->codec_tag = 0;
        pOutStream->codecpar->codec_type = AVMEDIA_TYPE_AUDIO;
        switch(_pAvArg->nAudioFormat){
                case TK_AUDIO_PCMU:
                        pOutStream->codecpar->codec_id = AV_CODEC_ID_PCM_MULAW;
                        break;
                case TK_AUDIO_PCMA:
                        pOutStream->codecpar->codec_id = AV_CODEC_ID_PCM_ALAW;
                        break;
                case TK_AUDIO_AAC:
                        pOutStream->codecpar->codec_id = AV_CODEC_ID_AAC;
                        break;
        }
        pOutStream->codecpar->sample_rate = _pAvArg->nSamplerate;
        pOutStream->codecpar->channels = _pAvArg->nChannels;
        pOutStream->codecpar->channel_layout = av_get_default_channel_layout(pOutStream->codecpar->channels);
        //end add audio
        
        //printf("==========Output Information==========\n");
        //av_dump_format(pTsMuxCtx->pFmtCtx_, 0, "xx.ts", 1);
        //printf("======================================\n");

        //Open output file
        if (!(pOutFmt->flags & AVFMT_NOFILE)) {
                if (avio_open(&pTsMuxCtx->pFmtCtx_->pb, "xx.ts", AVIO_FLAG_WRITE) < 0) {
                        logerror("Could not open output file '%s'", "xx.ts");
                        goto end;
                }
        }
        //Write file header
        int erno = 0;
        if ((erno = avformat_write_header(pTsMuxCtx->pFmtCtx_, NULL)) < 0) {
                char errstr[512] = { 0 };
                av_strerror(erno, errstr, sizeof(errstr));
                logerror("Error occurred when opening output file:%s\n", errstr);
                goto end;
        }
        
        pTsMuxCtx->asyncWait.function = waitToCompleUploadAndDestroyTsMuxContext;
        *_pTsMuxCtx = pTsMuxCtx;
        return 0;
end:
        if (pTsMuxCtx->pFmtCtx_) {
                if (pTsMuxCtx->pFmtCtx_ && !(pOutFmt->flags & AVFMT_NOFILE))
                        avio_close(pTsMuxCtx->pFmtCtx_->pb);
                avformat_free_context(pTsMuxCtx->pFmtCtx_);
                if (ret < 0 && ret != AVERROR_EOF) {
                        logerror("Error occurred.\n");
                        return -1;
                }
        }
        
        return ret;
}

static void setToken(TsMuxUploader* _PTsMuxUploader, char *_pToken)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        assert(strlen(_pToken) < sizeof(pFFTsMuxUploader->token_));
        strcpy(pFFTsMuxUploader->token_, _pToken);
}

static void setAccessKey(TsMuxUploader* _PTsMuxUploader, char *_pAk, int _nAkLen)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        assert(sizeof(pFFTsMuxUploader->ak_) - 1 > _nAkLen);
        memcpy(pFFTsMuxUploader->ak_, _pAk, _nAkLen);
}

static void setSecretKey(TsMuxUploader* _PTsMuxUploader, char *_pSk, int _nSkLen)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        assert(sizeof(pFFTsMuxUploader->sk_) - 1 > _nSkLen);
        memcpy(pFFTsMuxUploader->sk_, _pSk, _nSkLen);
}

static void setBucket(TsMuxUploader* _PTsMuxUploader, char *_pBucketName, int _nBucketNameLen)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        assert(sizeof(pFFTsMuxUploader->bucketName_) - 1 > _nBucketNameLen);
        memcpy(pFFTsMuxUploader->bucketName_, _pBucketName, _nBucketNameLen);
}

static void setCallbackUrl(TsMuxUploader* _PTsMuxUploader, char *_pCallbackUrl, int _nCallbackUrlLen)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        assert(sizeof(pFFTsMuxUploader->callback_) - 1 > _nCallbackUrlLen);
        memcpy(pFFTsMuxUploader->callback_, _pCallbackUrl, _nCallbackUrlLen);
}

static void setDeleteAfterDays(TsMuxUploader* _PTsMuxUploader, int nDays)
{
        FFTsMuxUploader * pFFTsMuxUploader = (FFTsMuxUploader *)_PTsMuxUploader;
        pFFTsMuxUploader->deleteAfterDays_ = nDays;
}

int NewTsMuxUploader(TsMuxUploader **_pTsMuxUploader, AvArg *_pAvArg)
{
        FFTsMuxUploader *pFFTsMuxUploader = (FFTsMuxUploader*)malloc(sizeof(FFTsMuxUploader));
        if (pFFTsMuxUploader == NULL) {
                return TK_NO_MEMORY;
        }
        memset(pFFTsMuxUploader, 0, sizeof(FFTsMuxUploader));
        pFFTsMuxUploader->nLastUploadVideoTimestamp = -1;
        
        int ret = 0;
        ret = pthread_mutex_init(&pFFTsMuxUploader->muxUploaderMutex_, NULL);
        if (ret != 0){
                free(pFFTsMuxUploader);
                return TK_MUTEX_ERROR;
        }
        
        pFFTsMuxUploader->tsMuxUploader_.SetToken = setToken;
        pFFTsMuxUploader->tsMuxUploader_.SetSecretKey = setSecretKey;
        pFFTsMuxUploader->tsMuxUploader_.SetAccessKey = setAccessKey;
        pFFTsMuxUploader->tsMuxUploader_.SetBucket = setBucket;
        pFFTsMuxUploader->tsMuxUploader_.SetCallbackUrl = setCallbackUrl;
        pFFTsMuxUploader->tsMuxUploader_.SetDeleteAfterDays = setDeleteAfterDays;
        pFFTsMuxUploader->tsMuxUploader_.PushAudio = PushAudio;
        pFFTsMuxUploader->tsMuxUploader_.PushVideo = PushVideo;
        
        pFFTsMuxUploader->avArg.nAudioFormat = _pAvArg->nAudioFormat;
        pFFTsMuxUploader->avArg.nChannels = _pAvArg->nChannels;
        pFFTsMuxUploader->avArg.nSamplerate = _pAvArg->nSamplerate;
        pFFTsMuxUploader->avArg.nVideoFormat = _pAvArg->nVideoFormat;
        
        *_pTsMuxUploader = (TsMuxUploader *)pFFTsMuxUploader;
        
        return 0;
}

int TsMuxUploaderStart(TsMuxUploader *_pTsMuxUploader)
{
        FFTsMuxUploader *pFFTsMuxUploader = (FFTsMuxUploader *)_pTsMuxUploader;
        
        assert(pFFTsMuxUploader->pTsMuxCtx == NULL);
        
        int ret = newFFTsMuxContext(&pFFTsMuxUploader->pTsMuxCtx, &pFFTsMuxUploader->avArg);
        if (ret != 0) {
                free(pFFTsMuxUploader);
                return ret;
        }
        if (pFFTsMuxUploader->token_[0] == 0) {
                pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetAccessKey(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_,
                                                                        pFFTsMuxUploader->ak_, strlen(pFFTsMuxUploader->ak_));
                pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetSecretKey(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_,
                                                                        pFFTsMuxUploader->sk_, strlen(pFFTsMuxUploader->sk_));
        } else {
                pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetToken(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_, pFFTsMuxUploader->token_);
        }
        pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetBucket(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_,
                                                             pFFTsMuxUploader->bucketName_, strlen(pFFTsMuxUploader->bucketName_));
        pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetCallbackUrl(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_,
                                                                  pFFTsMuxUploader->callback_, strlen(pFFTsMuxUploader->callback_));
        if (pFFTsMuxUploader->deleteAfterDays_ == 0) {
                pFFTsMuxUploader->deleteAfterDays_  = 7;
                pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetDeleteAfterDays(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_,
                                                                              pFFTsMuxUploader->deleteAfterDays_);
        }

        if (pFFTsMuxUploader->nSegmentId == 0) {
                pFFTsMuxUploader->nSegmentId = (int64_t)time(NULL);
        }
        pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->SetSegmentId(
                                                                pFFTsMuxUploader->pTsMuxCtx->pTsUploader_, pFFTsMuxUploader->nSegmentId );
        pFFTsMuxUploader->pTsMuxCtx->pTsUploader_->UploadStart(pFFTsMuxUploader->pTsMuxCtx->pTsUploader_);
        return 0;
}

void DestroyTsMuxUploader(TsMuxUploader **_pTsMuxUploader)
{
        FFTsMuxUploader *pFFTsMuxUploader = (FFTsMuxUploader *)(*_pTsMuxUploader);
        
        pushRecycle(pFFTsMuxUploader);
        if (pFFTsMuxUploader) {
                free(pFFTsMuxUploader);
        }
        return;
}
