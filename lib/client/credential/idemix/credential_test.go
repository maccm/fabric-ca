/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package idemix_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	lib "github.com/hyperledger/fabric-ca/lib"
	. "github.com/hyperledger/fabric-ca/lib/client/credential/idemix"
	"github.com/stretchr/testify/assert"
)

const (
	testDataDir          = "../../../../testdata"
	testSignerConfigFile = testDataDir + "/IdemixSignerConfig"
)

func TestIdemixCredential(t *testing.T) {
	clientHome, err := ioutil.TempDir(testDataDir, "idemixcredtest")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %s", err.Error())
	}
	defer os.RemoveAll(clientHome)

	signerConfig := filepath.Join(clientHome, "SignerConfig")
	client := &lib.Client{
		Config: &lib.ClientConfig{
			URL: fmt.Sprintf("http://localhost:7054"),
		},
		HomeDir: clientHome,
	}

	idemixCred := NewCredential(signerConfig, client)

	assert.Equal(t, idemixCred.Type(), CredType, "Type for a IdemixCredential instance must be Idemix")
	_, err = idemixCred.Val()
	assert.Error(t, err, "Val should return error if credential has not been loaded from disk or set")
	if err != nil {
		assert.Equal(t, err.Error(), "Credential value is not set")
	}
	_, err = idemixCred.EnrollmentID()
	assert.Error(t, err, "EnrollmentID should return an error if credential has not been loaded from disk or set")
	if err != nil {
		assert.Equal(t, err.Error(), "Credential value is not set")
	}

	err = idemixCred.SetVal("hello")
	assert.Error(t, err, "SetVal should fail as it expects an object of type *SignerConfig")

	err = idemixCred.Store()
	assert.Error(t, err, "Store should return an error if credential has not been set")

	err = idemixCred.Load()
	assert.Error(t, err, "Load should fail as %s is not found", signerConfig)

	err = ioutil.WriteFile(signerConfig, []byte("hello"), 0744)
	if err != nil {
		t.Fatalf("Failed to write to file %s: %s", signerConfig, err.Error())
	}
	err = idemixCred.Load()
	assert.Error(t, err, "Load should fail as %s contains invalid data", signerConfig)

	err = lib.CopyFile(testSignerConfigFile, signerConfig)
	if err != nil {
		t.Fatalf("Failed to copy %s to %s: %s", testSignerConfigFile, signerConfig, err.Error())
	}

	err = idemixCred.Load()
	assert.NoError(t, err, "Load should not return error as %s exists and is valid", signerConfig)

	val, err := idemixCred.Val()
	assert.NoError(t, err, "Val should not return error as credential is loaded")

	signercfg, _ := val.(*SignerConfig)
	cred := signercfg.GetCred()
	assert.NotNil(t, cred)
	assert.True(t, len(cred) > 0, "Credential bytes length should be more than zero")
	enrollID := signercfg.GetEnrollmentID()
	assert.Equal(t, "admin", enrollID, "Enrollment ID of the Idemix credential in testdata/IdemixSignerConfig should be admin")
	sk := signercfg.GetSk()
	assert.NotNil(t, sk, "secret key should not be nil")
	assert.True(t, len(sk) > 0, "Secret key bytes length should be more than zero")
	signercfg.GetOrganizationalUnitIdentifier()
	isAdmin := signercfg.GetIsAdmin()
	assert.False(t, isAdmin)

	err = idemixCred.SetVal(val)
	assert.NoError(t, err, "Setting the value that we got from the credential should not return an error")

	if err = os.Chmod(signerConfig, 0000); err != nil {
		t.Fatalf("Failed to chmod SignerConfig file %s: %v", signerConfig, err)
	}
	err = idemixCred.Store()
	assert.Error(t, err, "Store should fail as %s is not writable", signerConfig)

	if err = os.Chmod(signerConfig, 0644); err != nil {
		t.Fatalf("Failed to chmod SignerConfig file %s: %v", signerConfig, err)
	}
	err = idemixCred.Store()
	assert.NoError(t, err, "Store should not fail as %s is writable and Idemix credential value is set", signerConfig)

	_, err = idemixCred.Val()
	assert.NoError(t, err, "Val should not return error as Idemix credential has been loaded")

	_, err = idemixCred.EnrollmentID()
	assert.Error(t, err, "EnrollmentID is not implemented for Idemix credential")

	body := []byte("hello")
	req, err := http.NewRequest("GET", "localhost:7054/enroll", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create HTTP request: %s", err.Error())
	}
	_, err = idemixCred.CreateOAuthToken(req, body)
	assert.Error(t, err, "CreateOAuthToken is not implemented for Idemix credential")

	_, err = idemixCred.RevokeSelf()
	assert.Error(t, err, "RevokeSelf should fail as it is not implmented for Idemix credential")
}