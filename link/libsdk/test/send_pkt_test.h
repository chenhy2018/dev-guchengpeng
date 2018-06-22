// Last Update:2018-06-21 12:08:17
/**
 * @file send_pkt_test.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-06-14
 */

#ifndef SEND_PKT_TEST_H
#define SEND_PKT_TEST_H


TestSuit gSendPacketTestSuit;
#define MAX_CALLS 16
#define HOST "123.59.204.198"
#define EVENT_16_CHANNEL_OK 0x1124
#define EVENT_16_CHANNEL_CHECK_OK 0x01
#define EVENT_16_CHANNEL_CHECK_FAIL 0x02
#define CALL_NOT_FOUND -1
#define DATA_CHECK_ERROR -2
#define EVENT_NOT_COMPLETE -3

#endif  /*SEND_PKT_TEST_H*/
