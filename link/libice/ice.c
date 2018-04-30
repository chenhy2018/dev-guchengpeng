#include "ice.h"

void IceDestroyInstance(OUT IceInstance** _pInstance)
{
	IceInstance* pInstance = *_pInstance;

        /* destroy stream transport instance */
        if (pInstance->pIceStreamTransport != NULL) {
                pj_ice_strans_destroy(pInstance->pIceStreamTransport);
        }

        /* sleep 500 ms */
        pj_thread_sleep(500);
  
        /* wait for WorkerThread() to exit */
        pInstance->bThreadQuitFlag = PJ_TRUE;
        if (pInstance->pThread != NULL) {
                pj_thread_join(pInstance->pThread);
                pj_thread_destroy(pInstance->pThread);
        }

        /* destroy ioqueue that WorkerThread() is polling */
        if (pInstance->iceConfig.stun_cfg.ioqueue)
                pj_ioqueue_destroy(pInstance->iceConfig.stun_cfg.ioqueue);
  
        /* destroy timer heap */
        if (pInstance->iceConfig.stun_cfg.timer_heap)
                pj_timer_heap_destroy(pInstance->iceConfig.stun_cfg.timer_heap);
  
        /* destory ice instance caching pool */
        pj_caching_pool_destroy(&pInstance->cachingPool);

        pj_shutdown();

        /* free instance ptr */
        free(*_pInstance);
        *_pInstance = NULL;
}

static int IceWorkerThread(IN void* _pInstance)
{
        const IceInstance* pInstance = (IceInstance*)_pInstance;

        while (!pInstance->bThreadQuitFlag) {
                enum { MAX_NET_EVENTS = 1 };
                pj_time_val maxTimeout = {0, 0};
                pj_time_val timeout = {0, 0};
                unsigned nTotalCount = 0, nNetEventCount = 0;
                int nCount;

                /* max ioqueue poll timeout */
                maxTimeout.msec = 500;
  
                /* Poll the timer to run it and also to retrieve the earliest entry. */
                timeout.sec = timeout.msec = 0;
                nCount = pj_timer_heap_poll(pInstance->iceConfig.stun_cfg.timer_heap, &timeout);
                if (nCount > 0) {
                        nTotalCount += nCount;
                }
  
                /* timer_heap_poll should never ever returns negative value, or otherwise
                 * ioqueue_poll() will block forever!
                 */
                pj_assert(timeout.sec >= 0 && timeout.msec >= 0);
                if (timeout.msec >= 1000) {
                        timeout.msec = 999;
                }
  
                /* compare the value with the timeout to wait from timer, and use the
                 * minimum value.
                 */
                if (PJ_TIME_VAL_GT(timeout, maxTimeout)) {
                        timeout = maxTimeout;
                }
  
                /* Poll ioqueue */
                do {
                        nCount = pj_ioqueue_poll(pInstance->iceConfig.stun_cfg.ioqueue, &timeout);
                        if (nCount < 0) {
                                pj_status_t err = pj_get_netos_error();
                                pj_thread_sleep(PJ_TIME_VAL_MSEC(timeout));
                                return err;
                        } else if (nCount == 0) {
                                break;
                        } else {
                                nNetEventCount += nCount;
                                timeout.sec = timeout.msec = 0;
                        }
                } while (nCount > 0 && nNetEventCount < MAX_NET_EVENTS);
  
                nTotalCount += nNetEventCount;
        }

        return 0;
}

static pj_status_t IceCreateStreamTransport(OUT IceInstance* _pInstance)
{
        pj_ice_strans_cb iceCallbacks;

        /* init the callback */
        pj_bzero(&iceCallbacks, sizeof(iceCallbacks));
        iceCallbacks.on_rx_data = _pInstance->options.onRxData;
        iceCallbacks.on_ice_complete = _pInstance->options.onIceComplete;
  
        /* create the instance */
        return pj_ice_strans_create(NULL,                                     /* object name  */
                                    &_pInstance->iceConfig,                   /* settings     */
                                    _pInstance->options.nComponents,          /* comp_cnt     */
                                    (void*)_pInstance,                        /* user data    */
                                    &iceCallbacks,                            /* callback     */
                                    &_pInstance->pIceStreamTransport)         /* instance ptr */
                ;
}

