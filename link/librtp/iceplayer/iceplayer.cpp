#include "iceplayer.h"
#include <QtQuick/qquickwindow.h>
#include <QtGui/QOpenGLShaderProgram>
#include <QtGui/QOpenGLContext>
#include <QtWidgets/QGraphicsView>
#include <QtWidgets/QGraphicsScene>
#include <QFile>
#include "g711.h"
#define THIS_FILE "iceplayer.cpp"

#include <QFile>

QFile audioFile("/Users/liuye/Documents/qml/a.mulaw");
static int audioFileReadSize = 0;
int audioFeed(void *opaque, uint8_t *buf, int buf_size)
{
    if(!audioFile.isOpen()) {
        audioFile.open(QIODevice::ReadOnly);
    }
    if(audioFile.isOpen()) {
        int n = audioFile.read((char *)buf, buf_size);
        if (n == 0)
            return 0;
        audioFileReadSize+=n;
        return n;
    }
    return -1;
}

QFile videoFile("/Users/liuye/Documents/qml/v.h264");
static int videoFileReadSize = 0;
int videoFeed(void *opaque, uint8_t *buf, int buf_size)
{
    if(!videoFile.isOpen()) {
        videoFile.open(QIODevice::ReadOnly);
    }
    if(videoFile.isOpen()) {
        int n = videoFile.read((char *)buf, buf_size);
        if (n == 0)
            return 0;
        videoFileReadSize+=n;
        return n;
    }
    return -1;
}


QGLRenderer::~QGLRenderer()
{
    delete m_program;
}

QGLRenderer::QGLRenderer() : m_program(0) {
    m_textures[0] = 0;
    m_textures[1] = 0;
    m_textures[2] = 0;
    xleft = -1.0f;
    xright = 1.0f;
    ytop = 1.0f;
    ybottom = -1.0f;
    if (!m_program) {
        initializeOpenGLFunctions();

        m_program = new QOpenGLShaderProgram();
        m_program->addCacheableShaderFromSourceCode(QOpenGLShader::Vertex,
                                                    "attribute highp vec3 vertexIn;\n"
                                                    "attribute highp vec2 textureIn;\n"
                                                    "varying vec2 textureOut;\n"
                                                    "void main(void)\n"
                                                    "{\n"
                                                    "	gl_Position = vec4(vertexIn, 1.0);\n"
                                                    "	textureOut = textureIn;\n"
                                                    "}\n");
        m_program->addCacheableShaderFromSourceCode(QOpenGLShader::Fragment,
                                                    "varying vec2 textureOut;\n"
                                                    "uniform sampler2D tex0;\n"
                                                    "uniform sampler2D tex1;\n"
                                                    "uniform sampler2D tex2;\n"
                                                    "void main(void)\n"
                                                    "{\n"
                                                    "	vec3 yuv;\n"
                                                    "	vec3 rgb;\n"
                                                    "	yuv.x = texture2D(tex0, textureOut).r;\n"
                                                    "	yuv.y = texture2D(tex1, textureOut).r - 0.5;\n"
                                                    "	yuv.z = texture2D(tex2, textureOut).r - 0.5;\n"
                                                    "	rgb = mat3(1, 1, 1,\n"
                                                    "		0, -0.39465, 2.03211,\n"
                                                    "		1.13983, -0.58060, 0) * yuv;\n"
                                                    "	//gl_FragColor = vec4(0.0,0.0,1.0, 1);\n"
                                                    "	gl_FragColor = vec4(rgb, 1);\n"
                                                    "}\n");

        m_program->bindAttributeLocation("vertices", 0);
        m_program->link();

        glGenTextures(3, m_textures);
        for (int i = 0; i < 3; i++) {
            glBindTexture(GL_TEXTURE_2D, m_textures[i]); // All upcoming GL_TEXTURE_2D operations now have effect on our texture object
            // Set our texture parameters
            glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_S, GL_CLAMP_TO_EDGE);    // Note that we set our container wrapping method to GL_CLAMP_TO_EDGE
            glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_WRAP_T, GL_CLAMP_TO_EDGE);    // Note that we set our container wrapping method to GL_CLAMP_TO_EDGE
            // Set texture filtering
            glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST);
            glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST);
            glBindTexture(GL_TEXTURE_2D, m_textures[i]); // Unbind texture when done, so we won't accidentily mess up our texture.
        }
    }
}

