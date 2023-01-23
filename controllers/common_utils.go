package controllers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlUtil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

type RateLimiter struct {
	Burst           int
	Frequency       int
	BaseDelay       time.Duration
	FailureMaxDelay time.Duration
}

const (
	requeueInterval = time.Second * 3
	finalizer       = "sample.kyma-project.io/finalizer"
	debugLogLevel   = 2
	fieldOwner      = "sample.kyma-project.io/owner"
)

// parseManifestStringToObjects parses the string of resources into a list of unstructured resources.
func parseManifestStringToObjects(manifest string) (*ManifestResources, error) {
	objects := &ManifestResources{}
	reader := yamlUtil.NewYAMLReader(bufio.NewReader(strings.NewReader(manifest)))
	for {
		rawBytes, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return objects, nil
			}

			return nil, fmt.Errorf("invalid YAML doc: %w", err)
		}

		rawBytes = bytes.TrimSpace(rawBytes)
		unstructuredObj := unstructured.Unstructured{}
		if err := yaml.Unmarshal(rawBytes, &unstructuredObj); err != nil {
			objects.Blobs = append(objects.Blobs, append(bytes.TrimPrefix(rawBytes, []byte("---\n")), '\n'))
		}

		if len(rawBytes) == 0 || bytes.Equal(rawBytes, []byte("null")) || len(unstructuredObj.Object) == 0 {
			continue
		}

		objects.Items = append(objects.Items, &unstructuredObj)
	}
}
