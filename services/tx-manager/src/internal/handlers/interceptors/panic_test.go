package interceptors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoveryUnaryInterceptor(t *testing.T) {
	tests := []struct {
		name        string
		handler     grpc.UnaryHandler
		wantErrCode codes.Code
		wantErrMsg  string
	}{
		{
			name: "no panic - returns handler result",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return "ok", nil
			},
			wantErrCode: codes.OK,
			wantErrMsg:  "",
		},
		{
			name: "panic - interceptor recovers",
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				panic("boom")
			},
			wantErrCode: codes.Internal,
			wantErrMsg:  "Panic: `/test.Foo`",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &grpc.UnaryServerInfo{FullMethod: "/test.Foo"}

			resp, err := RecoveryUnaryInterceptor(context.Background(), "req", info, tt.handler)
			if tt.wantErrCode == codes.OK {
				assert.Nil(t, err, tt.name)
				assert.Equal(t, "ok", resp, tt.name)

				return
			}

			assert.NotNil(t, err, tt.name)
			st, ok := status.FromError(err)
			assert.True(t, ok, tt.name)
			assert.Equal(t, tt.wantErrCode, st.Code(), tt.name)
			assert.Contains(t, st.Message(), tt.wantErrMsg, tt.name)
			assert.Contains(t, st.Message(), "goroutine", tt.name)
		})
	}
}
