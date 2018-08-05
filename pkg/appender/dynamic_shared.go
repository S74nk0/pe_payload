package appender

// TODO maybe make this a variable
// 0.5 MB should be more than enough
const maxDynamicSize = 1000000 / 2

func calcPayloadMsgSize(payloadMsgSize, payloadMessageStep uint32) uint32 {
	mult := (payloadMsgSize / payloadMessageStep)
	if (payloadMsgSize % payloadMessageStep) != 0 {
		mult++
	}
	ret := mult * payloadMessageStep

	// fmt.Printf("## (%d / %d) + %d = %d ##\n", payloadMsgSize, payloadMessageStep, payloadMessageStep, ret)
	return ret
}