static void IceInitInstance(IN IceInstance* _pInstance, IN const struct IceOptions* _pOption)
{
        if (_pInstance == NULL) {
                return;
        }

        /* copy options */
        memcpy(&_pInstance->options, _pOption, sizeof(IceOptions));
        _pInstance->options.nameserver[MAX_NAMESERVER_SIZE - 1] = 0;
        _pInstance->options.stunHost[MAX_STUN_HOST_SIZE - 1] = 0;
        _pInstance->options.turnHost[MAX_TURN_HOST_SIZE - 1] = 0;
        _pInstance->options.turnUsername[MAX_TURN_USR_SIZE - 1] = 0;
        _pInstance->options.turnPassword[MAX_TURN_PWD_SIZE - 1] = 0;

        /* init ptrs */
        _pInstance->pPool = NULL;
        _pInstance->pThread = NULL;
        _pInstance->pIceStreamTransport = NULL;
}

IceInstance* IceCreateInstance(IN const struct IceOptions* _pOption)
{
        /* Initialize the libraries before anything else */
        pj_init();
        pjlib_util_init();
        pjnath_init();

        /* allocate one ice instance struct */
        IceInstance* pInstance = (IceInstance*)malloc(sizeof(IceInstance));
        if (pInstance == NULL) {
                return NULL;
        }
        IceInitInstance(pInstance, _pOption);

        pj_status_t status;
#define CHECK(expr) status = expr;                                \
        if (status != PJ_SUCCESS) {                               \
                IceDestroyInstance(&pInstance);                   \
                return NULL;                                      \
        }
  
        /* Must create pool factory, where memory allocations come from */
        pj_caching_pool_init(&pInstance->cachingPool, NULL, 0);

        /* Init our ICE settings with null values */
        pj_ice_strans_cfg_default(&pInstance->iceConfig);
  
        pInstance->iceConfig.stun_cfg.pf = &pInstance->cachingPool.factory;
  
        /* Create application memory pool */
        pInstance->pPool = pj_pool_create(&pInstance->cachingPool.factory, NULL, 512, 512, NULL);
  
        /* Create timer heap for timer stuff */
        CHECK(pj_timer_heap_create(pInstance->pPool, 100, &pInstance->iceConfig.stun_cfg.timer_heap));
  
        /* and create ioqueue for network I/O stuff */
        CHECK(pj_ioqueue_create(pInstance->pPool, 16, &pInstance->iceConfig.stun_cfg.ioqueue));
  
        /* something must poll the timer heap and ioqueue,
         * unless we're on Symbian where the timer heap and ioqueue run
         * on themselves.
         */
        CHECK(pj_thread_create(pInstance->pPool, NULL, &IceWorkerThread, (void*)pInstance, 0, 0, &pInstance->pThread));
  
        pInstance->iceConfig.af = pj_AF_INET();
  
        /* Create DNS resolver if nameserver is set */
        pj_str_t nameserver = pj_str(pInstance->options.nameserver);
        if (nameserver.slen) {
                CHECK(pj_dns_resolver_create(&pInstance->cachingPool.factory,
                                              "resolver",
                                              0,
                                              pInstance->iceConfig.stun_cfg.timer_heap,
                                              pInstance->iceConfig.stun_cfg.ioqueue,
                                              &pInstance->iceConfig.resolver));
                  CHECK(pj_dns_resolver_set_ns(pInstance->iceConfig.resolver, 1, &nameserver, NULL));
        }
  
        /* -= Start initializing ICE stream transport config =- */
  
        /* ICE */

        /* Maximum number of host candidates */
        if (pInstance->options.nMaxHosts != -1) {
                pInstance->iceConfig.stun.max_host_cands = pInstance->options.nMaxHosts;
        }
  
        /* Nomination strategy */
        if ((pj_bool_t)pInstance->options.bRegular) {
                pInstance->iceConfig.opt.aggressive = PJ_FALSE;
        } else {
                pInstance->iceConfig.opt.aggressive = PJ_TRUE;
        }
  
        /* STUN */
        pj_str_t stunHost = pj_str(pInstance->options.stunHost);
        if (stunHost.slen) {
                char *pPosition;
  
                /* Command line option may contain port number */
                if ((pPosition = pj_strchr(&stunHost, ':')) != NULL) {
                        pInstance->iceConfig.stun.server.ptr = stunHost.ptr;
                        pInstance->iceConfig.stun.server.slen = (pPosition - stunHost.ptr);
                          pInstance->iceConfig.stun.port = (pj_uint16_t)atoi(pPosition + 1);
                } else {
                        pInstance->iceConfig.stun.server = stunHost;
                        pInstance->iceConfig.stun.port = PJ_STUN_PORT;
                }
  
                /* set keep alive time */
                pInstance->iceConfig.stun.cfg.ka_interval = pInstance->options.nKeepAlive;
        }
  
        /* TURN */
        pj_str_t turnHost = pj_str(pInstance->options.turnHost);
        pj_str_t turnUsername = pj_str(pInstance->options.turnUsername);
        pj_str_t turnPassword = pj_str(pInstance->options.turnPassword);
        if (turnHost.slen) {
                char *pPosition;
  
                /* Command line option may contain port number */
                if ((pPosition=pj_strchr(&turnHost, ':')) != NULL) {
                        pInstance->iceConfig.turn.server.ptr = turnHost.ptr;
                        pInstance->iceConfig.turn.server.slen = (pPosition - turnHost.ptr);
                          pInstance->iceConfig.turn.port = (pj_uint16_t)atoi(pPosition + 1);
                } else {
                        pInstance->iceConfig.turn.server = turnHost;
                        pInstance->iceConfig.turn.port = PJ_STUN_PORT;
                }
  
                /* TURN credential */
                pInstance->iceConfig.turn.auth_cred.type = PJ_STUN_AUTH_CRED_STATIC;
                pInstance->iceConfig.turn.auth_cred.data.static_cred.username = turnUsername;
                pInstance->iceConfig.turn.auth_cred.data.static_cred.data_type = PJ_STUN_PASSWD_PLAIN;
                pInstance->iceConfig.turn.auth_cred.data.static_cred.data = turnPassword;
  
                /* Connection type to TURN server */
                if ((pj_bool_t)pInstance->options.bTurnTcp) {
                        pInstance->iceConfig.turn.conn_type = PJ_TURN_TP_TCP;
                } else {
                        pInstance->iceConfig.turn.conn_type = PJ_TURN_TP_UDP;
                }
  
                /* set turn refresh */
                pInstance->iceConfig.turn.alloc_param.ka_interval = pInstance->options.nKeepAlive;
        }
  
        /* create stream transport */
        CHECK(IceCreateStreamTransport(pInstance));

        /* -= That's it for now, initialization is complete =- */
        return pInstance;
}

