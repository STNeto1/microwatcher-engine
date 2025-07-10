package graph

// THIS CODE WILL BE UPDATED WITH SCHEMA CHANGES. PREVIOUS IMPLEMENTATION FOR SCHEMA CHANGES WILL BE KEPT IN THE COMMENT SECTION. IMPLEMENTATION FOR UNCHANGED SCHEMA WILL BE KEPT.

import (
	"context"

	"github.com/google/uuid"
	"github.com/microwatcher/shared/pkg/clickhouse"
	"github.com/microwatcher/webserver/internal/graph/model"
)

type Resolver struct {
	ChSource *clickhouse.ClickhouseSource
}

// Devices is the resolver for the devices field.
func (r *queryResolver) Devices(ctx context.Context) ([]*model.Device, error) {
	panic("not implemented")
}

// CreateDevice is the resolver for the createDevice field.
func (r *queryResolver) CreateDevice(ctx context.Context, input model.CreateDevice) (*model.Device, error) {
	panic("not implemented")
}

// ResetDeviceSecret is the resolver for the resetDeviceSecret field.
func (r *queryResolver) ResetDeviceSecret(ctx context.Context, deviceID uuid.UUID) (*model.Device, error) {
	panic("not implemented")
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
/*
	type Resolver struct{}
*/
