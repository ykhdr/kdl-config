package reflectutils

import (
	"reflect"
	"testing"
)

func TestGetNumericValue(t *testing.T) {
	type args struct {
		fv reflect.Value
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			name:    "int value",
			args:    args{fv: reflect.ValueOf(int(42))},
			want:    42,
			wantErr: false,
		},
		{
			name:    "uint value",
			args:    args{fv: reflect.ValueOf(uint(42))},
			want:    42,
			wantErr: false,
		},
		{
			name:    "float value",
			args:    args{fv: reflect.ValueOf(float64(42.42))},
			want:    42.42,
			wantErr: false,
		},
		{
			name:    "unsupported type",
			args:    args{fv: reflect.ValueOf("string")},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetNumericValue(tt.args.fv)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNumericValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetNumericValue() got = %v, want %v", got, tt.want)
			}
		})
	}
}
