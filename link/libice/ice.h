#ifndef __ICE_H__
#define __ICE_H__

#include <pjlib.h>
#include <pjlib-util.h>
#include <pjnath.h>
#include <pjmedia.h>

#define IN
#define OUT

#define MAX_NAMESERVER_SIZE 128
#define MAX_STUN_HOST_SIZE  128
#define MAX_TURN_HOST_SIZE  128
#define MAX_TURN_USR_SIZE   64
#define MAX_TURN_PWD_SIZE   64
#define MAX_ICE_USRPWD_SIZE 80

typedef struct IceInstance IceInstance;
typedef struct IceOptions IceOptions;

struct IceOptions
{
	unsigned     nComponents;
	char         nameserver[MAX_NAMESERVER_SIZE];
	int          nMaxHosts;
	int          bRegular;
	char         stunHost[MAX_STUN_HOST_SIZE];
	char         turnHost[MAX_TURN_HOST_SIZE];
	int          bTurnTcp;
	char         turnUsername[MAX_TURN_USR_SIZE];
	char         turnPassword[MAX_TURN_PWD_SIZE];
	int          nKeepAlive;

	/* NEVER manipulate the ice instance ptr in callback functions */

	void (*onRxData)(IN pj_ice_strans* pStreamTransport, IN unsigned nComponentId, IN void* pData,
			 IN pj_size_t size, IN const pj_sockaddr_t* pAddr, IN unsigned nAddrLen);

	/* stage: 0 -> init, 1 -> negotiation, 2 -> keep alive, 3 -> addr change */
	/* status: 0 -> succeeded, -1 -> failed */
	void (*onIceComplete)(IN pj_ice_strans *pStreamTransport, IN pj_ice_strans_op op, IN pj_status_t status);
};

struct IceInstance
{
	/* memory management */
        pj_caching_pool      cachingPool;
        pj_pool_t*           pPool;
        pj_thread_t*         pThread;
        pj_bool_t            bThreadQuitFlag;
        pj_ice_strans_cfg    iceConfig;
        pj_ice_strans*       pIceStreamTransport;
	struct IceOptions    options;

	/* will be filled after session is encoded */
	struct LocalInfo
	{
		char                 ufrag[MAX_ICE_USRPWD_SIZE];
		char                 password[MAX_ICE_USRPWD_SIZE];
		unsigned             nCandidateCount;
		pj_ice_sess_cand     candidates[PJ_ICE_ST_MAX_CAND];
	} local;

	/* remote info */
	struct RemoteInfo
	{
		char                     ufrag[MAX_ICE_USRPWD_SIZE];
		char                     password[MAX_ICE_USRPWD_SIZE];
		unsigned                 nComponentCount;
		unsigned                 nCandidateCount;
		pj_ice_sess_cand         candidates[PJ_ICE_ST_MAX_CAND];
	} remote;
};

/* step 1 : create ice instance */
extern IceInstance* IceCreateInstance(IN const struct IceOptions* pOption);
extern void IceDestroyInstance(OUT IceInstance** pInstance);

/* step 2 : create ice session */
extern int IceSetOfferer(IN IceInstance* pInstance);
extern int IceSetAnswerer(IN IceInstance* pInstance);
extern int IceStopSession(IN IceInstance* pInstance);

/* step 3 : input remote info */
extern int IceGetRemote(IN IceInstance* pInstance, IN pjmedia_sdp_media* pMedia);

/* step 4 : start negotiation and rx/tx */
extern int IceStartNegotiation(IN IceInstance* pInstance);

/* step 5 : send data to peer */
extern int IceSendToPeer(IN IceInstance* pInstance, IN unsigned nComponentId, IN const char* pData, IN size_t nSize);

/* util */
extern void IceCandidateString(IN const pj_ice_sess_cand* pCandidate, OUT char* pBuffer, IN size_t nSize);

#endif
