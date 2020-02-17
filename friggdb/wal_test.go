package friggdb

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/dgryski/go-farm"
	"github.com/grafana/frigg/pkg/friggpb"
	"github.com/grafana/frigg/pkg/util/test"
)

func TestCreateBlock(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(tempDir)
	assert.NoError(t, err, "unexpected error creating temp dir")

	wal, err := newWAL(&walConfig{
		filepath:        tempDir,
		indexDownsample: 2,
	})
	assert.NoError(t, err, "unexpected error creating temp wal")

	blockID := uuid.New()

	block, err := wal.NewBlock(blockID, testTenantID)
	assert.NoError(t, err, "unexpected error creating block")

	blocks, err := wal.AllBlocks()
	assert.NoError(t, err, "unexpected error getting blocks")
	assert.Len(t, blocks, 1)

	assert.Equal(t, block.(*headBlock).fullFilename(), blocks[0].(*headBlock).fullFilename())
}

func TestReadWrite(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(tempDir)
	assert.NoError(t, err, "unexpected error creating temp dir")

	wal, err := newWAL(&walConfig{
		filepath:        tempDir,
		indexDownsample: 2,
	})
	assert.NoError(t, err, "unexpected error creating temp wal")

	blockID := uuid.New()

	block, err := wal.NewBlock(blockID, testTenantID)
	assert.NoError(t, err, "unexpected error creating block")

	req := test.MakeRequest(10, []byte{0x00, 0x01})
	bReq, err := proto.Marshal(req)
	assert.NoError(t, err)
	err = block.Write([]byte{0x00, 0x01}, bReq)
	assert.NoError(t, err, "unexpected error creating writing req")

	foundBytes, err := block.Find([]byte{0x00, 0x01})
	assert.NoError(t, err, "unexpected error creating reading req")

	outReq := &friggpb.PushRequest{}
	err = proto.Unmarshal(foundBytes, outReq)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(req, outReq))
}

func TestIterator(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(tempDir)
	assert.NoError(t, err, "unexpected error creating temp dir")

	wal, err := newWAL(&walConfig{
		filepath:        tempDir,
		indexDownsample: 2,
	})
	assert.NoError(t, err, "unexpected error creating temp wal")

	blockID := uuid.New()

	block, err := wal.NewBlock(blockID, testTenantID)
	assert.NoError(t, err, "unexpected error creating block")

	numMsgs := 10
	reqs := make([]*friggpb.PushRequest, 0, numMsgs)
	for i := 0; i < numMsgs; i++ {
		req := test.MakeRequest(rand.Int()%1000, []byte{})
		reqs = append(reqs, req)
		bReq, err := proto.Marshal(req)
		assert.NoError(t, err)
		err = block.Write([]byte{}, bReq)
		assert.NoError(t, err, "unexpected error writing req")
	}

	i := 0
	err = block.(*headBlock).Iterator(func(id ID, msg []byte) (bool, error) {
		req := &friggpb.PushRequest{}
		err = proto.Unmarshal(msg, req)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(req, reqs[i]))
		i++

		return true, nil
	})

	assert.NoError(t, err, "unexpected error iterating")
	assert.Equal(t, numMsgs, i)
}

func TestCompleteBlock(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(tempDir)
	assert.NoError(t, err, "unexpected error creating temp dir")

	indexDownsample := 13
	wal, err := newWAL(&walConfig{
		filepath:        tempDir,
		indexDownsample: indexDownsample,
		bloomFP:         .01,
	})
	assert.NoError(t, err, "unexpected error creating temp wal")

	blockID := uuid.New()

	block, err := wal.NewBlock(blockID, testTenantID)
	assert.NoError(t, err, "unexpected error creating block")

	numMsgs := 100
	reqs := make([]*friggpb.PushRequest, 0, numMsgs)
	ids := make([][]byte, 0, numMsgs)
	for i := 0; i < numMsgs; i++ {
		id := make([]byte, 16)
		rand.Read(id)
		req := test.MakeRequest(rand.Int()%1000, id)
		reqs = append(reqs, req)
		ids = append(ids, id)
		bReq, err := proto.Marshal(req)
		assert.NoError(t, err)
		err = block.Write(id, bReq)
		assert.NoError(t, err, "unexpected error writing req")
	}

	assert.True(t, bytes.Equal(block.(*headBlock).records[0].ID, block.(*headBlock).meta.MinID))
	assert.True(t, bytes.Equal(block.(*headBlock).records[numMsgs-1].ID, block.(*headBlock).meta.MaxID))

	complete, err := block.Complete(wal)
	assert.NoError(t, err, "unexpected error completing block")
	// test downsample config
	assert.Equal(t, numMsgs/indexDownsample+1, len(complete.(*headBlock).records))

	assert.True(t, bytes.Equal(complete.(*headBlock).meta.MinID, block.(*headBlock).meta.MinID))
	assert.True(t, bytes.Equal(complete.(*headBlock).meta.MaxID, block.(*headBlock).meta.MaxID))

	for i, id := range ids {
		out := &friggpb.PushRequest{}
		foundBytes, err := complete.Find(id)
		assert.NoError(t, err)

		err = proto.Unmarshal(foundBytes, out)
		assert.NoError(t, err)

		assert.True(t, proto.Equal(out, reqs[i]))
		assert.True(t, complete.bloomFilter().Has(farm.Fingerprint64(id)))
	}

	// confirm order
	var prev *Record
	for _, r := range complete.(*headBlock).records {
		if prev != nil {
			assert.Greater(t, r.Start, prev.Start)
		}

		prev = r
	}
}

func TestWorkDir(t *testing.T) {
	tempDir, err := ioutil.TempDir("/tmp", "")
	defer os.RemoveAll(tempDir)
	assert.NoError(t, err, "unexpected error creating temp dir")

	err = os.MkdirAll(path.Join(tempDir, workDir), os.ModePerm)
	assert.NoError(t, err, "unexpected error creating workdir")

	_, err = os.Create(path.Join(tempDir, workDir, "testfile"))
	assert.NoError(t, err, "unexpected error creating testfile")

	_, err = newWAL(&walConfig{
		filepath:        tempDir,
		indexDownsample: 2,
	})
	assert.NoError(t, err, "unexpected error creating temp wal")

	_, err = os.Stat(path.Join(tempDir, workDir))
	assert.NoError(t, err, "work folder should exist")

	files, err := ioutil.ReadDir(path.Join(tempDir, workDir))
	assert.NoError(t, err, "unexpected reading work dir")

	assert.Len(t, files, 0, "work dir should be empty")
}
