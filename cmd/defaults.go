package cmd

import (
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

var DEFAULT_BASE = name.MustParseReference("gcr.io/distroless/nodejs:16")
var DEFAULT_PLATFORM, _ = v1.ParsePlatform("linux/arm64")
var DEFAULT_MTIME = time.UnixMicro(0)
