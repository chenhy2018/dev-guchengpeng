#include <ice.h>

int main(int argc, char** argv)
{
	IceOptions opt;
	memset(&opt, 0, sizeof(opt));
	opt.nComponents = 1;
	opt.nMaxHosts = 1;
	opt.bRegular = 1;
	char* turnhost = "123.59.204.198:3478";
	memcpy(&opt.turnHost, turnhost, strlen(turnhost));
	opt.nKeepAlive = 300;

	IceInstance* pa = IceCreateInstance(&opt);
	IceInstance* pb = IceCreateInstance(&opt);

	sleep(10);

	IceSetOfferer(pa);
	IceSetAnswerer(pb);

	pjmedia_sdp_media sdpa, sdpb;
	memset(&sdpa, 0, sizeof(sdpa));
	memset(&sdpb, 0, sizeof(sdpb));

        pj_caching_pool cp;
        pj_pool_t* pool;
        pj_caching_pool_init(&cp, NULL, 0);
        pool = pj_pool_create(&cp.factory, NULL, 512, 512, NULL);

	// for a
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

	// for b
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

	IceGetRemote(pa, &sdpb);
	IceGetRemote(pb, &sdpa);

	IceStartNegotiation(pa);
	sleep(5);
	IceStartNegotiation(pb);

	sleep(1000);
}
