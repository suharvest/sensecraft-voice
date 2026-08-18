package httputils

import "testing"

func Test_getObjectFromRequest(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantObj string
		wantSid string
		wantOk  bool
	}{
		{
			name:    "test0",
			path:    "",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test1",
			path:    "/",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test2",
			path:    "//",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test3",
			path:    "sensecraftVoice",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test4",
			path:    "/sensecraftVoice",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test5",
			path:    "/api/v1/",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test6",
			path:    "/api/v1/users",
			wantObj: "users",
			wantSid: "",
			wantOk:  true,
		},
		{
			name:    "test7",
			path:    "/api/v1/users/",
			wantObj: "users",
			wantOk:  false,
		},
		{
			name:    "test8",
			path:    "/api/v1/users/1",
			wantObj: "users",
			wantSid: "1",
			wantOk:  true,
		},
		{
			name:    "test9",
			path:    "/api/v1//",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test10",
			path:    "///",
			wantObj: "",
			wantOk:  false,
		},
		{
			name:    "test11",
			path:    "/api/v1/users/1/password",
			wantObj: "users",
			wantSid: "1",
			wantOk:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotObj, gotSid, gotOk := getObjectFromRequest(tt.path)
			if gotObj != tt.wantObj {
				t.Errorf("getObjectFromRequest() gotObj = %v, want %v", gotObj, tt.wantObj)
			}
			if gotSid != tt.wantSid {
				t.Errorf("getObjectFromRequest() gotSid = %v, want %v", gotSid, tt.wantSid)
			}
			if gotOk != tt.wantOk {
				t.Errorf("getObjectFromRequest() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
