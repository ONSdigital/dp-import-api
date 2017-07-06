package utils

import "testing"
import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
)

func TestGetEnvValue(t *testing.T) {
	Convey("When a env value is set, the value is returned", t, func() {
		envName := "PORT"
		envValue := "8080"
		os.Setenv(envName, envValue)
		So(GetEnvVariable(envName, "0000"), ShouldEqual, envValue)
	})
}

func TestGetDefaultValue(t *testing.T) {
	Convey("When no env value is set, the default value is returned", t, func() {
		envName := "PORT"
		defaultValue := "8080"
		// Force the value to be nothing
		os.Setenv(envName, "")
		So(GetEnvVariable(envName, defaultValue), ShouldEqual, defaultValue)
	})
}
