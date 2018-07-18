// Last Update:2018-06-08 20:13:25
/**
 * @file test.c
 * @brief 
 * @author
 * @version 0.1.00
 * @date 2018-06-05
 */
#include <string.h>
#include <stdio.h>
#ifdef WITH_P2P
#include "sdk_interface_p2p.h"
#else
#include "sdk_interface.h"
#endif
#include "dbg.h"
#include "unit_test.h"
#include "test.h"
#include <unistd.h> 
#include <stdlib.h>
#include <pthread.h>

#define ARRSZ(arr) (sizeof(arr)/sizeof(arr[0]))
#define HOST "180.97.147.176"
#define INVALID_SERVER "192.168.1.239"

#define MAX_COUNT 145
#define CHENGPENG 1
//#define MENGKE 1
int RegisterTestSuitCallback( TestSuit *this );
int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager );
int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase );

typedef struct {
    char *id;
    char *password;
    char *sigHost;
    char *mediaHost;
    char *imHost;
    int timeOut;
    unsigned char init;
    int accountid;
    int callid;
    int sendflag;
    int64_t timecount;
} RegisterData;

typedef struct {
    TestCase father;
    RegisterData data;
} RegisterTestCase;

#ifdef CHENGPENG
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1900", "dp3PAQ4k", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1900", "b7Ur3mVq", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1902", "XGCKX9gw", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1903", "HtqHl1cp", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1904", "kBncoGjn", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1905", "ZsSG7ZB4", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1906", "wrBc2kAB", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1907", "lUz1Ja3d", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1908", "mxtjoUlF", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1909", "DHZtZNRW", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1910", "uYbGD6vX", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1911", "WbfL9mt4", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1912", "3g8oRVn3", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1913", "uu1KjeO6", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1914", "HBlgmuUV", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1915", "kI2VzO44", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1916", "lMdEcKgn", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1917", "IoVhdjCV", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1918", "iH7aP2cN", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1919", "FXNoBJ2M", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1920", "BcyArZSU", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1921", "Esw7fR35", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1922", "ZNEwN6jo", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1923", "mBYhHieI", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1924", "qOUB0RmD", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1925", "7d5FFDCp", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1926", "Ta1KxOjt", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1927", "G8ThrckC", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1928", "iidDQszx", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1929", "avMoWjK8", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1930", "wBFCJBJV", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1931", "xcy3cPlO", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1932", "EB7KGEtV", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1933", "9jarO2Ng", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1934", "koBMp02K", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1935", "lmecN6ya", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1936", "1QhgElCw", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1937", "DihLecIY", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1938", "mlqeJS5z", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1939", "U9L6nEvz", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1940", "X4xux7EL", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1941", "LR3F373n", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1942", "gwIAajyZ", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1943", "clmRt82H", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1944", "UrKDcN3V", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1945", "wyQoaWiH", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1946", "3p1CqXJj", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1947", "tE3OfNX1", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1948", "bYeAfkIW", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1949", "73If5Lvh", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1950", "CSyQNII0", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1951", "f4IAYbjB", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1952", "TMo7XQvC", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1953", "BgGkHVwA", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1954", "130ShM6v", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1955", "uGf4bR7L", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1956", "O4zgz9CP", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1957", "9QGWp0vN", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1958", "buP66Iqe", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1959", "BdnTbxTu", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1960", "nz0cbU7Z", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1961", "6hjLrX0a", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1962", "ufHshieB", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1963", "utlRRC5t", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1964", "ZFFUBnGC", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1965", "8TdNNjzv", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1966", "8TdNNjzv", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1967", "75KCijq0", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1968", "HwuEyvhw", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1969", "zp75xdtU", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1970", "QDQX3IqU", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1971", "IkV9nZ2k", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1972", "zH4hd8qs", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1973", "3frhTLl2", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1974", "hMaTiQpo", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1975", "ZeMJq0F2", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1976", "64XC92PL", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1977", "ZuWPWEFt", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1978", "qF4sRUhR", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1979", "mzfHr8va", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1980", "1tom06ip", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1981", "mJPTF016", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1982", "9uAcic2M", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1983", "KgLWDyV1", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1984", "4Q87v3xA", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1985", "EaSvQqfs", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1986", "xDK92Vva", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1987", "wvA2buiT", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1988", "cHA8OEYD", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1989", "P7tm32yE", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1990", "gkom1sOP", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1991", "n40ffYR3", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1992", "hAsHOK3M", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "1993", "aggblq2S", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "1994", "0JsucanN", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1995", "Sh4vkkjJ", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1996", "SXEzuYoh", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1997", "OOrrS5mk", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1998", "vMy0h0lI", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "1999", "164AsPtd", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2000", "eBNd1LV4", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2001", "AvrgH71U", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2002", "hSv5U2Pk", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2003", "6qPlV8RO", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2004", "PIBL20Rn", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2005", "R0Sdt530", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "2006", "yACIsMXH", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2007", "4uTeU9U5", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2008", "cH7e95GP", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2009", "Btb5oj4I", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2010", "GbU6cXdd", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2011", "AFO5KRro", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2012", "8VHGIBfJ", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2013", "xBYCVYvV", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2014", "1ok2OwBH", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "2015", "7JNVZLCR", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2016", "EIo5Fejk", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2017", "ANVQzAQX", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2018", "IOQcFDjT", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2019", "44OtVolb", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2020", "y4qmaGEd", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2021", "vdK3TsK0", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2022", "Ua2tTBq0", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2023", "vkFjEJqp", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "2024", "xBlVPQp0", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2025", "wBZtu6xG", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2026", "36mERLrR", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2027", "WaejUS7N", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2028", "6ZcRv5ou", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2029", "Xdd2aapa", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2030", "KR0BGS8L", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2031", "xILswFie", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2032", "Dkx59bGx", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "2033", "fMDXAsmr", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2034", "o8nd63qd", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2035", "vtWfEJLg", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2036", "OQwmWUiv", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2037", "uJ8wikjq", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2038", "WZx6T3Er", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2039", "ZjTW0n8n", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "normal", 0 },
        { "2040", "vGrMmyIs", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "valid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2041", "LDb1wHu9", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2042", "qIwCXfUp", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTERED, NULL },
        { "2043", "IZYkxp5O", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTERED, NULL },
        { "2044", "WZx6T3Er", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTERED, NULL},
        { "2045", "vdK3TsK0", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "2046", "I1gj5Wef", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2047", "rx7HOnDT", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1899", "164AsPtd", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "2049", "1NR6D190", HOST, HOST, HOST, 100, 1 }
    },
};
#endif

