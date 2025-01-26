// SPDX-FileCopyrightText: 2023 Christoph Mewes
// SPDX-License-Identifier: MIT

package loader

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"

	"go.xrstf.de/pkiplot/pkg/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
)

type Options struct {
	Namespace      string
	FileExtensions []string
}

func NewDefaultOptions() *Options {
	return &Options{
		FileExtensions: []string{"yaml", "yml"},
	}
}

func LoadPKI(sources []string, opt *Options) (*types.PKI, error) {
	if len(sources) == 0 {
		return nil, nil
	}

	if opt == nil {
		opt = NewDefaultOptions()
	}

	result := &types.PKI{
		Secrets:        []corev1.Secret{},
		Certificates:   []certmanagerv1.Certificate{},
		Issuers:        []certmanagerv1.Issuer{},
		ClusterIssuers: []certmanagerv1.ClusterIssuer{},
	}

	// load from all sources

	for _, source := range sources {
		if err := loadManifestsSource(result, opt, source); err != nil {
			return nil, fmt.Errorf("failed to load from %q: %w", source, err)
		}
	}

	// forbid duplicates

	identifiers := sets.New[string]()
	for idx, cert := range result.Certificates {
		ident, err := getResourceIdentifier(&cert)
		if err != nil {
			return nil, fmt.Errorf("Certificate %d is invalid: %w", idx, err)
		}

		if identifiers.Has(ident) {
			return nil, fmt.Errorf("found multiple definitions for Certificate %s", ident)
		}
	}

	identifiers = sets.New[string]()
	for idx, secret := range result.Secrets {
		ident, err := getResourceIdentifier(&secret)
		if err != nil {
			return nil, fmt.Errorf("Secret %d is invalid: %w", idx, err)
		}

		if identifiers.Has(ident) {
			return nil, fmt.Errorf("found multiple definitions for Secret %s", ident)
		}
	}

	identifiers = identifiers.Clear()
	for idx, issuer := range result.Issuers {
		ident, err := getResourceIdentifier(&issuer)
		if err != nil {
			return nil, fmt.Errorf("Issuer %d is invalid: %w", idx, err)
		}

		if identifiers.Has(ident) {
			return nil, fmt.Errorf("found multiple definitions for Issuer %s", ident)
		}
	}

	identifiers = identifiers.Clear()
	for idx, clusterIssuer := range result.ClusterIssuers {
		ident, err := getResourceIdentifier(&clusterIssuer)
		if err != nil {
			return nil, fmt.Errorf("ClusterIssuer %d is invalid: %w", idx, err)
		}

		if identifiers.Has(ident) {
			return nil, fmt.Errorf("found multiple definitions for ClusterIssuer %s", ident)
		}
	}

	// sort all lists to ensure a stable output

	sort.Slice(result.Secrets, func(i, j int) bool {
		return resourceIsLess(&result.Secrets[i], &result.Secrets[j])
	})

	sort.Slice(result.Certificates, func(i, j int) bool {
		return resourceIsLess(&result.Certificates[i], &result.Certificates[j])
	})

	sort.Slice(result.Issuers, func(i, j int) bool {
		return resourceIsLess(&result.Issuers[i], &result.Issuers[j])
	})

	sort.Slice(result.ClusterIssuers, func(i, j int) bool {
		return resourceIsLess(&result.ClusterIssuers[i], &result.ClusterIssuers[j])
	})

	return result, nil
}

func getResourceIdentifier(res metav1.Object) (string, error) {
	base, err := getResourceName(res)
	if err != nil {
		return "", err
	}

	if ns := res.GetNamespace(); ns != "" {
		return ns + "/" + base, nil
	}

	return base, nil
}

func getResourceName(res metav1.Object) (string, error) {
	base := res.GetName()
	if base != "" {
		return base, nil
	}

	base = res.GetGenerateName()
	if base != "" {
		return base, nil
	}

	return "", errors.New("resource has neither name nor generateName")
}

func resourceIsLess(a, b metav1.Object) bool {
	nsA := a.GetNamespace()
	nsB := b.GetNamespace()

	// cluster-wide resources always come before (are "less") than namespaced

	// if scope differs...
	if (nsA != "") != (nsB != "") {
		return nsA == "" // then a < b if a is cluster-wide
	}

	// if both are in different namespaces...
	if nsA != nsB {
		return nsA < nsB
	}

	nameA, _ := getResourceName(a)
	nameB, _ := getResourceName(b)

	// if both are in the same namespace or no namespace...
	return nameA < nameB
}

func loadManifestsSource(result *types.PKI, opt *Options, source string) error {
	if source == "-" {
		// thank you https://stackoverflow.com/a/26567513
		stat, _ := os.Stdin.Stat()
		if stat.Mode()&os.ModeCharDevice != 0 {
			return errors.New("no data provided on stdin")
		}

		return loadManifestsSourceReader(result, opt, os.Stdin)
	}

	stat, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	if stat.IsDir() {
		absSource, err := filepath.Abs(source)
		if err != nil {
			return fmt.Errorf("failed to determine absolute path: %w", err)
		}

		return loadManifestsSourceDirectory(result, opt, absSource)
	}

	return loadManifestsSourceFile(result, opt, source)
}

const (
	bufSize = 5 * 1024 * 1024
)

func loadManifestsSourceFile(result *types.PKI, opt *Options, source string) error {
	f, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return loadManifestsSourceReader(result, opt, f)
}

