package network

import (
	"bytes"
	"go-chain/core"
	"go-chain/network"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetBlocksMessage(t *testing.T) {
	msg := &network.GetBlocksMessage{
		From: 1,
		To:   10,
	}

	var buf bytes.Buffer
	err := msg.Encode(&buf)
	assert.NoError(t, err)

	decodedMsg := &network.GetBlocksMessage{}
	err = decodedMsg.Decode(&buf)
	assert.NoError(t, err)

	assert.Equal(t, msg.From, decodedMsg.From)
	assert.Equal(t, msg.To, decodedMsg.To)
}

func TestBlocksMessage(t *testing.T) {
	blocks := []*core.Block{
		{Header: &core.BlockHeader{Height: 1}},
		{Header: &core.BlockHeader{Height: 2}},
	}
	msg := &network.BlocksMessage{
		Blocks: blocks,
	}

	var buf bytes.Buffer
	err := msg.Encode(&buf)
	assert.NoError(t, err)

	decodedMsg := &network.BlocksMessage{}
	err = decodedMsg.Decode(&buf)
	assert.NoError(t, err)

	assert.Equal(t, len(msg.Blocks), len(decodedMsg.Blocks))
	for i, block := range msg.Blocks {
		assert.Equal(t, block.Header.Height, decodedMsg.Blocks[i].Header.Height)
	}
}

func TestGetStatusMessage(t *testing.T) {
	msg := &network.GetStatusMessage{}

	var buf bytes.Buffer
	err := msg.Encode(&buf)
	assert.NoError(t, err)

	decodedMsg := &network.GetStatusMessage{}
	err = decodedMsg.Decode(&buf)
	assert.NoError(t, err)
}

func TestStatusMessage(t *testing.T) {
	msg := &network.StatusMessage{
		ID:            "testID",
		Version:       1,
		CurrentHeight: 100,
	}

	var buf bytes.Buffer
	err := msg.Encode(&buf)
	assert.NoError(t, err)

	decodedMsg := &network.StatusMessage{}
	err = decodedMsg.Decode(&buf)
	assert.NoError(t, err)

	assert.Equal(t, msg.ID, decodedMsg.ID)
	assert.Equal(t, msg.Version, decodedMsg.Version)
	assert.Equal(t, msg.CurrentHeight, decodedMsg.CurrentHeight)
}