static void IceResetRemoteInfo(IN IceInstance* _pInstance)
{
	pj_bzero(&_pInstance->remote, sizeof(_pInstance->remote));
}

static int IceEncodeSession(IN IceInstance* _pInstance)
{
	memset(&_pInstance->local, 0, sizeof(_pInstance->local));

	/* generate local ufrag and password */
	pj_str_t ufrag, pwd;
	pj_ice_strans_get_ufrag_pwd(_pInstance->pIceStreamTransport, &ufrag, &pwd, NULL, NULL);
	memcpy(&_pInstance->local.ufrag[0], ufrag.ptr, ufrag.slen);
	memcpy(&_pInstance->local.password[0], pwd.ptr, pwd.slen);

	/* fill candidates */
	for (int i = 0; i < _pInstance->options.nComponents; i++) {
		unsigned nCandidates = PJ_ICE_ST_MAX_CAND - _pInstance->local.nCandidateCount;
		pj_status_t status = pj_ice_strans_enum_cands(_pInstance->pIceStreamTransport, i + 1,
							      &nCandidates, &_pInstance->local.candidates[_pInstance->local.nCandidateCount]);
		if (status != PJ_SUCCESS) {
			return -2;
		}
		_pInstance->local.nCandidateCount += nCandidates;
	}

	/* print candidates */
	PJ_LOG(3, ("IceEncodeSession", "get local ufrag=%s, pwd=%s, candidates=%d",
		   _pInstance->local.ufrag, _pInstance->local.password, _pInstance->local.nCandidateCount));

	char ipaddr[PJ_INET6_ADDRSTRLEN];
	for (int j = 0; j < _pInstance->local.nCandidateCount; j++) {
		PJ_LOG(3, ("IceEncodeSession", "a=candidate:%.*s %u UDP %u %s %u typ %s\n",
			   (int)_pInstance->local.candidates[j].foundation.slen,
			   _pInstance->local.candidates[j].foundation.ptr,
			   (unsigned)_pInstance->local.candidates[j].comp_id,
			   _pInstance->local.candidates[j].prio,
			   pj_sockaddr_print(&_pInstance->local.candidates[j].addr, ipaddr, sizeof(ipaddr), 0),
			   (unsigned)pj_sockaddr_get_port(&_pInstance->local.candidates[j].addr),
			   pj_ice_get_cand_type_name(_pInstance->local.candidates[j].type)));
	}

	return 0;
}

