package protocols

import "log"

const EncodeRequestType string = "encode_request"

type EncodeRequest struct {
	Input      string
	Output     string
	CRF        string
	Preset     string
	AudioCodec string
	VideoCodec string
}

func NewEncodeRequestPacket(input string, output string, crf string, preset string, audioCodec string, videoCodec string) *Packet {
	return &Packet{
		Type: EncodeRequestType,
		Data: EncodeRequest{
			Input:      input,
			Output:     output,
			CRF:        crf,
			Preset:     preset,
			AudioCodec: audioCodec,
			VideoCodec: videoCodec,
		},
	}
}

func NewEncodeRequestFromPacketData(data any) *EncodeRequest {
	m, ok := data.(map[string]any)
	if !ok {
		log.Println("Failed to create EncodeRequest from data:", data)
		return nil
	}

	return &EncodeRequest{
		Input:      m["Input"].(string),
		Output:     m["Output"].(string),
		CRF:        m["CRF"].(string),
		Preset:     m["Preset"].(string),
		AudioCodec: m["AudioCodec"].(string),
		VideoCodec: m["VideoCodec"].(string),
	}
}
