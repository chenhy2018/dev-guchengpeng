// Last Update:2018-06-03 20:20:48
/**
 * @file framework.h
 * @brief 
 * @author liyq
 * @version 0.1.00
 * @date 2018-05-31
 */

#ifndef FRAMEWORK_H
#define FRAMEWORK_H

SipAnswerCode cbOnIncomingCall(IN const int _nAccountId, IN const int _nCallId,
                               IN const char *pFrom, IN const void *pUser, IN const void *pMedia);
void cbOnRegStatusChange(IN const int nAccountId, IN const SipAnswerCode RegStatusCode, IN const void *pUser);
void cbOnCallStateChange(IN const int nCallId, IN const int nAccountId, IN const SipInviteState State,
                         IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void *pMedia);

#endif  /*FRAMEWORK_H*/