void QGLRenderer::SetFrame(std::shared_ptr<MediaFrame> &frame)
{
    std::lock_guard<std::mutex> lock(m_frameMutex);
    m_frame = frame;
}

void QGLRenderer::ClearFrame()
{
    std::lock_guard<std::mutex> lock(m_frameMutex);
    m_frame.reset();
}

void QGLRenderer::setDrawRect(QRectF & s, QRectF &it){
    qreal sw = (s.right() - s.left())/2;
    qreal sh = (s.bottom() - s.top())/2;

    xleft = (it.left() - s.left() - sw) / sw;
    xright = (it.right() - s.left() - sw) / sw;
    ytop = (s.bottom() - it.top() - sh) / sh;
    ybottom = (s.bottom() - it.bottom() - sh) / sh;

    logtrace("\n{} {}\n{} {}\n{} {}\n {} {}\{} {}\n{} {}",
            xleft, ybottom ,
            xright ,ytop,
            xleft , ytop,
            xleft , ybottom,
            xright ,ybottom,
            xright ,ytop);
}


void QGLRenderer::paint()
{
    float values[] = {
        //left top triangle
        xleft, ybottom, 0.0f,     0.0f, 1.0f,
        xright, ytop, 0.0f,       1.0f, 0.0f,
        xleft, ytop, 0.0f,      0.0f, 0.0f,
        //right down triangle
        xleft, ybottom, 0.0f,     0.0f, 1.0f,
        xright, ybottom, 0.0f,      1.0f, 1.0f,
        xright, ytop, 0.0f,       1.0f, 0.0f,
    };

    std::lock_guard<std::mutex> lock(m_frameMutex);
    AVFrame * f = nullptr;
    if (m_frame.get() != nullptr)
        f = m_frame->AvFrame();

    if (f != nullptr) {
        m_program->bind();
        m_program->enableAttributeArray(0);
        m_program->enableAttributeArray(1);

        m_program->setAttributeArray(0, GL_FLOAT, values, 3, 5 * sizeof(GLfloat));
        m_program->setAttributeArray(1, GL_FLOAT, &values[3], 2, 5 * sizeof(GLfloat));

        for (int i = 0, j = 1; i < 3; i++, j = 2) {
            char name[5] = {0};
            sprintf(name, "tex%d", i);
            int location = m_program->uniformLocation(name);;
            glActiveTexture(GL_TEXTURE0 + i);
            glBindTexture(GL_TEXTURE_2D, m_textures[i]);

            glPixelStorei(GL_UNPACK_ROW_LENGTH,
                          f->linesize[i]);
            glTexImage2D(GL_TEXTURE_2D, 0, GL_LUMINANCE,
                         f->width/j,
                         f->height/j, 0, GL_LUMINANCE, GL_UNSIGNED_BYTE, f->data[i]);
            glUniform1i(location, i);
        }

        glViewport(0, 0, m_viewportSize.width(), m_viewportSize.height());

        //glDisable(GL_DEPTH_TEST);
        //glClearColor(0, 0, 0, 1);
        //glClear(GL_COLOR_BUFFER_BIT);

        //glEnable(GL_BLEND);
        //glBlendFunc(GL_SRC_ALPHA, GL_ONE);


        glDrawArrays(GL_TRIANGLES, 0, 6);
        m_program->disableAttributeArray(0);
        m_program->disableAttributeArray(1);
        m_program->release();

        // Not strictly needed for this example, but generally useful for when
        // mixing with raw OpenGL.
        m_window->resetOpenGLState();
        //glSwapAPPLE();
    }
}