#ifdef MENGKE
RegisterTestCase gRegisterTestCases[] =
{
    {
        { "valid_account1", CALL_STATUS_REGISTERED, UA1_EventLoopThread },
        { "1711", "aUSEOnOy", HOST, HOST, HOST, 10, 1 }
    },
    {
        { "valid_account2", CALL_STATUS_REGISTERED, UA2_EventLoopThread },
        { "1712", "Q0EEBOEc", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account1", CALL_STATUS_REGISTER_FAIL, UA3_EventLoopThread },
        { "1713", "IeFxv0sP", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_account2", CALL_STATUS_REGISTER_FAIL, UA4_EventLoopThread },
        { "1714", "9ZykZwsJ", HOST, HOST, HOST, 10, 0 }
    },
    {
        { "invalid_sip_register_server", CALL_STATUS_REGISTER_FAIL, UA5_EventLoopThread },
        { "1715", "AzkaVAo0", INVALID_SERVER, INVALID_SERVER, INVALID_SERVER, 10, 0 }
    },
    {
        { "invalid_account", 0 },
        { "1716", "5XPM9DUv", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1717", "W9DaI77R", HOST, HOST, HOST, 100, 1 }
    },
    {
        { "invalid_account", 0 },
        { "1718", "4JoChsXl", HOST, HOST, HOST, 100, 0 }
    },
    {
        { "normal", 0 },
        { "1719", "NVWrpASp", HOST, HOST, HOST, 100, 1 }
    },
};
#endif
TestSuit gRegisterTestSuit =
{
    "Register",
    RegisterTestSuitCallback,
    RegisterTestSuitInit,
    RegisterTestSuitGetTestCase,
    (void*)&gRegisterTestCases,
    1
};

int RegisterTestSuitInit( TestSuit *this, TestSuitManager *_pManager )
{
    this->total = ARRSZ(gRegisterTestCases);
    this->index = 0;
    this->pManager = _pManager;

    return 0;
}

int RegisterTestSuitGetTestCase( TestSuit *this, TestCase **testCase )
{
    RegisterTestCase *pTestCases = NULL;

    if ( !testCase || !this ) {
        return -1;
    }

    pTestCases = (RegisterTestCase *)this->testCases;
    *testCase = (TestCase *)&pTestCases[this->index];

    return 0;
}

void Mthread1(void* data)
{   
    RegisterData *pData = (RegisterData*) data;
    Event * event= (Event*) malloc(sizeof(Event));
    ErrorID id = 0;
    EventType type;
    int timecount = 0;
    int sendcount = 0;
    int sendflag = 0;
    int misscount = 0;
    int discount = 0;
    char data1[100] = {1};
    DBG_LOG("send pack *****%d call id %d\n", pData->accountid, pData->callid);
    while (1) {     
                    if (sendflag) {
                            if (sendcount > 500) {
                                    misscount = sendcount - timecount;
                                    DBG_LOG("UnRegister *******send count %d recv count %d********* misscount %d \n", sendcount, timecount, misscount);
                                    UnRegister(pData->accountid);
#ifdef WITH_P2P
                                    pData->accountid = Register(pData->id, pData->password, NULL, NULL, pData->imHost);
#else
                                    pData->accountid = Register(pData->id, pData->password, NULL, pData->imHost);
#endif
                                    sendflag = 0;
                                    sendcount = 0;
                                    timecount = 0;
                                    continue;
                            }
                            Report(pData->accountid, pData->id, "test11111",10); 
                            sendcount += 1;
                            usleep(100000);
                            //continue;
                    }
                    id = PollEvent(pData->accountid, &type, &event, 100);
                    if (id != RET_OK) {
                           sleep(1);
                           continue;
                    }
                    switch (type) {
                            case EVENT_CALL:
                            {
                                  CallEvent *pCallEvent = &(event->body.callEvent);
                                  DBG_LOG("Call status %d call id %d call account id %d\n", pCallEvent->status, pCallEvent->callID, pData->accountid);
                                  if (pCallEvent->status == CALL_STATUS_INCOMING) {
                                      DBG_LOG("AnswerCall ******************\n");
                                      AnswerCall(pData->accountid, pCallEvent->callID);
                                      DBG_LOG("AnswerCall end *****************\n");
                                  }
                                  if (pCallEvent->status == CALL_STATUS_ERROR || pCallEvent->status == CALL_STATUS_HANGUP) {
                                        DBG_LOG("makecall **************************** reason %d\n", pCallEvent->reasonCode);
#if 0
                                        do {
                                                sleep(1);
                                                id = MakeCall(pData->accountid, "2040", HOST, &pData->callid);
                                        } while (id != RET_OK);
#endif
                                  }
                                  if (pCallEvent->status == CALL_STATUS_ESTABLISHED) {
#ifdef WITH_P2P
                                        MediaInfo* info = (MediaInfo *)pCallEvent->context;
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d mediacount %d, type 1 %d type 2 %d\n",
                                                 pCallEvent->callID, pData->accountid, info->nCount, info->media[0].codecType, info->media[1].codecType);
#else
                                        DBG_LOG("CALL_STATUS_ESTABLISHED call id %d account id %d \n", pCallEvent->callID, pData->accountid);
#endif                                        
                                        sendflag = 1;
                                  }

                                  break;
                            }
#ifdef WITH_P2P
                            case EVENT_DATA:
                            {     
                                  DataEvent *pDataEvent = &(event->body.dataEvent);
                                  //DBG_LOG("Data size %d call id %d call account id %d timestamp %lld \n", pDataEvent->size, pDataEvent->callID, pData->accountid, pDataEvent->pts);
                                  break;
                            }
#endif
                            case EVENT_MESSAGE:
                            {
                                  MessageEvent *pMessage = &(event->body.messageEvent);
                                  //DBG_LOG("Message %s status id %d account id %d\n", pMessage->message, pMessage->status, pData->accountid);
                                  if (MESSAGE_STATUS_CONNECT == pMessage->status) {
                                          sendflag = 1;
                                          //DBG_LOG("Subscribe test/test\n");
                                          Subscribe(pData->accountid, pData->id);
                                          DBG_LOG("Message %s status id %d account id %d\n", pMessage->message, pMessage->status, pData->accountid);
                                  } else if (MESSAGE_STATUS_DATA == pMessage->status) {
                                          //timecount += 1;
                                          if (timecount % 10 == 0) {
                                                 //DBG_LOG("Message %s status id %d account id %d\n", pMessage->message, pMessage->status, pData->accountid);
                                          }
                                          ++timecount;
                                  }
                                  else if (MESSAGE_STATUS_DISCONNECT == pMessage->status) {
                                          ++discount;
                                          DBG_LOG("MESSAGE_STATUS_DISCONNECT id %s count %d \n", pData->id, discount);
                                  }
                                  break;
                            }
                    }
           }
}

int RegisterTestSuitCallback( TestSuit *this )
{
    RegisterTestCase *pTestCases = NULL;
    RegisterData *pData = NULL;
    Media media[2];
    media[0].streamType = STREAM_VIDEO;
    media[0].codecType = CODEC_H264;
    media[0].sampleRate = 90000;
    media[0].channels = 0;
    media[1].streamType = STREAM_AUDIO;
    media[1].codecType = CODEC_G711A;
    media[1].sampleRate = 8000;
    media[1].channels = 1;
    int i = 0;
    ErrorID sts = 0;

    if ( !this ) {
        return -1;
    }
    int all_count = 0;
    pTestCases = (RegisterTestCase *) this->testCases;
    UT_LOG("this->index = %d\n", this->index );
    pData = &pTestCases[this->index].data;

    if ( pData->init ) {
        UT_LOG("InitSDK");
        sts = InitSDK( &media[0], 2 );
        if ( RET_OK != sts ) {
            UT_ERROR("sdk init error\n");
            return TEST_FAIL;
        }
    }
    setPjLogLevel(2);
    UT_STR( pData->id );
    for (int count = 0; count < MAX_COUNT; ++ count) {
            pData = &pTestCases[count].data;
            UT_STR(pData->password);
            UT_STR(pData->sigHost);
            UT_LOG("Register in\n");
#ifdef WITH_P2P
            pData->accountid = Register(pData->id, pData->password, NULL, NULL, pData->imHost);
#else
            pData->accountid = Register(pData->id, pData->password, NULL, pData->imHost);
#endif
            UT_LOG("Register out %x %x\n", pData->accountid, pTestCases->father.expact);
            int nCallId1 = -1;
    }
    sleep(10);

        pthread_t t_1;
        pthread_attr_t attr_1;
        pthread_attr_init(&attr_1);
        pthread_attr_setdetachstate(&attr_1, PTHREAD_CREATE_DETACHED);

    Event* event = (Event*) malloc(sizeof(Event));
    ErrorID id;
    for (int count = 0; count < MAX_COUNT; ++ count) {
           pData = &pTestCases[count].data;
           pData->sendflag = 0;
           pData->timecount = 0;
           UT_LOG("MakeCall in accountid %d\n", pData->accountid);
           pData->callid = 0;
#if 0
           id = MakeCall(pData->accountid, "2040", HOST, &pData->callid);
           if (RET_OK != id) {
                    fprintf(stderr, "call error %d \n", id);
                     continue;
           }
#endif
           UT_LOG("MakeCall in callidid %d\n", pData->callid);
           pthread_create(&t_1, &attr_1, Mthread1, pData);
    }
    while(1) { sleep(10); }
}

int InitAllTestSuit()
{
    AddTestSuit( &gRegisterTestSuit );
    return 0;
}

int main()
{
    UT_LOG("+++++ enter main...\n");
    TestSuitManagerInit();
    InitAllTestSuit();
    RunAllTestSuits();
    return 0;
}

