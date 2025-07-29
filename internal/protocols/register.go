package protocols

const RegisterType string = "register"

type Register struct {
	Name   string
	Codecs []string
}

func NewRegisterPacket(name string, codecs []string) *Packet {
	return &Packet{
		Type: RegisterType,
		Data: Register{
			Name:   name,
			Codecs: codecs,
		},
	}
}

func NewRegisterFromPacketData(data any) *Register {
	m, ok := data.(map[string]any)

	if !ok {
		return nil
	}

	var codecs []string
	raw_codecs := m["Codecs"].([]any)

	for i := range len(raw_codecs) {
		s, ok := raw_codecs[i].(string)
		if !ok {
			return nil
		}
		codecs = append(codecs, s)
	}

	return &Register{
		Name:   m["Name"].(string),
		Codecs: codecs,
	}
}