AudioRender::AudioRender(){

}

void AudioRender::Init(QAudioFormat config){
    m_audioConfig = config;
    m_inited = false;
    QAudioDeviceInfo info = QAudioDeviceInfo::defaultOutputDevice();
    if (!info.isFormatSupported(m_audioConfig)) {
        logerror("default format not supported try to use nearest");
        m_audioConfig = info.nearestFormat(m_audioConfig);
        return;
    }
    m_inited = true;
    m_audioOutput = std::make_shared<QAudioOutput>(m_audioConfig);
    m_device = m_audioOutput->start();
    m_audioOutput->setVolume(0.8);
    loginfo("init volume:{}", m_audioOutput->volume());
}

void AudioRender::Init(std::shared_ptr<MediaFrame> & frame)
{
    AVFrame * f = frame->AvFrame();
    if (!m_inited) {
        bool configOk = true;
        QAudioFormat config;
        config.setSampleRate(f->sample_rate);
        config.setChannelCount(f->channels);
        config.setCodec("audio/pcm");
        config.setByteOrder(QAudioFormat::LittleEndian);

        switch(f->format) {
        case AV_SAMPLE_FMT_U8:
            qDebug()<<"AV_SAMPLE_FMT_U8";
            config.setSampleSize(8);
            config.setSampleType(QAudioFormat::UnSignedInt);
            break;
        case AV_SAMPLE_FMT_S16:
            qDebug()<<"AV_SAMPLE_FMT_S16";
            config.setSampleSize(16);
            config.setSampleType(QAudioFormat::SignedInt);
            break;
        case AV_SAMPLE_FMT_S32:
            qDebug()<<"AV_SAMPLE_FMT_S32";
            config.setSampleSize(32);
            config.setSampleType(QAudioFormat::SignedInt);
            break;
        case AV_SAMPLE_FMT_FLT:
            qDebug()<<"AV_SAMPLE_FMT_FLT";
            config.setSampleSize(32);
            config.setSampleType(QAudioFormat::Float);
            break;

        case AV_SAMPLE_FMT_U8P:
        case AV_SAMPLE_FMT_S16P:
        case AV_SAMPLE_FMT_S32P:
        case AV_SAMPLE_FMT_FLTP:
        case AV_SAMPLE_FMT_DBL:
        case AV_SAMPLE_FMT_DBLP:
        case AV_SAMPLE_FMT_S64:
        case AV_SAMPLE_FMT_S64P:
            configOk = false;
            logerror("not soupport:{}", f->format);
            break;
        }
        if(configOk)
            Init(config);
        qDebug()<<"audio renderer configOk:"<<configOk << " rate:"<<f->sample_rate
               <<" channel:"<< f->channels << " initok:"<< IsInited();
    }
}

void AudioRender::Uninit(){
    if (m_inited) {
        if (m_device) {
            m_device->destroyed();
            m_device = nullptr;
        }

        if (m_audioOutput) {
            m_audioOutput->stop();
            m_audioOutput = nullptr;
        }
        m_inited = false;
    }
}

void AudioRender::PushData(void *pcmData,int size){
    if(m_inited == false)
        return;
    m_device->write((const char *)pcmData, size);
    //qint64 ret = m_device->write((const char *)pcmData, size);
    //qDebug()<<"wiret audio return:"<<ret;
}

void AudioRender::PushG711Data(void *g711Data, int size, int lawType){
    if(m_inited == false)
        return;
    short pcm[1280];
    unsigned char *src = (unsigned char*)g711Data;


    if (lawType == alawType) {
        for (int i = 0; i < size; i++) {
            pcm[i] = alaw2linear(src[i]);
        }
        PushData(pcm, size * 2);
    } else if (lawType == ulawType) {
        for (int i = 0; i < size; i++) {
            pcm[i] = ulaw2linear(src[i]);
        }
        PushData(pcm, size * 2);
    }
}

