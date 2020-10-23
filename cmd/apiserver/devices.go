//--------------------------------------------------------------------------
// Copyright 2018 Infinite Devices GmbH
// www.infinimesh.io
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//--------------------------------------------------------------------------

package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/infinimesh/infinimesh/pkg/apiserver/apipb"
	"github.com/infinimesh/infinimesh/pkg/node/nodepb"
	"github.com/infinimesh/infinimesh/pkg/registry/registrypb"
)

type deviceAPI struct {
	client        registrypb.DevicesClient
	accountClient nodepb.AccountServiceClient
}

func (d *deviceAPI) Create(ctx context.Context, request *registrypb.CreateRequest) (response *registrypb.CreateResponse, err error) {

	//Added logging
	log.Info("Create Device API Method: Function Invoked", zap.String("Requestor ID", ctx.Value("account_id").(string)))

	//Added the requestor account id to context metadata so that it can be passed on to the server
	ctx = metadata.AppendToOutgoingContext(ctx, "requestorid", ctx.Value("account_id").(string))

	//Invoke the Update Account controller for server
	res, err := d.client.Create(ctx, request)
	if err != nil {
		//Added logging
		log.Error("Create Device API Method: Failed to update Account", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Added logging
	log.Info("Create Device API Method: Device succesfully created")
	return res, nil
}

func (d *deviceAPI) Update(ctx context.Context, request *registrypb.UpdateRequest) (response *registrypb.UpdateResponse, err error) {

	//Added logging
	log.Info("Update Device API Method: Function Invoked", zap.String("Requestor ID", ctx.Value("account_id").(string)))

	//Added the requestor account id to context metadata so that it can be passed on to the server
	ctx = metadata.AppendToOutgoingContext(ctx, "requestorid", ctx.Value("account_id").(string))

	//Invoke the Update Account controller for server
	res, err := d.client.Update(ctx, request)
	if err != nil {
		//Added logging
		log.Error("Update Device API Method: Failed to update Account", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Added logging
	log.Info("Update Device API Method: Device succesfully updated")
	return res, nil
}

func (d *deviceAPI) Get(ctx context.Context, request *registrypb.GetRequest) (response *registrypb.GetResponse, err error) {

	//Added logging
	log.Info("Get Device API Method: Function Invoked", zap.String("Requestor ID", ctx.Value("account_id").(string)))

	account, ok := ctx.Value("account_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "The account is not authenticated.")
	}

	//Check if the user has access to get the device details
	resp, err := d.accountClient.IsAuthorized(ctx, &nodepb.IsAuthorizedRequest{
		Node:    request.Id,
		Account: account,
		Action:  nodepb.Action_READ,
	})
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "Could not get permission to list devices.")
	}

	fmt.Println("decision", resp.Decision.Value)
	if !resp.GetDecision().GetValue() {
		return nil, status.Error(codes.PermissionDenied, "The account does not have permission to get device list.")
	}

	//Get the device if the user has access
	return d.client.Get(ctx, request)

}
func (d *deviceAPI) List(ctx context.Context, request *apipb.ListDevicesRequest) (response *registrypb.ListResponse, err error) {

	//Added logging
	log.Info("List Devices API Method: Function Invoked", zap.String("Requestor ID", ctx.Value("account_id").(string)))

	account, ok := ctx.Value("account_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "The account is not authenticated.")
	}

	isRootResp, err := d.accountClient.IsRoot(ctx, &nodepb.IsRootRequest{
		Account: account,
	})
	if err != nil {
		return nil, err
	}

	//If Root provide all access
	if isRootResp.IsRoot {
		return d.client.List(ctx, &registrypb.ListDevicesRequest{Namespace: request.Namespace})
	}

	//Check if the user has access to the namespace
	resp, err := d.accountClient.IsAuthorizedNamespace(ctx, &nodepb.IsAuthorizedNamespaceRequest{
		Namespaceid: request.Namespace,
		Account:     account,
		Action:      nodepb.Action_READ,
	})
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "Could not get permission to list device.")
	}
	if !resp.GetDecision().GetValue() {
		return nil, status.Error(codes.PermissionDenied, "The account does not have permission to list device.")
	}

	//List the devices if the user has access
	list, err := d.client.ListForAccount(ctx, &registrypb.ListDevicesRequest{Namespace: request.Namespace, Account: account})
	return list, err
}
func (d *deviceAPI) Delete(ctx context.Context, request *registrypb.DeleteRequest) (response *registrypb.DeleteResponse, err error) {

	//Added logging
	log.Info("Delete Device API Method: Function Invoked", zap.String("Requestor ID", ctx.Value("account_id").(string)))

	account, ok := ctx.Value("account_id").(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "The account is not authenticated.")
	}

	//Check if the user has access to delete the device
	resp, err := d.accountClient.IsAuthorized(ctx, &nodepb.IsAuthorizedRequest{
		Node:    request.Id,
		Account: account,
		Action:  nodepb.Action_WRITE,
	})
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "Could not get permission to list devices.")
	}

	fmt.Println("decision", resp.Decision.Value)
	if !resp.GetDecision().GetValue() {
		return nil, status.Error(codes.PermissionDenied, "The account does not have permission to delete the device.")
	}

	//Delete the device if the user has access
	return d.client.Delete(ctx, request)
}
