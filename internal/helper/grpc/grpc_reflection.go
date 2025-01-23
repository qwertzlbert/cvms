package grpchelper

import (
	"fmt"
	"strings"

	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func ResolveMessage(fullMethodName string, rcli *grpcreflect.Client) (protoreflect.MethodDescriptor, error) {
	// assume that fully-qualified method name cosists of
	// FULL_SERVER_NAME + "." + METHOD_NAME
	// so split the last dot to get service name
	n := strings.LastIndex(fullMethodName, ".")
	if n < 0 {
		return nil, fmt.Errorf("invalid method name: %v", fullMethodName)
	}
	serviceName := fullMethodName[0:n]
	methodName := fullMethodName[n+1:]

	sdesc, err := rcli.ResolveService(serviceName)
	if err != nil {
		return nil, fmt.Errorf("service couldn't be resolved: %v: %v", err, serviceName)
	}
	mdesc := sdesc.UnwrapService().Methods().ByName(protoreflect.Name(methodName))

	if mdesc == nil {
		return nil, fmt.Errorf("method couldn't be found")
	}

	return mdesc, nil
}

func CreateMessage(mdesc protoreflect.MethodDescriptor, inputJsonString string) (*dynamicpb.Message, error) {

	msg := dynamicpb.NewMessage(mdesc.Input())
	if err := protojson.Unmarshal([]byte(inputJsonString), msg); err != nil {
		return nil, fmt.Errorf("unmarshal %v", err)
	}
	return msg, nil
}
