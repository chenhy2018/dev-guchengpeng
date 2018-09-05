#ifndef ICEPLAYER_H
#define ICEPLAYER_H

#include <QtQuick/QQuickItem>
#include <QtGui/QOpenGLShaderProgram>
#include <QtGui/QOpenGLFunctions>
#include <QtMultimedia>
#include <QTimer>
#include <memory>
#include <input.hpp>
#include "linking.h"
#include "ThreadCleaner.h"

class QGLRenderer : public QObject, protected QOpenGLFunctions
{
    Q_OBJECT
    friend class IcePlayer;
public:
    QGLRenderer();
    ~QGLRenderer();

    void setViewportSize(const QSize &size) { m_viewportSize = size; }
    void setWindow(QQuickWindow *window) { m_window = window; }
    void setDrawRect(QRectF & s, QRectF &it);

    const int TextureYIndex = 0;
    const int TextureUIndex = 1;
    const int TextureVIndex = 2;
    void SetFrame(std::shared_ptr<MediaFrame> &frame);
    void ClearFrame();

public slots:
    void paint();

private:
    QSize m_viewportSize;
    QOpenGLShaderProgram *m_program;
    QQuickWindow *m_window;
    GLuint m_textures[3];
    GLfloat xleft;
    GLfloat xright;
    GLfloat ytop;
    GLfloat ybottom;

    std::mutex m_frameMutex;
    std::shared_ptr<MediaFrame> m_frame;
};

class AudioRender{
public:
    const static int ulawType = 1;
    const static int alawType = 2;
    AudioRender();
    void PushData(void *pcmData,int size);
    void PushG711Data(void *g711Data, int size, int lawType);
    void Init(QAudioFormat config);
    void Init(std::shared_ptr<MediaFrame> & frame);
    void Uninit();
    bool IsInited(){return m_inited;}
private:
    QAudioFormat m_audioConfig;
    std::shared_ptr<QAudioOutput> m_audioOutput;
    bool m_inited = false;
    QIODevice *m_device;
};

class IcePlayer : public QQuickItem
{
    Q_OBJECT

public:
    explicit IcePlayer();
    ~IcePlayer();

signals:
    void tChanged();
    void pictureReady();
    void getFirstAudioPktTime(QString timestr);
    void getFirstVideoPktTime(QString timestr);
    void streamInfoUpdate(QString infoStr);

public slots:
    void sync();
    void cleanup();
    void Stop();
    void repaint();
    void setSourceType(QVariant stype);

    void call(QVariant sipAccount);
    void makeCall();
    void hangup();

    void firstAudioPktTime(QString);
    void firstVideoPktTime(QString);

private slots:
    void handleWindowChanged(QQuickWindow *win);
    void updateStreamInfo();

private:
    static void getFrameCallback(void * userData, std::shared_ptr<MediaFrame> & frame);
    static int feedFrameCallbackAudio(void *, uint8_t *buf, int buf_size);
    static int feedFrameCallbackVideo(void *, uint8_t *buf, int buf_size);
    void push(std::shared_ptr<MediaFrame> && frame);
    void StopStream();

private:
    qreal m_t;
    QGLRenderer *m_vRenderer;
    AudioRender m_aRenderer;
    bool quit_ = false;

    std::shared_ptr<Input> m_stream1;
    std::shared_ptr<Input> m_stream2;

    std::shared_ptr<linking> iceSource_;
    std::shared_ptr<std::vector<uint8_t>> buffer_; //视频帧ffmpeg一次读取不完，所以需要记录下来下次读
    bool registerOk;
    int sourceType_ = 0; // 0 sip, 1 file, 2 h264/pcmu file


    //同步线程，player上做一个缓冲，同步音频和视频
    //同步：对齐到同一个墙上时钟，按照各自当前时间渲染
    int64_t firstFrameTime_ = 0; //可能是音频也可能是视频
    bool canRender_ = false;
    std::thread avsync_;
    std::mutex mutex_;
    std::condition_variable condition_;
    std::deque<std::shared_ptr<MediaFrame>> Abuffer_;
    std::deque<std::shared_ptr<MediaFrame>> Vbuffer_;

    std::mutex streamMutex_;
    std::thread timerThread_;
};

#endif // ICEPLAYER_H