static int IceSetSession(IN IceInstance* _pInstance, pj_ice_sess_role _role)
{
	if (_pInstance == NULL) {
		return -1;
	}

	/* return error if a session already exists */
	if (pj_ice_strans_has_sess(_pInstance->pIceStreamTransport)) {
		return -2;
	}

	pj_status_t status = pj_ice_strans_init_ice(_pInstance->pIceStreamTransport, _role, NULL, NULL);
	if (status != PJ_SUCCESS) {
		return -3;
	}

	int nStatus = IceEncodeSession(_pInstance);
	if (nStatus < 0) {
		return -4;
	}

	return 0;
}

int IceSetOfferer(IN IceInstance* _pInstance)
{
	int nStatus = IceSetSession(_pInstance, PJ_ICE_SESS_ROLE_CONTROLLING);
	if (nStatus < 0) {
		return nStatus;
	}

	IceResetRemoteInfo(_pInstance);

	return 0;
}

int IceSetAnswerer(IN IceInstance* _pInstance)
{
	int nStatus = IceSetSession(_pInstance, PJ_ICE_SESS_ROLE_CONTROLLED);
	if (nStatus < 0) {
		return nStatus;
	}

	IceResetRemoteInfo(_pInstance);

	return 0;
}

int IceStopSession(IN IceInstance* _pInstance)
{
	pj_status_t status;
  
	if (_pInstance == NULL) {
		return -1;
	}
  
	if (!pj_ice_strans_has_sess(_pInstance->pIceStreamTransport)) {
		return -2;
	}
  
	status = pj_ice_strans_stop_ice(_pInstance->pIceStreamTransport);
	if (status != PJ_SUCCESS) {
		return -3;
	}

	IceResetRemoteInfo(_pInstance);

	return 0;
}

int IceGetRemote(IN IceInstance* _pInstance, IN pjmedia_sdp_media* _pMedia)
{
	IceResetRemoteInfo(_pInstance);

	size_t nLength;
	for (int i = 0; i < _pMedia->attr_count; i++) {
		if (pj_strcmp2(&_pMedia->attr[i]->name, "ice-ufrag") == 0) {
			/* initialize ice username */
			nLength = _pMedia->attr[i]->value.slen > 80 ? 80 : _pMedia->attr[i]->value.slen;
			memcpy(_pInstance->remote.ufrag, _pMedia->attr[i]->value.ptr, nLength);
		} else if (pj_strcmp2(&_pMedia->attr[i]->name, "ice-pwd") == 0) {
			/* initialize ice password */
			nLength = _pMedia->attr[i]->value.slen > 80 ? 80 : _pMedia->attr[i]->value.slen;
			memcpy(_pInstance->remote.password, _pMedia->attr[i]->value.ptr, nLength);
		} else if (pj_strcmp2(&_pMedia->attr[i]->name, "candidate") == 0) {
			/* get candidate params */
			char foundation[32], transport[12], ipaddr[80], type[32];
			int nCompId, nPrio, nPort;
			int nCount = sscanf(_pMedia->attr[i]->value.ptr, "%s %d %s %d %s %d typ %s",
					    foundation, &nCompId, transport, &nPrio, ipaddr, &nPort, type);
			if (nCount != 7) {
				return -1;
			}

			/* initialize candidate array */
			pj_ice_sess_cand* pCandidate = &_pInstance->remote.candidates[_pInstance->remote.nCandidateCount];
			pj_bzero(pCandidate, sizeof(*pCandidate));

			/* type */
			if (strcmp(type, "host") == 0) pCandidate->type = PJ_ICE_CAND_TYPE_HOST;
			else if (strcmp(type, "srflx") == 0) pCandidate->type = PJ_ICE_CAND_TYPE_SRFLX;
			else if (strcmp(type, "relay") == 0) pCandidate->type = PJ_ICE_CAND_TYPE_RELAYED;
			else return -2;

			/* component id */
			pCandidate->comp_id = (pj_uint8_t)nCompId;

			/* foundation*/
			pj_strdup2(_pInstance->pPool, &pCandidate->foundation, foundation);

			/* priority */
			pCandidate->prio = nPrio;

			/* ip address */
			int af;
			if (strchr(ipaddr, ':')) {
				af = pj_AF_INET6();
			} else {
				af = pj_AF_INET();
			}
			pj_str_t addr = pj_str(ipaddr);
			pj_sockaddr_init(af, &pCandidate->addr, NULL, 0);
			pj_status_t status = pj_sockaddr_set_str_addr(af, &pCandidate->addr, &addr);
			if (status != PJ_SUCCESS) {
				return -3;
			}
			pj_sockaddr_set_port(&pCandidate->addr, (pj_uint16_t)nPort);
			_pInstance->remote.nCandidateCount++;
			if (pCandidate->comp_id > _pInstance->remote.nComponentCount) {
				_pInstance->remote.nComponentCount = pCandidate->comp_id;
			}
		}
	}

	if (_pInstance->remote.nCandidateCount == 0 || _pInstance->remote.ufrag[0] == 0 || _pInstance->remote.password[0] == 0 ||
	    _pInstance->remote.nComponentCount == 0) {
		return -5;
	}

	return 0;
}

