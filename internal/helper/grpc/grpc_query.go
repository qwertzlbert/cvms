package grpchelper

import (
	"context"
	"strings"

	"log"

	"github.com/pkg/errors"

	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

func GrpcMakeDescriptor(reflectionClient *grpcreflect.Client, queryPath string) (protoreflect.MethodDescriptor, error) {
	methodDescriptor, err := ResolveMessage(queryPath, reflectionClient)
	if err != nil {
		return nil, errors.Wrapf(err, "by query path: %s", queryPath)
	}

	return methodDescriptor, nil
}

func GrpcInvokeQuery(
	ctx context.Context,
	methodDescriptor protoreflect.MethodDescriptor,
	conn *grpc.ClientConn,
	queryData string,
) (string, error) {

	msgRaw, err := CreateMessage(methodDescriptor, queryData)
	if err != nil {
		// client.Logger.Errorf("grpc api failed to create proto message: %v", err)
		return "", err
	}
	var msg []byte
	err = proto.Unmarshal(msg, msgRaw)

	if err != nil {
		// client.Logger.Errorf("grpc api failed to mashal jsonpb: %v", err)
		return "", err
	}
	response := dynamicpb.NewMessage(methodDescriptor.Output())
	var headerMD metadata.MD
	err = conn.Invoke(ctx, string(msgRaw.Descriptor().FullName()), msg, response, grpc.Header(&headerMD))
	if err != nil {
		// client.Logger.Errorf("grpc api failed to invoke rpc: %v", err)
		return "", err
	}

	var resolver CosmosAnyMessageResolver

	marshaller := protojson.MarshalOptions{
		AllowPartial:    true,
		Indent:          "  ",
		UseProtoNames:   false,
		EmitUnpopulated: false,
		Resolver:        resolver,
	}
	respJSON, err := marshaller.Marshal(response)
	if err != nil {
		// client.Logger.Errorf("grpc api failed to marshal string with grpc data: %v", err)
		return "", err
	}

	return string(respJSON), nil
}

func GrpcDynamicQuery(
	ctx context.Context,
	reflectionClient *grpcreflect.Client,
	conn *grpc.ClientConn,
	queryPath string,
	queryData string,
) (string, error) {
	methodDescriptor, err := ResolveMessage(queryPath, reflectionClient)
	if err != nil {
		log.Printf("grpc api failed to resolve proto message: %s", err.Error())
		return "", err
	}

	// Creates proto message body from query data formated in json
	msg, err := CreateMessage(methodDescriptor, queryData)
	if err != nil {
		log.Printf("grpc api failed to create proto message: %s", err.Error())
		return "", err
	}

	idx := strings.LastIndex(queryPath, ".")
	fullMethodName := queryPath[:idx] + "/" + queryPath[idx+1:]

	response := dynamicpb.NewMessage(methodDescriptor.Output())

	var headerMD metadata.MD
	err = conn.Invoke(ctx, fullMethodName, msg, response, grpc.Header(&headerMD))
	if err != nil {
		log.Printf("grpc api failed to invoke rpc: %s", err)
		return "", err
	}

	var resolver CosmosAnyMessageResolver

	marshaller := protojson.MarshalOptions{
		AllowPartial:    true,
		Indent:          "  ",
		UseProtoNames:   false,
		EmitUnpopulated: false,
		Resolver:        resolver,
	}
	respJSON, err := marshaller.Marshal(response)
	if err != nil {
		log.Printf("grpc api failed to marshal string with grpc data: %s", err.Error())
		return "", err
	}

	return string(respJSON), nil
}
