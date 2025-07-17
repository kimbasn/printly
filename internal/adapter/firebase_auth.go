package adapter

import (
	"context"

	ierrors "github.com/kimbasn/printly/internal/errors"

	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/errorutils"
)

//go:generate mockgen -destination=../mocks/mock_firebase_auth_client.go -package=mocks github.com/kimbasn/printly/internal/adapter FirebaseAuthClient

// FirebaseAuthClient defines an interface for Firebase Auth operations, allowing for mocking.
type FirebaseAuthClient interface {
	CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error)
	GetUser(ctx context.Context, uid string) (*auth.UserRecord, error)
	DeleteUser(ctx context.Context, uid string) error
	//SetCustomerClaims(ctx context.Context, uid string, claims map[string]any) error
}

type firebaseAuthClient struct {
	client *auth.Client
}

func NewFirebaseAuth(client *auth.Client) *firebaseAuthClient {
	return &firebaseAuthClient{client: client}
}

func (a *firebaseAuthClient) CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error) {
	u, err := a.client.CreateUser(ctx, user)
	if err != nil {
		switch {
		case auth.IsEmailAlreadyExists(err):
			return nil, ierrors.ErrEmailAlreadyExists
		case errorutils.IsInvalidArgument(err):
			return nil, ierrors.NewWithCause(ierrors.InvalidArgument, "invalid user data", err)
		case errorutils.IsUnauthenticated(err):
			return nil, ierrors.NewWithCause(ierrors.Unauthenticated, "firebase authentication failed", err)
		default:
			return nil, ierrors.NewWithCause(ierrors.Internal, "firebase failed to create user", err)
		}
	}
	return u, nil
}

func (a *firebaseAuthClient) GetUser(ctx context.Context, uid string) (*auth.UserRecord, error) {
	u, err := a.client.GetUser(ctx, uid)
	if err != nil {
		switch {
		case auth.IsUserNotFound(err):
			return nil, ierrors.ErrUserNotFound
		case auth.IsUserDisabled(err):
			return nil, ierrors.ErrUserDisabled

		default:
			return nil, ierrors.NewWithCause(ierrors.Internal, "failed to get user from firebase", err)
		}
	}
	return u, nil
}

func (a *firebaseAuthClient) DeleteUser(ctx context.Context, uid string) error {
	err := a.client.DeleteUser(ctx, uid)
	if err != nil {
		switch {
		case auth.IsUserNotFound(err):
			return ierrors.ErrUserNotFound
		case errorutils.IsInvalidArgument(err):
			return ierrors.NewWithCause(ierrors.InvalidArgument, "invalid user UID", err)
		case errorutils.IsUnauthenticated(err):
			return ierrors.NewWithCause(ierrors.Unauthenticated, "firebase authentication failed", err)
		case errorutils.IsPermissionDenied(err):
			return ierrors.NewWithCause(ierrors.PermissionDenied, "permission denied", err)
		default:
			return ierrors.NewWithCause(ierrors.Internal, "firebase: failed to delete user", err)
		}
	}
	return nil
}
