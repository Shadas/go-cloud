// Copyright 2019 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package awskms

import (
	"context"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"gocloud.dev/internal/testing/setup"
	"gocloud.dev/secrets/driver"
	"gocloud.dev/secrets/drivertest"
)

const (
	keyID  = "alias/test-secrets"
	region = "us-west-1"
)

type harness struct {
	client *kms.KMS
	close  func()
}

func (h *harness) MakeDriver(ctx context.Context) (driver.Keeper, error) {
	return &keeper{
		keyID:  keyID,
		client: h.client,
	}, nil
}

func (h *harness) Close() {
	h.close()
}

func newHarness(ctx context.Context, t *testing.T) (drivertest.Harness, error) {
	sess, _, done := setup.NewAWSSession(t, region)
	return &harness{
		client: kms.New(sess),
		close:  done,
	}, nil
}

func TestConformance(t *testing.T) {
	drivertest.RunConformanceTests(t, newHarness)
}

// KMS-specific tests.

func TestNoSessionProvidedError(t *testing.T) {
	if _, err := Dial(nil); err == nil {
		t.Error("got nil, want no AWS session provided")
	}
}

func TestNoConnectionError(t *testing.T) {
	prevAccessKey := os.Getenv("AWS_ACCESS_KEY")
	prevSecretKey := os.Getenv("AWS_SECRET_KEY")
	prevRegion := os.Getenv("AWS_REGION")
	os.Setenv("AWS_ACCESS_KEY", "myaccesskey")
	os.Setenv("AWS_SECRET_KEY", "mysecretkey")
	os.Setenv("AWS_REGION", "us-east-1")
	defer func() {
		os.Setenv("AWS_ACCESS_KEY", prevAccessKey)
		os.Setenv("AWS_SECRET_KEY", prevSecretKey)
		os.Setenv("AWS_REGION", prevRegion)
	}()
	sess, err := session.NewSession()
	if err != nil {
		t.Fatal(err)
	}

	client, err := Dial(sess)
	if err != nil {
		t.Fatal(err)
	}
	keeper := NewKeeper(client, keyID, nil)

	if _, err := keeper.Encrypt(context.Background(), []byte("test")); err == nil {
		t.Error("got nil, want UnrecognizedClientException")
	}
}
