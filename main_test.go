package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	Acct = "112168818644504200034"
)

func TestReturnOne(t *testing.T) {
	if ReturnOne() != 1 {
		t.Error("Expected 1")
	}
}

func TestGetBucketLs(t *testing.T) {
	err := ListBucketContents("midi_file_storage")
	require.NoError(t, err)

}

func TestInitGCP(t *testing.T) {
	_, err := InitGCPWithServiceAccount(GCP_project, "/Users/jesselopez/Documents/repos/midi-file-server/gothic_key.json")
	require.NoError(t, err)
}
