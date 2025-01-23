package grpchelper

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	// Below are necessary imports for the cosmos SDK types
	// Otherwise we would have to manually import the relevant SDK types
	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
)

type CosmosAnyMessageResolver struct {
	protoregistry.MessageTypeResolver
	protoregistry.ExtensionTypeResolver
}

func (r CosmosAnyMessageResolver) FindMessageByURL(typeURL string) (protoreflect.MessageType, error) {
	// Only the part of typeUrl after the last slash is relevant.
	mname := typeURL
	if slash := strings.LastIndex(mname, "/"); slash >= 0 {
		mname = mname[slash+1:]
	}

	a, err := protoregistry.GlobalTypes.FindMessageByURL(mname)
	if err != nil {
		return nil, err
	}

	return a, nil
}
