#include "fastfwd.hpp"

unsigned int fastfwd::global::nLogLevel = 4;
int fastfwd::nStatPeriod = 10;

int main(int _nArgc, char** _pArgv)
{
        if (_nArgc != 3) {
                Error("%s <url> <speed>\n", _pArgv[0]);
                exit(1);
        }

        std::string url = _pArgv[1];
        std::string xspeed = _pArgv[2];
        int x = fastfwd::x2;
        if (xspeed.compare("x2") == 0) {
                x = fastfwd::x2;
        } else if (xspeed.compare("x4") == 0) {
                x = fastfwd::x4;
        } else if (xspeed.compare("x8") == 0) {
                x = fastfwd::x8;
        } else if (xspeed.compare("x16") == 0) {
                x = fastfwd::x16;
        } else if (xspeed.compare("x32") == 0) {
                x = fastfwd::x32;
        } else {
                Error("supported speed: x2, x4, x8, x16, x32");
                exit(1);
        }

        {
                auto i = 0;
                auto pPumper = std::make_unique<fastfwd::StreamPumper>(url, x, 4096);

                std::ofstream outfile ("demo.mp4",std::ofstream::binary);

                std::vector<char> buffer;
                while (true) {
                        auto nStatus = pPumper->Pump(buffer, 5 * 1000);
                        if (nStatus == 0) {
                                // now, we are receiving buffer.size() bytes of buffer.data()
                                outfile.write(buffer.data(), buffer.size());
                        } else if (nStatus == fastfwd::StreamPumper::eof) {
                                Info("end of file");
                                return 0;
                        } else {
                                Info("error on connection");
                        }
                        i++;
                        Debug("count=%d", i);
                        if (i > 1500) {
                                break;
                        }
                }
        }

        sleep(3);

        return 0;
}

