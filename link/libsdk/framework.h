// Last Update:2018-05-31 15:53:41
/**
 * @file framework.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-05-31
 */

#ifndef FRAMEWORK_H
#define FRAMEWORK_H

SipAnswerCode cbOnIncomingCall(int _nAccountId, int _nCallId, const char *_pFrom);
void cbOnRegStatusChange(int _nAccountId, SipAnswerCode _StatusCode);
void cbOnCallStateChange(int _nCallId, IN const int _nAccountId, SipInviteState _State, SipAnswerCode _StatusCode);

#endif  /*FRAMEWORK_H*/
