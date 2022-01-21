// +build test

package user

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
	"ohmnyom/domain/user"
	fs "ohmnyom/internal/firestore"
	"ohmnyom/internal/path"
	"ohmnyom/internal/time"
)

func TestService_Delete(t *testing.T) {
	ctx := context.TODO()
	cli := newTestClient(ctx)
	setupTestData(ctx, cli)
	defer teardownTestData(ctx, cli)
	type fields struct {
		client *firestore.Client
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{"empty param", fields{cli}, args{ctx, ""}, assert.Error},
		{"not found", fields{cli}, args{ctx, "id-notfound"}, assert.NoError},
		{"user1", fields{cli}, args{ctx, users[0].Id}, assert.NoError},
		{"user2", fields{cli}, args{ctx, users[1].Id}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				client: tt.fields.client,
			}
			tt.wantErr(t, s.Delete(tt.args.ctx, tt.args.id), fmt.Sprintf("Delete(%v, %v)", tt.args.ctx, tt.args.id))
		})
	}
}

func TestService_Get(t *testing.T) {
	ctx := context.TODO()
	cli := newTestClient(ctx)
	setupTestData(ctx, cli)
	defer teardownTestData(ctx, cli)

	type fields struct {
		client *firestore.Client
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr assert.ErrorAssertionFunc
	}{
		{"empty param", fields{cli}, args{ctx, ""}, nil, assert.Error},
		{"not found", fields{cli}, args{ctx, "id-notfound"}, nil, assert.Error},
		{"user1", fields{cli}, args{ctx, users[0].Id}, users[0], assert.NoError},
		{"user2", fields{cli}, args{ctx, users[1].Id}, users[1], assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				client: tt.fields.client,
			}
			got, err := s.Get(tt.args.ctx, tt.args.id)
			if !tt.wantErr(t, err, fmt.Sprintf("Get(%v, %v)", tt.args.ctx, tt.args.id)) {
				return
			}
			assert.Equalf(t, tt.want, got, "Get(%v, %v)", tt.args.ctx, tt.args.id)
		})
	}
}

func TestService_GetByEmail(t *testing.T) {
	ctx := context.TODO()
	cli := newTestClient(ctx)
	setupTestData(ctx, cli)
	defer teardownTestData(ctx, cli)

	type fields struct {
		client *firestore.Client
	}
	type args struct {
		ctx   context.Context
		email string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr assert.ErrorAssertionFunc
	}{
		{"empty param", fields{cli}, args{ctx, ""}, nil, assert.Error},
		{"not found", fields{cli}, args{ctx, "email-notfound@test.com"}, nil, assert.Error},
		{"user1", fields{cli}, args{ctx, users[0].Email}, users[0], assert.NoError},
		{"user2", fields{cli}, args{ctx, users[1].Email}, users[1], assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				client: tt.fields.client,
			}
			got, err := s.GetByEmail(tt.args.ctx, tt.args.email)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByEmail(%v, %v)", tt.args.ctx, tt.args.email)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByEmail(%v, %v)", tt.args.ctx, tt.args.email)
		})
	}
}

func TestService_GetByOAuth(t *testing.T) {
	ctx := context.TODO()
	cli := newTestClient(ctx)
	setupTestData(ctx, cli)
	defer teardownTestData(ctx, cli)

	type fields struct {
		client *firestore.Client
	}
	type args struct {
		ctx       context.Context
		oauthType user.OAuthType
		oauthId   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *user.User
		wantErr assert.ErrorAssertionFunc
	}{
		// {"empty param", fields{cli}, args{ctx, users[1].OAuthType, ""}, nil, assert.Error},
		// {"none type", fields{cli}, args{ctx, user.OAuthType_NONE, users[1].OAuthId}, nil, assert.Error},
		// {"not found", fields{cli}, args{ctx, user.OAuthTYpe_KAKAO, users[1].OAuthId}, nil, assert.Error},
		// {"user1", fields{cli}, args{ctx, users[1].OAuthType, users[1].OAuthId}, users[1], assert.NoError},
		// {"user2", fields{cli}, args{ctx, users[2].OAuthType, users[2].OAuthId}, users[2], assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				client: tt.fields.client,
			}
			got, err := s.GetByOAuth(tt.args.ctx, tt.args.oauthType, tt.args.oauthId)
			if !tt.wantErr(t, err, fmt.Sprintf("GetByOAuth(%v, %v, %v)", tt.args.ctx, tt.args.oauthType, tt.args.oauthId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetByOAuth(%v, %v, %v)", tt.args.ctx, tt.args.oauthType, tt.args.oauthId)
		})
	}
}

func TestService_Put(t *testing.T) {
	ctx := context.TODO()
	cli, err := fs.NewClient(ctx, "ohmnyom", filepath.Join(path.Root(), "assets", "ohmnyom-77df675cb827.json"))
	if err != nil {
		t.Fatal(err)
	}
	// cli := newTestClient(ctx)
	// setupTestData(ctx, cli)
	// defer teardownTestData(ctx, cli)

	type fields struct {
		client *firestore.Client
	}
	type args struct {
		ctx  context.Context
		user *user.User
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{"empty param", fields{cli}, args{ctx, nil}, assert.Error},
		{"duplicate(update)", fields{cli}, args{ctx, users[0]}, assert.NoError},
		{"new", fields{cli}, args{ctx, &user.User{
			Id:       "id-temp",
			Name:     "name-temp",
			Email:    "email-temp@test.com",
			Password: "password-temp",
			OAuthInfo: []*user.OAuthInfo{
				{OAuthType: user.OAuthType_GOOGLE, OAuthId: "googleid-temp"},
			},
			Photourl: "photourl-temp",
			SignedUp: time.Time{},
			Pets:     nil,
		}}, assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				client: tt.fields.client,
			}
			tt.wantErr(t, s.Put(tt.args.ctx, tt.args.user), fmt.Sprintf("Put(%v, %v)", tt.args.ctx, tt.args.user))
		})
	}
}
