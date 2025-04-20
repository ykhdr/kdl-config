package kdlconfig

import (
	"os"
	"testing"
)

func TestLoader_Load(t *testing.T) {
	type args struct {
		cfg  interface{}
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid config",
			args: args{
				cfg: &struct{ Port int }{},
				path: func() string {
					f, err := os.CreateTemp("", "valid_config.kdl")
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					f.WriteString("port 8080")
					return f.Name()
				}(),
			},
			wantErr: false,
		},
		{
			name: "non-existent file",
			args: args{
				cfg:  &struct{ Port int }{},
				path: "non_existent_file.kdl",
			},
			wantErr: true,
		},
		{
			name: "invalid config",
			args: args{
				cfg: &struct{ Port int }{},
				path: func() string {
					f, err := os.CreateTemp("", "invalid_config.kdl")
					if err != nil {
						t.Fatal(err)
					}
					defer f.Close()
					f.WriteString("port: not_a_number")
					return f.Name()
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Loader{}
			if err := l.Load(tt.args.cfg, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
