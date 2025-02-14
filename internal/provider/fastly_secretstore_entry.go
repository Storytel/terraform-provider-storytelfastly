// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/fastly/go-fastly/v9/fastly"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SecretStoreEntriesResource{}
var _ resource.ResourceWithImportState = &SecretStoreEntriesResource{}

func NewSecretStoreEntriesResource() resource.Resource {
	return &SecretStoreEntriesResource{}
}

// SecretStoreEntriesResource defines the resource implementation.
type SecretStoreEntriesResource struct {
	client *fastly.Client
}

// SecretStoreEntriesModel describes the resource data model.
type SecretStoreEntriesModel struct {
	StoreID   types.String `tfsdk:"store_id"`
	Key       types.String `tfsdk:"key"`
	Value     types.String `tfsdk:"value"`
	Digest    types.String `tfsdk:"digest"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *SecretStoreEntriesResource) augmentStateFromSecret(ctx context.Context, secret *fastly.Secret, model *SecretStoreEntriesModel) {
	model.CreatedAt = types.StringValue(secret.CreatedAt.Format(time.RFC3339))
	model.Digest = types.StringValue(hex.EncodeToString(secret.Digest))
}

func (r *SecretStoreEntriesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secretstore_entry"
}

func (r *SecretStoreEntriesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Entries into a Fastly SecretStore",

		Attributes: map[string]schema.Attribute{
			"store_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the store to insert the entry.",
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The key of the secret",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				MarkdownDescription: "The value of the secret",
				Required:            true,
				Sensitive:           true,
			},
			"digest": schema.StringAttribute{
				MarkdownDescription: "Fastly digest of the secret. Used to detect drift.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Date and time  when this secret was created in ISO 8601.",
				Computed:            true,
			},
		},
	}
}

func (r *SecretStoreEntriesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*fastly.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *SecretStoreEntriesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SecretStoreEntriesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.CreateSecret(&fastly.CreateSecretInput{
		Method:  "POST",
		Name:    data.Key.ValueString(),
		Secret:  []byte(data.Value.ValueString()),
		StoreID: data.StoreID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to create secret with fastly", err.Error())
		return
	}

	r.augmentStateFromSecret(ctx, secret, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretStoreEntriesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecretStoreEntriesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.GetSecret(&fastly.GetSecretInput{
		Name:    data.Key.ValueString(),
		StoreID: data.StoreID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"could not get secret with name "+data.Key.ValueString()+" from store "+data.StoreID.ValueString(),
			err.Error(),
		)
		return
	}

	r.augmentStateFromSecret(ctx, secret, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretStoreEntriesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SecretStoreEntriesModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	secret, err := r.client.CreateSecret(&fastly.CreateSecretInput{
		Method:  "PATCH",
		Name:    data.Key.ValueString(),
		Secret:  []byte(data.Value.ValueString()),
		StoreID: data.StoreID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to update secret with key "+data.Key.ValueString(), err.Error())
		return
	}

	r.augmentStateFromSecret(ctx, secret, &data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SecretStoreEntriesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SecretStoreEntriesModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecret(&fastly.DeleteSecretInput{
		Name:    data.Key.ValueString(),
		StoreID: data.StoreID.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete secret with key "+data.Key.ValueString(), err.Error())
		return
	}
}

func (r *SecretStoreEntriesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ".")
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("store_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("key"), parts[1])...)
}