#include <QDir>
IcePlayer::IcePlayer()
  : m_t(0)
  , m_vRenderer(0)
  , registerOk(false)
{
    connect(this, &QQuickItem::windowChanged, this, &IcePlayer::handleWindowChanged);
    m_vRenderer = nullptr;
    loginfo("pwd:{}", QDir::currentPath().toStdString());


    //emit player->pictureReady();

    auto timerHandle = [this]() {
        while(!quit_) {
            os_sleep_ms(1000);
            updateStreamInfo();
        }
    };
    timerThread_ = std::thread(timerHandle);
    auto avsync = [this]() {

        while(!quit_) {
            int asize = 0;
            int vsize = 0;
            std::shared_ptr<MediaFrame> aframe;
            std::shared_ptr<MediaFrame> vframe;
            int64_t aPts = -1;
            int64_t vPts = -1;

            {
                std::unique_lock<std::mutex> lock(mutex_);
                asize = Abuffer_.size();
                vsize = Vbuffer_.size();
                if (canRender_ || (asize == 0 && vsize == 0)) {
                    condition_.wait(lock);
                    asize = Abuffer_.size();
                    vsize = Vbuffer_.size();

                    if (asize == 0 && vsize == 0){
                        continue;
                    }

                    if (asize > 0){
                        aframe = Abuffer_.front();
                        Abuffer_.pop_front();
                        aPts = aframe->pts;
                    }
                    if (vsize > 0){
                        vframe = Vbuffer_.front();
                        Vbuffer_.pop_front();
                        vPts = vframe->pts;
                    }
                }
            }

            bool aRenderFist = false;
            if ((vPts == -1) || (aPts != -1 && aPts < vPts))
                aRenderFist = true;

            if (!m_aRenderer.IsInited() && aframe.get() != nullptr){
                m_aRenderer.Init(aframe);
            }

            int64_t now = os_gettime_ms();
            int diff = now - firstFrameTime_;
            if (diff < 0)
                diff = 0;

            logdebug("ready to render:{} {}", aPts, vPts);
            if (aRenderFist) {
                if (diff < aPts - 1) {
                    qDebug()<<"a first: asleep:"<<aPts - 1 - diff << " apts:" << aPts << " vPts:"<<vPts;
                    os_sleep_ms(aPts - 1 - diff);
                }
                 m_aRenderer.PushData(aframe->AvFrame()->data[0], aframe->AvFrame()->linesize[0]);
                 if (vPts != -1) {
                     if (diff < vPts - 1) {
                         qDebug()<<"a first: vsleep:"<<vPts - 1 - diff << " apts:" << aPts << " vPts:"<<vPts;;
                         os_sleep_ms(vPts - 1 - diff);
                     }
                     m_vRenderer->SetFrame(vframe);
                     emit pictureReady();
                 }
            } else {
                if (vPts != -1 && diff < vPts - 1) {
                    qDebug()<<"v first: vsleep:"<<vPts - 1 - diff<< " apts:" << aPts << " vPts:"<<vPts;;
                    os_sleep_ms(vPts - 1 - diff);
                }
                m_vRenderer->SetFrame(vframe);
                emit pictureReady();
                if (aPts != -1) {
                    if (diff < aPts - 1) {
                        qDebug()<<"v first: asleep:"<<aPts - 1 - diff<< " apts:" << aPts << " vPts:"<<vPts;;
                        os_sleep_ms(aPts - 1 - diff);
                    }
                    m_aRenderer.PushData(aframe->AvFrame()->data[0], aframe->AvFrame()->linesize[0]);
                }
            }

        }

    };
    avsync_ = std::thread(avsync);
}

