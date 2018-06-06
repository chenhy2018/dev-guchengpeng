#include "qrtc.h"
int InitPeerConnectoin(OUT PeerConnection ** pPeerConnectoin,
                        IN IceConfig *pIceConfig)
{
  return 1;
}

int AddAudioTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pAudioConfig)
{
  return 1;
}

int AddVideoTrack(IN OUT PeerConnection * pPeerConnection, IN MediaConfig *pVideoConfig)
{
  return 1;
}

int createOffer(IN OUT PeerConnection * pPeerConnection, OUT pjmedia_sdp_session **pOffer)
{
  return 1;
}

int createAnswer(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session *pOffer,
                 OUT pjmedia_sdp_session **pAnswer)
{
  return 1;
}

int setLocalDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp)
{
  return 1;
}

int setRemoteDescription(IN OUT PeerConnection * pPeerConnection, IN pjmedia_sdp_session * pSdp)
{
  return 1;
}

int StartNegotiation(IN PeerConnection * pPeerConnection)
{
  return 1;
}

int SendPacket_1(IN PeerConnection *pPeerConnection, IN OUT RtpPacket * pPacket)
{
  return 1;
}

void ReleasePeerConnectoin(IN OUT PeerConnection * _pPeerConnection)
{
  return 1;
}

void InitIceConfig(IN OUT IceConfig *pIceConfig)
{

}