func loadManifestsSourceReader(result *types.PKI, opt *Options, source io.ReadCloser) error {
	docSplitter := yamlutil.NewDocumentDecoder(source)
	defer docSplitter.Close()

	for i := 1; true; i++ {
		buf := make([]byte, bufSize) // 5 MB, same as chunk size in decoder
		read, err := docSplitter.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("document %d is larger than the internal buffer", i)
		}

		fileContents, err := parseFileContents(opt, buf[:read])
		if err != nil {
			if errors.Is(err, io.EOF) {
				continue
			}
			return fmt.Errorf("document %d is invalid: %w", i, err)
		}

		if fileContents == nil {
			continue
		}

		result.Secrets = append(result.Secrets, fileContents.Secrets...)
		result.Certificates = append(result.Certificates, fileContents.Certificates...)
		result.Issuers = append(result.Issuers, fileContents.Issuers...)
		result.ClusterIssuers = append(result.ClusterIssuers, fileContents.ClusterIssuers...)
	}

	return nil
}

func loadManifestsSourceDirectory(result *types.PKI, opt *Options, rootDir string) error {
	contents, err := os.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range contents {
		fullPath := filepath.Join(rootDir, entry.Name())

		if entry.IsDir() {
			if err := loadManifestsSourceDirectory(result, opt, fullPath); err != nil {
				return fmt.Errorf("failed to read directory %s: %w", fullPath, err)
			}
		} else if hasExtension(entry.Name(), opt.FileExtensions) {
			if err := loadManifestsSourceFile(result, opt, fullPath); err != nil {
				return fmt.Errorf("failed to read file %s: %w", fullPath, err)
			}
		}
	}

	return nil
}

func hasExtension(filename string, extensions []string) bool {
	parts := strings.Split(filename, ".")
	extension := parts[len(parts)-1]

	for _, ext := range extensions {
		if ext == extension {
			return true
		}
	}

	return false
}

func parseFileContents(opt *Options, data []byte) (*types.PKI, error) {
	candidate := unstructured.Unstructured{}

	err := yamlutil.NewYAMLOrJSONDecoder(bytes.NewReader(data), 1024).Decode(&candidate)
	if err != nil {
		return nil, fmt.Errorf("document is not valid Kubernetes YAML: %w", err)
	}

	result := &types.PKI{
		Secrets:        []corev1.Secret{},
		Certificates:   []certmanagerv1.Certificate{},
		Issuers:        []certmanagerv1.Issuer{},
		ClusterIssuers: []certmanagerv1.ClusterIssuer{},
	}

	if err := parseUnstructured(opt, candidate, result); err != nil {
		return nil, err
	}

	return result, nil
}

func parseUnstructured(opt *Options, candidate unstructured.Unstructured, result *types.PKI) error {
	// recurse into lists
	if candidate.IsList() {
		list, err := candidate.ToList()
		if err != nil {
			return fmt.Errorf("object looks like List, but: %w", err)
		}

		for _, obj := range list.Items {
			if err := parseUnstructured(opt, obj, result); err != nil {
				return err
			}
		}

		return nil
	}

	makeError := func(kind string, err error) error {
		return fmt.Errorf("document is not valid %s: %w", kind, err)
	}

	switch candidate.GroupVersionKind().GroupKind().String() {
	case "Secret":
		secret := corev1.Secret{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(candidate.Object, &secret); err != nil {
			return makeError("Secret", err)
		}

		// ignore non-TLS secrets as they should not influence the PKI structure (there might be a secret
		// for some ACME stuff, but those are not relevant here)
		if secret.Type != corev1.SecretTypeTLS {
			return nil
		}

		if err := injectNamespace(&secret, opt); err != nil {
			return makeError("Secret", err)
		}
		if resourceMatchesOpt(&secret, opt) {
			result.Secrets = append(result.Secrets, secret)
		}

	case "Certificate.cert-manager.io":
		cert := certmanagerv1.Certificate{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(candidate.Object, &cert); err != nil {
			return makeError("Certificate", err)
		}
		if err := injectNamespace(&cert, opt); err != nil {
			return makeError("Certificate", err)
		}
		if resourceMatchesOpt(&cert, opt) {
			result.Certificates = append(result.Certificates, cert)
		}

	case "Issuer.cert-manager.io":
		issuer := certmanagerv1.Issuer{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(candidate.Object, &issuer); err != nil {
			return makeError("Issuer", err)
		}
		if err := injectNamespace(&issuer, opt); err != nil {
			return makeError("Issuer", err)
		}
		if resourceMatchesOpt(&issuer, opt) {
			result.Issuers = append(result.Issuers, issuer)
		}

	case "ClusterIssuer.cert-manager.io":
		clusterIssuer := certmanagerv1.ClusterIssuer{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(candidate.Object, &clusterIssuer); err != nil {
			return makeError("ClusterIssuer", err)
		}
		// strip out misleading metadata
		clusterIssuer.Namespace = ""

		result.ClusterIssuers = append(result.ClusterIssuers, clusterIssuer)
	}

	return nil
}

func injectNamespace(res metav1.Object, opt *Options) error {
	if res.GetNamespace() == "" {
		if opt.Namespace == "" {
			return errors.New("no metadata.namespace set and no --namespace provided")
		}

		res.SetNamespace(opt.Namespace)
	}

	return nil
}

func resourceMatchesOpt(res metav1.Object, opt *Options) bool {
	return opt.Namespace == "" || res.GetNamespace() == opt.Namespace
}
