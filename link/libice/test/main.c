#include <ice.h>

void AonRx(IN pj_ice_strans* pStreamTransport, IN unsigned nComponentId, IN void* pData,
	      IN pj_size_t size, IN const pj_sockaddr_t* pAddr, IN unsigned nAddrLen)
{
	char data[500];
	memset(data, 0, 500);
	memcpy(data, pData, (size > 499 ? 499 : size));

	PJ_LOG(3,("info", "A received: %s", data));
}

void BonRx(IN pj_ice_strans* pStreamTransport, IN unsigned nComponentId, IN void* pData,
	      IN pj_size_t size, IN const pj_sockaddr_t* pAddr, IN unsigned nAddrLen)
{
	char data[500];
	memset(data, 0, 500);
	memcpy(data, pData, (size > 499 ? 499 : size));

	PJ_LOG(3,("info", "B received: %s", data));
}

void AonIce(IN pj_ice_strans *pStreamTransport, IN pj_ice_strans_op op, IN pj_status_t status)
{
	PJ_LOG(3,("info", "A ICE => %d, status=%d", op, status));
}

void BonIce(IN pj_ice_strans *pStreamTransport, IN pj_ice_strans_op op, IN pj_status_t status)
{
	PJ_LOG(3,("info", "B ICE => %d, status=%d", op, status));
}

int main(int argc, char** argv)
{
	// turn server information
	IceOptions opt;
	memset(&opt, 0, sizeof(opt));
	opt.nComponents = 1;
	opt.nMaxHosts = 1;
	opt.bRegular = 1;
	opt.bTurnTcp = 1;
	char* turnhost = "123.59.204.198:3478";
	memcpy(&opt.turnHost, turnhost, strlen(turnhost));
	char* turnuser = "root";
	memcpy(&opt.turnUsername, turnuser, strlen(turnuser));
	char* turnpsw = "root";
	memcpy(&opt.turnPassword, turnpsw, strlen(turnpsw));
	opt.nKeepAlive = 300;

	IceOptions opta, optb;
	memcpy(&opta, &opt, sizeof(opt));
	memcpy(&optb, &opt, sizeof(opt));

	opta.onRxData = AonRx;
	optb.onRxData = BonRx;
	opta.onIceComplete = AonIce;
	optb.onIceComplete = BonIce;

	// create instance
	IceInstance* pa = IceCreateInstance(&opta);
	IceInstance* pb = IceCreateInstance(&optb);

	// wait for turn to return relayed ip
	sleep(10);

	// set offerer and anwerer
	IceSetOfferer(pa);
	IceSetAnswerer(pb);

	// generate sdp
	pjmedia_sdp_media sdpa, sdpb;
	memset(&sdpa, 0, sizeof(sdpa));
	memset(&sdpb, 0, sizeof(sdpb));

        pj_caching_pool cp;
        pj_pool_t* pool;
        pj_caching_pool_init(&cp, NULL, 0);
        pool = pj_pool_create(&cp.factory, NULL, 512, 512, NULL);

	// ice usr/pwd/cand for a
	pj_str_t ufraga = pj_str(pa->local.ufrag);
	pjmedia_sdp_attr* attr1 = pjmedia_sdp_attr_create(pool, "ice-ufrag", &ufraga);
	pj_str_t pwda = pj_str(pa->local.password);
	pjmedia_sdp_attr* attr2 = pjmedia_sdp_attr_create(pool, "ice-pwd", &pwda);
	char ra[100];
	for (int i; i < pa->local.nCandidateCount; i++) {
		if (pa->local.candidates[i].type == PJ_ICE_CAND_TYPE_RELAYED) {
			IceCandidateString(&pa->local.candidates[i], ra, 100);
			break;
		}
	}
	pj_str_t relaya = pj_str(ra);
	pjmedia_sdp_attr* attr3 = pjmedia_sdp_attr_create(pool, "candidate", &relaya);

	pjmedia_sdp_media_add_attr(&sdpa, attr1);
	pjmedia_sdp_media_add_attr(&sdpa, attr2);
	pjmedia_sdp_media_add_attr(&sdpa, attr3);

	// ice usr/pwd/cand for b
	pj_str_t ufragb = pj_str(pb->local.ufrag);
	pjmedia_sdp_attr* attr4 = pjmedia_sdp_attr_create(pool, "ice-ufrag", &ufragb);
	pj_str_t pwdb = pj_str(pb->local.password);
	pjmedia_sdp_attr* attr5 = pjmedia_sdp_attr_create(pool, "ice-pwd", &pwdb);
	char rb[100];
	for (int i; i < pb->local.nCandidateCount; i++) {
		if (pb->local.candidates[i].type == PJ_ICE_CAND_TYPE_RELAYED) {
			IceCandidateString(&pb->local.candidates[i], rb, 100);
			break;
		}
	}
	pj_str_t relayb = pj_str(rb);
	pjmedia_sdp_attr* attr6 = pjmedia_sdp_attr_create(pool, "candidate", &relayb);

	pjmedia_sdp_media_add_attr(&sdpb, attr4);
	pjmedia_sdp_media_add_attr(&sdpb, attr5);
	pjmedia_sdp_media_add_attr(&sdpb, attr6);

	// exchange remote sdp
	IceGetRemote(pa, &sdpb);
	IceGetRemote(pb, &sdpa);

	// start negotiation
	IceStartNegotiation(pa);
	IceStartNegotiation(pb);

	// send some data to others
	int nCount = 0;
	while (1) {
		char stra[1500];
		char strb[1500];
		memset(stra, ' ', 1400);
		memset(strb, ' ', 1400);
		stra[20] = '\0';
		strb[20] = '\0';

		char count[10];
		snprintf(count, 8, "%07d ", nCount);
		memcpy(stra, count, 7);
		memcpy(strb, count, 7);
		nCount++;

		int i = IceSendToPeer(pa, 1, stra, strlen(stra));
//		int j = IceSendToPeer(pb, 1, strb, strlen(strb));
		int s = (int)pj_ice_strans_get_state(pa->pIceStreamTransport);

		usleep(1000 * 20);
	}
}
