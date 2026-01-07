// Copyright (c) Kyle Huggins and contributors
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"log/slog"

	goversion "github.com/caarlos0/go-version"
)

func main() {
	slog.Info("skopos svcmon starting", "version", goversion.GetVersionInfo().GitVersion)
}
