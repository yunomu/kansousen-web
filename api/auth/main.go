package main

import (
	"context"
	"os"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"

	"github.com/yunomu/kansousen/lib/lambda"
	apipb "github.com/yunomu/kansousen/proto/api"
)

type server struct {
	clientId string
	c        *cognitoidentityprovider.CognitoIdentityProvider
}

func (s *server) signUp(ctx context.Context, r *apipb.SignUpRequest) (*apipb.AuthResponse, error) {
	out, err := s.c.SignUpWithContext(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId: aws.String(s.clientId),
		Username: aws.String(r.Username),
		Password: aws.String(r.Password),
		UserAttributes: []*cognitoidentityprovider.AttributeType{
			{
				Name:  aws.String("email"),
				Value: aws.String(r.Email),
			},
		},
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.InvalidPasswordException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}

	return &apipb.AuthResponse{
		AuthResponseSelect: &apipb.AuthResponse_ResponseSignUp{
			ResponseSignUp: &apipb.SignUpResponse{
				CodeDeliveryType:        aws.StringValue(out.CodeDeliveryDetails.DeliveryMedium),
				CodeDeliveryDestination: aws.StringValue(out.CodeDeliveryDetails.Destination),
			},
		},
	}, nil
}

func (s *server) confirmSignUp(ctx context.Context, r *apipb.ConfirmSignUpRequest) (*apipb.AuthResponse, error) {
	_, err := s.c.ConfirmSignUpWithContext(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(s.clientId),
		Username:         aws.String(r.Username),
		ConfirmationCode: aws.String(r.ConfirmationCode),
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.CodeMismatchException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.ExpiredCodeException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.UserNotFoundException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}

	return &apipb.AuthResponse{}, nil
}

func (s *server) resendConfirmationCode(ctx context.Context, r *apipb.ResendConfirmationCodeRequest) (*apipb.AuthResponse, error) {
	out, err := s.c.ResendConfirmationCodeWithContext(ctx, &cognitoidentityprovider.ResendConfirmationCodeInput{
		ClientId: aws.String(s.clientId),
		Username: aws.String(r.Username),
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}
	return &apipb.AuthResponse{
		AuthResponseSelect: &apipb.AuthResponse_ResponseSignUp{
			ResponseSignUp: &apipb.SignUpResponse{
				CodeDeliveryType:        aws.StringValue(out.CodeDeliveryDetails.DeliveryMedium),
				CodeDeliveryDestination: aws.StringValue(out.CodeDeliveryDetails.Destination),
			},
		},
	}, nil
}

func (s *server) forgotPassword(ctx context.Context, r *apipb.ForgotPasswordRequest) (*apipb.AuthResponse, error) {
	out, err := s.c.ForgotPasswordWithContext(ctx, &cognitoidentityprovider.ForgotPasswordInput{
		ClientId: aws.String(s.clientId),
		Username: aws.String(r.Username),
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}
	return &apipb.AuthResponse{
		AuthResponseSelect: &apipb.AuthResponse_ResponseForgotPassword{
			ResponseForgotPassword: &apipb.ForgotPasswordResponse{
				CodeDeliveryType:        aws.StringValue(out.CodeDeliveryDetails.DeliveryMedium),
				CodeDeliveryDestination: aws.StringValue(out.CodeDeliveryDetails.Destination),
			},
		},
	}, nil
}

func (s *server) confirmForgotPassword(ctx context.Context, r *apipb.ConfirmForgotPasswordRequest) (*apipb.AuthResponse, error) {
	_, err := s.c.ConfirmForgotPasswordWithContext(ctx, &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         aws.String(s.clientId),
		Username:         aws.String(r.Username),
		Password:         aws.String(r.Password),
		ConfirmationCode: aws.String(r.ConfirmationCode),
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.InvalidPasswordException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}
	return &apipb.AuthResponse{}, nil
}

func (s *server) signIn(ctx context.Context, r *apipb.SignInRequest) (*apipb.AuthResponse, error) {
	out, err := s.c.InitiateAuthWithContext(ctx, &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.clientId),
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		AuthParameters: map[string]*string{
			"USERNAME": aws.String(r.Username),
			"PASSWORD": aws.String(r.Password),
		},
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.PasswordResetRequiredException); ok {
			return nil, status.Errorf(codes.FailedPrecondition, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.UserNotFoundException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.UserNotConfirmedException); ok {
			return nil, status.Errorf(codes.FailedPrecondition, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}

	return &apipb.AuthResponse{
		AuthResponseSelect: &apipb.AuthResponse_ResponseSignIn{
			ResponseSignIn: &apipb.SignInResponse{
				Token:        aws.StringValue(out.AuthenticationResult.IdToken),
				RefreshToken: aws.StringValue(out.AuthenticationResult.RefreshToken),
			},
		},
	}, nil
}

func (s *server) tokenRefresh(ctx context.Context, r *apipb.TokenRefreshRequest) (*apipb.AuthResponse, error) {
	out, err := s.c.InitiateAuthWithContext(ctx, &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(s.clientId),
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(r.RefreshToken),
		},
	})
	if err != nil {
		if e, ok := err.(*cognitoidentityprovider.InvalidParameterException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.PasswordResetRequiredException); ok {
			return nil, status.Errorf(codes.FailedPrecondition, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.UserNotFoundException); ok {
			return nil, status.Errorf(codes.InvalidArgument, e.GoString())
		} else if e, ok := err.(*cognitoidentityprovider.UserNotConfirmedException); ok {
			return nil, status.Errorf(codes.FailedPrecondition, e.GoString())
		}
		return nil, status.Errorf(codes.Internal, "SignUp: %v", err)
	}

	return &apipb.AuthResponse{
		AuthResponseSelect: &apipb.AuthResponse_ResponseTokenRefresh{
			ResponseTokenRefresh: &apipb.TokenRefreshResponse{
				Token: aws.StringValue(out.AuthenticationResult.IdToken),
			},
		},
	}, nil
}

func (s *server) Serve(ctx context.Context, payload []byte) ([]byte, error) {
	request := &apipb.AuthRequest{}
	if err := protojson.Unmarshal(payload, request); err != nil {
		return nil, err
	}

	var err error
	var res *apipb.AuthResponse
	switch t := request.AuthRequestSelect.(type) {
	case *apipb.AuthRequest_RequestSignUp:
		res, err = s.signUp(ctx, t.RequestSignUp)
	case *apipb.AuthRequest_RequestConfirmSignUp:
		res, err = s.confirmSignUp(ctx, t.RequestConfirmSignUp)
	case *apipb.AuthRequest_RequestSignIn:
		res, err = s.signIn(ctx, t.RequestSignIn)
	case *apipb.AuthRequest_RequestTokenRefresh:
		res, err = s.tokenRefresh(ctx, t.RequestTokenRefresh)
	default:
		return nil, status.Errorf(codes.Unimplemented, "Unimplemented")
	}
	if err != nil {
		return nil, err
	}

	return protojson.Marshal(res)
}

func main() {
	ctx := context.Background()

	region := os.Getenv("REGION")

	session := session.New(aws.NewConfig())

	s := &server{
		clientId: os.Getenv("COGNITO_CLIENT_ID"),
		c:        cognitoidentityprovider.New(session, aws.NewConfig().WithRegion(region)),
	}

	h := lambda.NewHandler(s)

	h.Start(ctx)
}
