package listener

import (
	"Alarm/internal/pkg/rule"
	"Alarm/internal/web/models"
	"reflect"
	"testing"
)

func TestNewListener(t *testing.T) {
	type args struct {
		url  string
		rcp  *models.Cache
		Id   int
		Rule map[int]rule.Rule
	}
	tests := []struct {
		name    string
		args    args
		want    *Listener
		wantErr bool
	}{
		{
			name: "url empty",
			args: args{
				url:  "",
				rcp:  &models.Cache{},
				Id:   1,
				Rule: map[int]rule.Rule{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rcp nil",
			args: args{
				url:  "amqp://user:mkjsix7@172.16.0.15:5672/",
				rcp:  nil,
				Id:   1,
				Rule: map[int]rule.Rule{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Id zero",
			args: args{
				url:  "amqp://user:mkjsix7@172.16.0.15:5672/",
				rcp:  &models.Cache{},
				Id:   0,
				Rule: map[int]rule.Rule{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Rule nil",
			args: args{
				url:  "amqp://user:mkjsix7@172.16.0.15:5672/",
				rcp:  &models.Cache{},
				Id:   1,
				Rule: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid arguments",
			args: args{
				url:  "amqp://user:mkjsix7@172.16.0.15:5672/",
				rcp:  &models.Cache{},
				Id:   1,
				Rule: map[int]rule.Rule{},
			},
			want:    &Listener{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewListener(tt.args.url, tt.args.rcp, tt.args.Id, tt.args.Rule)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewListener() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.TypeOf(tt.want) != reflect.TypeOf(got) {
				t.Errorf("NewListener() = %v, want %v", got, tt.want)
			}

		})
	}
}