IcePlayer::~IcePlayer()
{

}
void IcePlayer::updateStreamInfo()
{
    std::lock_guard<std::mutex> lock(streamMutex_);

    if (m_stream1.get() && m_stream2.get()) {
        MediaStatInfo info;
        info.Add(m_stream1->GetMediaStatInfo());
        info.Add(m_stream2->GetMediaStatInfo());

        char infoStr[128] = {0};
        sprintf(infoStr, "vFps:%d vBr:%d kbps vCount:%d | aFps:%d aBr:%d aCount:%d",
                        info.videoFps, info.videoBitrate * 8 / 1000, info.totalVideoFrameCount,
                        info.audioFps, info.audioBitrate * 8 / 1000, info.totalAudioFrameCount);
        qDebug()<<infoStr;
        emit streamInfoUpdate(infoStr);
        return;
    }
    if (m_stream1.get()){
        const char * infoStr = m_stream1->GetMediaStatInfoStr();
        qDebug()<<infoStr;
        emit streamInfoUpdate(infoStr);
    }
}

void IcePlayer::handleWindowChanged(QQuickWindow *win)
{
    if (win) {
        connect(win, &QQuickWindow::beforeSynchronizing, this, &IcePlayer::sync, Qt::DirectConnection);
        connect(win, &QQuickWindow::sceneGraphInvalidated, this, &IcePlayer::cleanup, Qt::DirectConnection);
        // If we allow QML to do the clearing, they would clear what we paint
        // and nothing would show.
        //win->setClearBeforeRendering(false);
    }
}

void IcePlayer::cleanup()
{
    if (m_vRenderer) {
        delete m_vRenderer;
        m_vRenderer = 0;
    }
}

void IcePlayer::repaint()
{
    if (m_vRenderer) {
        if (window()) {
            loginfo("render one frame");
            window()->update();
        }
    }
}


void IcePlayer::sync()
{
    if (!m_vRenderer) {
        m_vRenderer = new QGLRenderer();
        connect(window(), &QQuickWindow::afterRendering, m_vRenderer, &QGLRenderer::paint, Qt::DirectConnection);
        connect(this, &IcePlayer::pictureReady, this, &IcePlayer::repaint, Qt::QueuedConnection);
    }
    QSize wsize = window()->size();
    logdebug("window size: width:{} height:{}", wsize.width(), wsize.height());

    QObject * obj = window()->findChild<QObject *>("player");
    if(obj != nullptr){
#ifdef MORE_DETAILS
        QVariant h = obj->property("height");
        QVariant w = obj->property("width");
        logtrace("name:height type:{} value:{}", h.type().toString().toStdString(), h.toDouble());
        logtrace("name:width: type:{} value:{}", w.type().toString().toStdString(), w.toDouble());
        QVariant n = obj->property("objectName");
        logtrace("name:objectName type:{} value:{}", n.type().toString().toStdString(),
                 n.toString().toStdString());
        obj->dumpObjectInfo();
#endif

        QQuickItem * item = dynamic_cast<QQuickItem *>(obj);
        if (item != nullptr){
            QRectF rect = item->boundingRect();
            //qDebug()<<"      --->:boundingRect:"<<item->rect();
            logtrace("playerRect's boundingRect:({} {} {} {})",
                     rect.left(), rect.top(), rect.right(), rect.bottom());
            QRectF s = item->mapRectToScene(item->boundingRect());
            //qDebug()<<"      --->:map:"<< s;
            logtrace("playerRect's mapRectToScene:({} {} {} {})",
                     s.left(), s.top(), s.right(), s.bottom());

            QRect wrect = window()->geometry();
            //qDebug()<<"      --->:window rect:"<< wrect;
            logtrace("window geometry:({} {} {} {})",
                wrect.left(), wrect.top(), wrect.right(), wrect.bottom());

            QRectF c;
            c.setLeft(0.0);
            c.setBottom(0.0);
            c.setRight(window()->geometry().width());
            c.setBottom(window()->geometry().height());
            //qDebug()<<"      --->:window rect:"<< c;
            logtrace("window geometry:({} {} {} {})",
                c.left(), c.top(), c.right(), c.bottom());
            m_vRenderer->setDrawRect(c, s);
        }
    } else {
        logerror("not found player");
    }

    m_vRenderer->setViewportSize(wsize * window()->devicePixelRatio());
    m_vRenderer->setWindow(window());
}

