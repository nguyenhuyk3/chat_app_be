package websocketv2

import (
	"log"

	"github.com/pion/webrtc/v3"
)

func (h *Hub) SetupPeerConnection() (*webrtc.PeerConnection, error) {
	// Create configuration for Peer connection
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				// STUN server
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	// Create Peer connection
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func (h *Hub) addMediaTracks(pc *webrtc.PeerConnection) error {
	// Get video from webcam
	videoTrack, err := webrtc.NewTrackLocalStaticSample(
		webrtc.RTPCodecCapability{MimeType: "video/VP8"},
		"video",
		"pion",
	)
	if err != nil {
		return err
	}
	// Add track video into Peer connection
	_, err = pc.AddTrack(videoTrack)
	if err != nil {
		return err
	}
	return nil
}

func (h *Hub) handleOffer(sdp string, clientOnMasterRoom *ClientOnMasterRoom) {
	peerConnection, err := h.SetupPeerConnection()
	if err != nil {
		log.Println("Error setting up peer connection:", err)
		return
	}
	// Thiết lập OnICECandidate để gửi candidate tới peer khác
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c != nil {
			clientOnMasterRoom.SendICECandidate(c.ToJSON())
		}
	})
	// Thiết lập OnTrack để nhận track media từ peer khác
	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		// Xử lý track media ở đây
	})
	// Create and set Remote description
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}
	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Println("Error setting remote description:", err)
		return
	}
	// Create answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Error creating answer:", err)
		return
	}
	// Set Local Description
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Error setting local description:", err)
		return
	}
	// Gửi SDP answer tới peer khác
	clientOnMasterRoom.SendSDP(answer.SDP)
}
