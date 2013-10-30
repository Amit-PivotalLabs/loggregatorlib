package testhelpers

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
	"testing"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/cloudfoundry/loggregatorlib/signature"
)

func MarshalledErrorLogMessage(t *testing.T, messageString string, appId string) []byte {
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledLogMessage(t *testing.T, messageString string, appId string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledDrainedLogMessage(t *testing.T, messageString string, appId string, drainUrls ...string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_WARDEN_CONTAINER
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)
	protoMessage.DrainUrls = drainUrls

	return marshalProtoBuf(t, protoMessage)
}

func MarshalledDrainedNonWardenLogMessage(t *testing.T, messageString string, appId string, drainUrls ...string) []byte {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_DEA
	protoMessage := generateLogMessage(messageString, appId, messageType, sourceType)
	protoMessage.DrainUrls = drainUrls

	return marshalProtoBuf(t, protoMessage)
}


func NewLogMessage(messageString, appId string) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_OUT
	sourceType := logmessage.LogMessage_WARDEN_CONTAINER

	return generateLogMessage(messageString, appId, messageType, sourceType)
}

func MarshalledLogEnvelope(t *testing.T, unmarshalledMessage *logmessage.LogMessage, secret string) []byte {
	signatureOfMessage, err := signature.Encrypt(secret, signature.Digest(unmarshalledMessage.String()))
	assert.NoError(t, err)

	envelope := &logmessage.LogEnvelope{
		LogMessage: unmarshalledMessage,
		RoutingKey: proto.String(*unmarshalledMessage.AppId),
		Signature:  signatureOfMessage,
	}

	return marshalProtoBuf(t, envelope)
}

func AssertProtoBufferMessageEquals(t *testing.T, expectedMessage string, actual []byte) {
	receivedMessage := &logmessage.LogMessage{}
	err := proto.Unmarshal(actual, receivedMessage)
	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, string(receivedMessage.GetMessage()))
}

func generateLogMessage(messsageString, appId string, messageType logmessage.LogMessage_MessageType, sourceType logmessage.LogMessage_SourceType) *logmessage.LogMessage {
	currentTime := time.Now()
	logMessage := &logmessage.LogMessage{
		Message:     []byte(messsageString),
		AppId:       proto.String(appId),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	return logMessage
}

func marshalProtoBuf(t *testing.T, pb proto.Message) []byte {
	marshalledProtoBuf, err := proto.Marshal(pb)
	assert.NoError(t, err)

	return marshalledProtoBuf
}