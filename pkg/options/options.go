package options

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/spf13/viper"
)

const (
	// configDefaultBaseImage is the default base image if not specified in .ko.yaml.
	configDefaultBaseImage = "gcr.io/distroless/nodejs:16"
)

type BuildOptions struct {
	BaseImage        string
	WorkingDirectory string
	CreationTime     v1.Time
}

func getTimeFromEnv(env string) (*v1.Time, error) {
	epoch := os.Getenv(env)
	if epoch == "" {
		return nil, nil
	}

	seconds, err := strconv.ParseInt(epoch, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("the environment variable %s should be the number of seconds since January 1st 1970, 00:00 UTC, got: %w", env, err)
	}
	return &v1.Time{Time: time.Unix(seconds, 0)}, nil
}

func getCreationTime() (*v1.Time, error) {
	return getTimeFromEnv("SOURCE_DATE_EPOCH")
}

func (buildops *BuildOptions) Load() error {
	if buildops.WorkingDirectory == "" {
		buildops.WorkingDirectory = "./"
	}

	creationTime, err := getCreationTime()
	if err != nil {
		return err
	}

	if creationTime != nil {
		buildops.CreationTime = *creationTime
	} else {
		buildops.CreationTime = v1.Time{}
	}

	vp := viper.New()
	vp.SetDefault("defaultBaseImage", configDefaultBaseImage)
	vp.SetConfigName(".no")
	vp.SetEnvPrefix("no")
	vp.AutomaticEnv()
	vp.AddConfigPath(buildops.WorkingDirectory)

	if err := vp.ReadInConfig(); err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	if buildops.BaseImage == "" {
		ref := vp.GetString("defaultBaseImage")
		if _, err := name.ParseReference(ref); err != nil {
			return fmt.Errorf("'defaultBaseImage': error parsing %q as image reference: %w", ref, err)
		}
		buildops.BaseImage = ref
	}

	return nil
}
