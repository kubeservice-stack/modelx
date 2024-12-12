/*
Copyright 2022 The KubeService-Stack Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/blang/semver/v4"
	"gopkg.in/yaml.v3"

	"kubegems.io/modelx/test/framework"
)

var (
	testframework *framework.Framework
	testImage     *string
)

type Version struct {
	Variables map[string]string `yaml:"variables"`
}

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		"",
		"kube config path, e.g. $HOME/.kube/config",
	)
	testImage = flag.String(
		"test-image",
		"",
		"test image, e.g. docker.io/openlocal/open-local",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

	logger := log.New(os.Stdout, "", log.Lshortfile)

	fp, err := os.Open("../../VERSION.yml")
	if err != nil {
		logger.Printf("failed to open version file: %v\n", err)
		os.Exit(1)
	}
	defer fp.Close()

	dec := yaml.NewDecoder(fp)
	// 配置对象
	var verison Version
	err = dec.Decode(&verison)
	if err != nil {
		logger.Printf("failed to read version fp: %v\n", err)
		os.Exit(1)
	}

	currentSemVer, err := semver.ParseTolerant(verison.Variables["TAG_REALSE_VERSION"])
	if err != nil {
		logger.Printf("failed to parse current version: %v\n", err)
		os.Exit(1)
	}

	exampleDir := "../../"
	resourcesDir := "../framework/resources"

	nextSemVer, err := semver.ParseTolerant(fmt.Sprintf("0.%d.0", currentSemVer.Minor))
	if err != nil {
		logger.Printf("failed to parse next version: %v\n", err)
		os.Exit(1)
	}

	if testframework, err = framework.New(*kubeconfig, *testImage, exampleDir, resourcesDir, nextSemVer); err != nil {
		logger.Printf("failed to setup fk: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

func TestE2E(t *testing.T) {
	testCtx := testframework.NewTestCtx(t)
	defer testCtx.Cleanup(t)
	t.Run("TestDeployment", testDeplayment(context.Background()))
}
