#include <QGuiApplication>
#include <QQmlApplicationEngine>
#include "iceplayer.h"
#define THIS_FILE "qmlmain.cpp"

static Media mediaConfig[3] = {
    {STREAM_AUDIO, CODEC_G711U, 8000 ,1},
    {STREAM_AUDIO, CODEC_G711A, 8000 ,1},
    {STREAM_VIDEO, CODEC_H264, 90000, 0}
};
static int mediaLength = 3;

int main(int argc, char *argv[])
{
    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling);

    QGuiApplication app(argc, argv);

    qmlRegisterType<IcePlayer>("OpenGLUnderQML", 1, 0, "IcePlayer");

    logger_init_file_output("player.log");
    logger_set_level_debug();
    //logger_set_level_trace();

    ErrorID err = InitSDK(mediaConfig, mediaLength);
    if (err != RET_OK) {
        logerror("InitSDK fail:{}", err);
        return (int)err;
    }
    SetLogLevel(4);

    QQmlApplicationEngine engine;
    engine.load(QUrl(QStringLiteral("qrc:/main.qml")));
    if (engine.rootObjects().isEmpty())
        return -1;

    int ret = app.exec();
    UninitSDK();

    ThreadCleaner::GetThreadCleaner()->Stop();


    return ret;
}
