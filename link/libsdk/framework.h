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

SipAnswerCode cbOnIncomingCall(const int _nAccountId, const const char *_pFrom, const void *_pUser,
                               IN const void *_pMedia, OUT int *pCallId);
void cbOnRegStatusChange(IN const int nAccountId, IN const SipAnswerCode RegStatusCode, IN const void *pUser);
void cbOnCallStateChange(IN const int nCallId, IN const int nAccountId, IN const SipInviteState State,
                         IN const SipAnswerCode StatusCode, IN const void *pUser, IN const void *pMedia);

#endif  /*FRAMEWORK_H*/
