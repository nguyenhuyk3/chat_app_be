package websocketv2

// import (
// 	"be_chat_app/models"
// 	"log"

// 	"github.com/pion/webrtc/v3"
// )

// func NewPeerConnection() (*webrtc.PeerConnection, error) {
// 	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
// 		ICEServers: []webrtc.ICEServer{
// 			{
// 				URLs: []string{
// 					"stun:stun.l.google.com:19302",
// 					"stun:stun1.l.google.com:19302",
// 				},
// 			},
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, typ := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
// 		if _, err := peerConnection.AddTransceiverFromKind(typ, webrtc.RTPTransceiverInit{
// 			Direction: webrtc.RTPTransceiverDirectionSendrecv,
// 		}); err != nil {
// 			return nil, err
// 		}
// 	}
// 	return peerConnection, nil
// }

// func (h *Hub) HandleOffer(commingMessage models.CommingMessage) {
// 	peerConnection, err := NewPeerConnection()
// 	if err != nil {
// 		log.Println("Error creating peer connection (HandleOffer):", err)
// 		return
// 	}

// 	isOnMasterRoom := false
// 	if masterRoom, ok := h.MasterRooms[commingMessage.ReceiverId]; !ok {
// 		masterRoom.ClientOnMasterRoom.PeerConnection = peerConnection
// 		isOnMasterRoom = true
// 	}

// 	if !isOnMasterRoom {
// 		if client, ok := h.MessageBoxes[commingMessage.MessageBoxId].Clients[commingMessage.ReceiverId]; !ok {
// 			client.PeerConnection = peerConnection
// 		}
// 	}

// 	peerConnection.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 		if track.Kind() == webrtc.RTPCodecTypeVideo {
// 			log.Printf("Received remote video track (HandleOffer) (%s %s): %s\n", commingMessage.SenderId, commingMessage.ReceiverId, track.ID())
// 		} else if track.Kind() == webrtc.RTPCodecTypeAudio {
// 			log.Printf("Received remote audio track (HandleOffer) (%s %s): %s\n", commingMessage.SenderId, commingMessage.ReceiverId, track.ID())
// 		}
// 	})

// 	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
// 		if candidate != nil {
// 			candidateInit := candidate.ToJSON()
// 			if !isOnMasterRoom {
// 				candidateData := &models.Message{
// 					SenderId:  commingMessage.ReceiverId,
// 					Type:      "ice-candidate",
// 					Candidate: &candidateInit,
// 				}

// 				if client, ok := h.MessageBoxes[commingMessage.MessageBoxId].Clients[commingMessage.SenderId]; ok {
// 					client.Message <- candidateData
// 				}
// 			}
// 		}
// 	})

// 	err = peerConnection.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: webrtc.SDPTypeOffer,
// 		SDP:  commingMessage.Sdp.SDP,
// 	})
// 	if err != nil {
// 		log.Printf("Error setting remote description (HandleOffer) (%s %s): %v\n", commingMessage.SenderId, commingMessage.ReceiverId, err)
// 		return
// 	}

// 	answer, err := peerConnection.CreateAnswer(nil)
// 	if err != nil {
// 		log.Printf("Error creating answer (HandleOffer) (%s %s): %v\n", commingMessage.SenderId, commingMessage.ReceiverId, err)
// 		return
// 	}

// 	err = peerConnection.SetLocalDescription(answer)
// 	if err != nil {
// 		log.Printf("Error setting local description (HandleOffer) (%s %s): %v\n", commingMessage.SenderId, commingMessage.ReceiverId, err)
// 		return
// 	}
// 	if !isOnMasterRoom {
// 		answerData := &models.Message{
// 			SenderId: commingMessage.SenderId,
// 			Type:     "answer",
// 			Sdp:      &answer,
// 		}
// 		if client, ok := h.MessageBoxes[commingMessage.MessageBoxId].Clients[commingMessage.SenderId]; ok {
// 			client.Message <- answerData
// 		}
// 	} else {

// 	}
// }

// func (h *Hub) HandleAnswer(commingMessage models.CommingMessage) {
// 	var peerConnection *webrtc.PeerConnection

// 	// * This case is client on MessageBox
// 	if client, ok := h.MessageBoxes[commingMessage.MessageBoxId].Clients[commingMessage.ReceiverId]; ok {
// 		peerConnection = client.PeerConnection
// 	}
// 	// * This case is client on MasterRoom
// 	if masterRoom, ok := h.MasterRooms[commingMessage.ReceiverId]; ok {
// 		peerConnection = masterRoom.ClientOnMasterRoom.PeerConnection
// 	}
// 	if peerConnection == nil {
// 		log.Printf("PeerConnection not found for ReceiverId: %s\n", commingMessage.ReceiverId)
// 		return
// 	}
// 	if commingMessage.Sdp == nil {
// 		log.Printf("No SDP provided in answer message (HandleAnswer) (%s %s)", commingMessage.SenderId, commingMessage.ReceiverId)
// 		return
// 	}
// 	err := peerConnection.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: webrtc.SDPTypeAnswer,
// 		SDP:  commingMessage.Sdp.SDP,
// 	})
// 	if err != nil {
// 		log.Println("Error setting remote description (HandleAnswer):", err)
// 		return
// 	}
// }

// func (h *Hub) HandleNewIceCandidate(commingMessage models.CommingMessage) {
// 	var peerConnection *webrtc.PeerConnection
// 	var isOnMasterRoom bool = false
// 	// * This case is client on MasterRoom
// 	if masterRoom, ok := h.MasterRooms[commingMessage.ReceiverId]; ok {
// 		peerConnection = masterRoom.ClientOnMasterRoom.PeerConnection
// 		isOnMasterRoom = true
// 	}
// 	// * This case is client on MessageBox
// 	if client, ok := h.MessageBoxes[commingMessage.MessageBoxId].Clients[commingMessage.ReceiverId]; ok && !isOnMasterRoom {
// 		peerConnection = client.PeerConnection
// 	}
// 	// * This case will be enabled when two case above occur
// 	if peerConnection == nil {
// 		log.Printf("PeerConnection not found for ReceiverId: %s\n", commingMessage.ReceiverId)
// 		return
// 	}
// 	go func() {
// 		if err := peerConnection.AddICECandidate(*commingMessage.Candidate); err != nil {
// 			log.Printf("Error adding ICE candidate (HandleNewIceCandidate): %v\n", err)
// 		}
// 	}()
// }