int IceStartNegotiation(IN IceInstance* _pInstance)
{
	if (_pInstance == NULL) {
		PJ_LOG(1, ("IceStartNegotiation", "instance is nullptr"));
		return -1;
	}

	if (!pj_ice_strans_has_sess(_pInstance->pIceStreamTransport)) {
		PJ_LOG(1, ("IceStartNegotiation", "session not created"));
		return -2;
	}

	if (_pInstance->remote.nCandidateCount == 0) {
		PJ_LOG(1, ("IceStartNegotiation", "candidate count == 0"));
		return -3;
	}

	/* start negotiation */
	pj_str_t rufrag, rpwd;
	pj_status_t status = pj_ice_strans_start_ice(_pInstance->pIceStreamTransport,
						     pj_cstr(&rufrag, _pInstance->remote.ufrag),
						     pj_cstr(&rpwd, _pInstance->remote.password),
						     _pInstance->remote.nCandidateCount,
						     _pInstance->remote.candidates);
	if (status != PJ_SUCCESS) {
		PJ_LOG(1, ("IceStartNegotiation", "ice negotiation failed"));
		return -4;
	}

	return 0;
}

int IceSendToPeer(IN IceInstance* _pInstance, IN unsigned _nComponentId, IN const char* _pData, IN size_t _nSize)
{
	if (_pInstance == NULL) {
		return -1;
	}

	if (!pj_ice_strans_has_sess(_pInstance->pIceStreamTransport)) {
		return -2;
	}

	if (!pj_ice_strans_sess_is_complete(_pInstance->pIceStreamTransport)) {
		return -3;
	}

	if (_nComponentId < 1 || _nComponentId > pj_ice_strans_get_running_comp_cnt(_pInstance->pIceStreamTransport)) {
		return -4;
	}

	/* find the candidate with relayed address */
	pj_ice_sess_cand* pCandidate = NULL;
	for (int i = 0; i < _pInstance->remote.nCandidateCount; i++) {
		if (_pInstance->remote.candidates[i].type == PJ_ICE_CAND_TYPE_RELAYED) {
			pCandidate = &_pInstance->remote.candidates[i];
			break;
		}
	}
	if (pCandidate == NULL) {
		return -5;
	}

	pj_status_t status = pj_ice_strans_sendto(_pInstance->pIceStreamTransport, _nComponentId, _pData, _nSize,
						  &pCandidate->addr, pj_sockaddr_get_len(&pCandidate->addr));
	if (status !=  PJ_SUCCESS) {
		return -6;
	}

	return 0;
}

extern void IceCandidateString(IN const pj_ice_sess_cand* pCandidate, OUT char* pBuffer, IN size_t nSize)
{
	char ipaddr[PJ_INET6_ADDRSTRLEN];
	snprintf(pBuffer, nSize, "%.*s %u UDP %u %s %u typ %s\n",
		 (int)pCandidate->foundation.slen,
		 pCandidate->foundation.ptr,
		 (unsigned)pCandidate->comp_id,
		 pCandidate->prio,
		 pj_sockaddr_print(&pCandidate->addr, ipaddr, sizeof(ipaddr), 0),
		 (unsigned)pj_sockaddr_get_port(&pCandidate->addr),
		 pj_ice_get_cand_type_name(pCandidate->type));
}
