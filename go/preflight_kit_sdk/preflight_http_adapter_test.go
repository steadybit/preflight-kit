package preflight_kit_sdk

import (
	"github.com/steadybit/preflight-kit/go/preflight_kit_api"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockResponseWriter struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.Headers == nil {
		m.Headers = make(http.Header)
	}
	return m.Headers
}

func (m *mockResponseWriter) Write(body []byte) (int, error) {
	m.Body = body
	return len(body), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.StatusCode = statusCode
}

func Test_parseStartRequest(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		body []byte
	}
	tests := []struct {
		name  string
		args  args
		want  preflight_kit_api.StartPreflightRequestBody
		want1 error
		want2 bool
	}{
//		{
//			name: "valid_json",
//			args: args{
//				w:    &mockResponseWriter{},
//				body: []byte(`{
//  "preflightActionExecutionId" : "01958b44-6a7f-79dd-900b-e0aedf554be7",
//  "experimentExecution" : {
//    "id" : 267,
//    "key" : "ADM-9",
//    "name" : "asd",
//    "hypothesis" : null,
//    "requested" : "2025-03-12T16:51:11.569986Z",
//    "created" : "2025-03-12T16:51:11.616908Z",
//    "started" : null,
//    "experimentVersion" : 2,
//    "createdBy" : {
//      "username" : "admin",
//      "name" : "admin",
//      "pictureUrl" : "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCA0OCA0OCIgd2lkdGg9IjQ4IiBoZWlnaHQ9IjQ4Ij48cmVjdCBmaWxsPSIjMkE4MUE2IiB4PSIwIiB5PSIwIiB3aWR0aD0iNDgiIGhlaWdodD0iNDgiLz48dGV4dCB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmb250LWZhbWlseT0iSW50ZXIgVUksIHNhbnMtc2VyaWYiIHg9IjI0IiB5PSIzMiIgZm9udC1zaXplPSIyNCIgZmlsbD0iI2ZmZiI+QTwvdGV4dD48L3N2Zz4=",
//      "email" : null
//    },
//    "createdVia" : "UI",
//    "steps" : [ {
//      "id" : "01958b44-6a53-713b-ba05-d116305616f3",
//      "stepType" : "WAIT",
//      "predecessorId" : null,
//      "state" : "PREPARED",
//      "started" : null,
//      "ended" : null,
//      "reason" : null,
//      "ignoreFailure" : false,
//      "parameters" : {
//        "duration" : "10s"
//      },
//      "customLabel" : null
//    } ],
//    "canceledBy" : null,
//    "ended" : null,
//    "state" : "CREATED",
//    "reason" : null,
//    "variables" : { }
//  }
//}`),
//			},
//			want: preflight_kit_api.StartPreflightRequestBody{
//				PreflightActionExecutionId: uuid.MustParse("01958b44-6a7f-79dd-900b-e0aedf554be7"),
//
//			},
//			want1: nil,
//			want2: false,
//		},
		{
			name: "invalid_json",
			args: args{
				w:    &mockResponseWriter{},
				body: []byte(`{invalid json}`),
			},
			want:  preflight_kit_api.StartPreflightRequestBody{},
			want1: nil,
			want2: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := parseStartRequest(tt.args.w, tt.args.body)
			assert.Equalf(t, tt.want, got, "parseStartRequest(%v, %v)", tt.args.w, tt.args.body)
			assert.Equalf(t, tt.want1, got1, "parseStartRequest(%v, %v)", tt.args.w, tt.args.body)
			assert.Equalf(t, tt.want2, got2, "parseStartRequest(%v, %v)", tt.args.w, tt.args.body)
		})
	}
}