void IcePlayer::setSourceType(QVariant stype){
    sourceType_ = stype.toInt();
}

void IcePlayer::StopStream() {
    logger_flush();
    qDebug()<<"iceplayer stop";
    std::lock_guard<std::mutex> lock(streamMutex_);
    if (m_stream1.get() != nullptr){
        ThreadCleaner::GetThreadCleaner()->Push(m_stream1);
        m_stream1.reset();
        qDebug()<<"stream1 stop";
    }
    if (m_stream2.get() != nullptr){
        ThreadCleaner::GetThreadCleaner()->Push(m_stream2);
        m_stream2.reset();
        qDebug()<<"stream2 stop";
    }

}

void IcePlayer::Stop() {
    ThreadCleaner::GetThreadCleaner()->Push(iceSource_);
    hangup();
    mutex_.lock();
    quit_ = true;
    condition_.notify_one();
    mutex_.unlock();
    if (timerThread_.joinable()) {
        timerThread_.join();
    }
    if(avsync_.joinable()) {
        avsync_.join();
    }
}

void IcePlayer::getFrameCallback(void * userData, std::shared_ptr<MediaFrame> & frame) {
    IcePlayer * player = (IcePlayer *)(userData);
    std::unique_lock<std::mutex> lock(player->mutex_);
    if (player->canRender_ == false) {
        qDebug()<<"xxxxxxx canRender:"<<player->canRender_;
        return;
    }

    if(frame->GetStreamType() == STREAM_AUDIO) {
        logdebug("audio framepts:{}", frame->pts);
        player->push(std::move(frame));
    } else {
        logdebug("video framepts:{}", frame->pts);
        player->push(std::move(frame));
    }
    player->condition_.notify_one();
}

void IcePlayer::call(QVariant sipAccount){
    QString strSipAcc = sipAccount.toString();
    qDebug()<<strSipAcc;

    if(sourceType_ == 0) {
        if (iceSource_.get() == nullptr) {
            iceSource_ = std::make_shared<linking>();
            iceSource_->SetCallee(strSipAcc.toStdString());
            qDebug()<<"first call"<<strSipAcc;
            connect(iceSource_.get(), SIGNAL(registerSuccess()), this, SLOT(makeCall()));
            connect(iceSource_.get(), SIGNAL(onFirstAudio(QString)), this, SLOT(firstAudioPktTime(QString)));
            connect(iceSource_.get(), SIGNAL(onFirstVideo(QString)), this, SLOT(firstVideoPktTime(QString)));
        }
        iceSource_->SetCallee(strSipAcc.toStdString());

        auto state = iceSource_->GetState();
         loginfo("make call to {} {}", strSipAcc.toStdString(), state);
        if (state == CALL_STATUS_IDLE || state == CALL_STATUS_REGISTER_FAIL) {
            logdebug("sip state wrong");
            return;
        }
    }
    makeCall();
    canRender_ = true;
}

void IcePlayer::firstAudioPktTime(QString timestr) {
    qDebug()<<"afirst:"<<timestr;
    emit getFirstAudioPktTime(timestr);
}
void IcePlayer::firstVideoPktTime(QString timestr) {
    qDebug()<<"vfirst:"<<timestr;
    emit getFirstVideoPktTime(timestr);
}

