package api

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/rpaas-operator/rpaas"
	"github.com/tsuru/rpaas-operator/rpaas/fake"
)

func Test_cachePurge(t *testing.T) {
	testCases := []struct {
		description  string
		instanceName string
		requestBody  string
		expectedCode int
		expectedBody string
		manager      rpaas.RpaasManager
	}{
		{
			description:  "returns bad request when request body is empty",
			instanceName: "my-instance",
			requestBody:  "",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Request body can't be empty",
			manager:      &fake.RpaasManager{},
		},
		{
			description:  "returns 400 when manager returns ValidationError",
			instanceName: "my-instance",
			requestBody:  "path=/index.html&preserve_path=true",
			expectedCode: http.StatusBadRequest,
			expectedBody: "Some validation failed",
			manager: &fake.RpaasManager{
				FakePurgeCache: func(instanceName string, args rpaas.PurgeCacheArgs) (int, error) {
					return 0, rpaas.ValidationError{Msg: "Some validation failed"}
				},
			},
		},
		{
			description:  "returns not found when manager returns NotFoundError",
			instanceName: "my-instance",
			requestBody:  "path=/index.html&preserve_path=true",
			expectedCode: http.StatusNotFound,
			expectedBody: "Something was not found",
			manager: &fake.RpaasManager{
				FakePurgeCache: func(instanceName string, args rpaas.PurgeCacheArgs) (int, error) {
					return 0, rpaas.NotFoundError{Msg: "Something was not found"}
				},
			},
		},
		{
			description:  "returns conflict when manager returns ConflictError",
			instanceName: "my-instance",
			requestBody:  "path=/index.html&preserve_path=true",
			expectedCode: http.StatusConflict,
			expectedBody: "Something already exists",
			manager: &fake.RpaasManager{
				FakePurgeCache: func(instanceName string, args rpaas.PurgeCacheArgs) (int, error) {
					return 0, rpaas.ConflictError{Msg: "Something already exists"}
				},
			},
		},
		{
			description:  "returns OK with the total number of servers where the cache was successfully purged",
			instanceName: "my-instance",
			requestBody:  "path=/index.html&preserve_path=true",
			expectedCode: http.StatusOK,
			expectedBody: "Path found and purged on 36 servers",
			manager: &fake.RpaasManager{
				FakePurgeCache: func(instanceName string, args rpaas.PurgeCacheArgs) (int, error) {
					return 36, nil
				},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.description, func(t *testing.T) {
			srv := newTestingServer(t, tt.manager)
			defer srv.Close()
			path := fmt.Sprintf("%s/resources/%s/purge", srv.URL, tt.instanceName)
			request, err := http.NewRequest(http.MethodPost, path, strings.NewReader(tt.requestBody))
			require.NoError(t, err)
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rsp, err := srv.Client().Do(request)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedCode, rsp.StatusCode)
			assert.Regexp(t, tt.expectedBody, bodyContent(rsp))
		})
	}
}
