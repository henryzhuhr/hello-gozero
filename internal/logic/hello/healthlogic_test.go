// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package hello

import (
	"context"
	"testing"

	"hello-gozero/internal/config"
	"hello-gozero/internal/svc"
	"hello-gozero/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthLogic_Health(t *testing.T) {
	c := config.Config{}
	mockSvcCtx := svc.NewServiceContext(c)
	// init mock service context here

	tests := []struct {
		name       string
		ctx        context.Context
		setupMocks func()

		wantErr   bool
		checkResp func(resp *types.Response, err error)
	}{
		{
			name: "response error",
			ctx:  context.Background(),
			setupMocks: func() {
				// mock data for this test case
			},

			wantErr: true,
			checkResp: func(resp *types.Response, err error) {
				// TODO: Add your check logic here
			},
		},
		{
			name: "successful",
			ctx:  context.Background(),
			setupMocks: func() {
				// Mock data for this test case
			},

			wantErr: false,
			checkResp: func(resp *types.Response, err error) {
				// TODO: Add your check logic here
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			l := NewHealthLogic(tt.ctx, mockSvcCtx)
			resp, err := l.Health()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
			}
			tt.checkResp(resp, err)
		})
	}
}