void IcePlayer::makeCall(){
    qDebug()<<"qmlmakeCall";
    if(sourceType_ == 0) {
        iceSource_->call();
    }

    InputParam param1;
    InputParam param2;
    if(sourceType_ == 2 || sourceType_ == 0) {
        param1.userData_ = this;
        param1.name_ = "video";
        param1.feedCbOpaqueArg_ = this;
        param1.formatHint_ = "h264";
        param1.getFrameCb_ = IcePlayer::getFrameCallback;


        param2.userData_ = this;
        param2.name_ = "audio";
        param2.formatHint_ = "mulaw";
        param2.getFrameCb_ = IcePlayer::getFrameCallback;
        param2.feedCbOpaqueArg_ = this;
        param2.audioOpts.push_back("ar");
        param2.audioOpts.push_back("8000");


        param1.feedDataCb_ = videoFeed;
        param2.feedDataCb_ = audioFeed;
    }

    if(sourceType_ == 0) {
        param1.feedDataCb_ = feedFrameCallbackVideo;
        param2.feedDataCb_ = feedFrameCallbackAudio;
    }

    if (sourceType_ == 1) {
        param1.userData_ = this;
        param1.name_ = "test";
        param1.url_ = "rtmp://live.hkstv.hk.lxdns.com/live/hks";
        param1.getFrameCb_ = IcePlayer::getFrameCallback;
    }

    loginfo("start stream1 and stream2");
    m_stream1 = std::make_shared<Input>(param1);
    m_stream1->Start();

    if(sourceType_ == 0 || sourceType_ == 2) {
        m_stream2 = std::make_shared<Input>(param2);
        m_stream2->Start();
    }
}

void IcePlayer::hangup(){
    qDebug()<<"hangup invoked";

    StopStream();
    if (sourceType_ == 0) {
        if (iceSource_.get() != nullptr) {
            qDebug()<<"hangup call";
            iceSource_->hangup();
            iceSource_->Reset();
        }
    }
    if(audioFile.isOpen()) {
        audioFile.close();
    }
    if(videoFile.isOpen()) {
        videoFile.close();
    }

    m_vRenderer->ClearFrame();
    window()->update();

    mutex_.lock();
    m_aRenderer.Uninit();
    firstFrameTime_ = 0;
    canRender_ = false;
    Abuffer_.clear();
    Vbuffer_.clear();
    mutex_.unlock();

    loginfo("--------------flush----------------");
    logger_flush();
}

//audio
int IcePlayer::feedFrameCallbackAudio(void *opaque, uint8_t *buf, int buf_size)
{
    IcePlayer *p = (IcePlayer*)opaque;
    std::shared_ptr<std::vector<uint8_t>> audioData;
    int times = 5;
    while(times > 0) {
        auto s = p->iceSource_.get();
        if (s == nullptr) {
            return -1;
        }
        audioData = s->PopAudioData();
        if (audioData.get() == nullptr) {
            sleep(1);
            times--;
        } else {
            memcpy(buf, audioData->data(), audioData->size());
            return audioData->size();
        }
    }
    return -1;
}

//video
int IcePlayer::feedFrameCallbackVideo(void *opaque, uint8_t *buf, int buf_size)
{
    IcePlayer *p = (IcePlayer*)opaque;
    int times = 5;
    while(true) {
        if (times == 0)
            return -1;
        if (p->buffer_.get() != nullptr) {
            auto storeSize = p->buffer_->size();
            if (storeSize > 0) {
                if (buf_size >= storeSize) {
                    std::copy(p->buffer_->begin(), p->buffer_->end(), buf);
                    p->buffer_->resize(0);
                    return storeSize;
                } else {
                    std::copy(p->buffer_->begin(), p->buffer_->begin() + buf_size, buf);
                    std::copy(p->buffer_->begin() + buf_size, p->buffer_->end(), p->buffer_->begin());
                    p->buffer_->resize(storeSize - buf_size);
                    return buf_size;
                }
            }
        }

        std::shared_ptr<std::vector<uint8_t>> videoData;
        while(times > 0) {
            auto s = p->iceSource_.get();
            if (s == nullptr) {
                return -1;
            }
            videoData = s->PopVideoData();
            if (videoData.get() == nullptr) {
                sleep(1);
                times--;
            } else {
                p->buffer_ = videoData;
                break;
            }
        }
    }
    return -1;
}

void IcePlayer::push(std::shared_ptr<MediaFrame> && frame)
{
    if(firstFrameTime_ == 0) {
        firstFrameTime_ = os_gettime_ms();
    }
    if(frame->GetStreamType() == STREAM_AUDIO) {
        Abuffer_.emplace_back(frame);
    } else {
        Vbuffer_.emplace_back(frame);
    }
}
