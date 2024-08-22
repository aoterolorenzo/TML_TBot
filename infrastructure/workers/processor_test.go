package workers

import (
	"TML_TBot/application/interfaces"
	"TML_TBot/application/usecases"
	"reflect"
	"testing"
)

func TestParseUseCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want interfaces.UseCase
	}{
		{
			name: "test de job desconocido",
			args: args{str: "jobDesconocido"},
			want: nil,
		},
		{
			name: "test de weather",
			args: args{str: "weather"},
			want: &usecases.WeatherController{},
		},
		{
			name: "test de lineUp",
			args: args{str: "lineUp"},
			want: usecases.NewTMLLineUpController(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseUseCase(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseUseCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
