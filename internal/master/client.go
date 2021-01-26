package master

import (
	"sync"
	"time"

	"github.com/zilliztech/milvus-distributed/internal/proto/commonpb"
	"github.com/zilliztech/milvus-distributed/internal/proto/indexpb"
	writerclient "github.com/zilliztech/milvus-distributed/internal/writenode/client"
)

type WriteNodeClient interface {
	FlushSegment(segmentID UniqueID, collectionID UniqueID, partitionTag string, timestamp Timestamp) error
	DescribeSegment(segmentID UniqueID) (*writerclient.SegmentDescription, error)
	GetInsertBinlogPaths(segmentID UniqueID) (map[UniqueID][]string, error)
}

type MockWriteNodeClient struct {
	segmentID    UniqueID
	flushTime    time.Time
	partitionTag string
	timestamp    Timestamp
	collectionID UniqueID
	lock         sync.RWMutex
}

func (m *MockWriteNodeClient) FlushSegment(segmentID UniqueID, collectionID UniqueID, partitionTag string, timestamp Timestamp) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.flushTime = time.Now()
	m.segmentID = segmentID
	m.collectionID = collectionID
	m.partitionTag = partitionTag
	m.timestamp = timestamp
	return nil
}

func (m *MockWriteNodeClient) DescribeSegment(segmentID UniqueID) (*writerclient.SegmentDescription, error) {
	now := time.Now()
	m.lock.RLock()
	defer m.lock.RUnlock()
	if now.Sub(m.flushTime).Seconds() > 2 {
		return &writerclient.SegmentDescription{
			SegmentID: segmentID,
			IsClosed:  true,
			OpenTime:  0,
			CloseTime: 1,
		}, nil
	}
	return &writerclient.SegmentDescription{
		SegmentID: segmentID,
		IsClosed:  false,
		OpenTime:  0,
		CloseTime: 1,
	}, nil
}

func (m *MockWriteNodeClient) GetInsertBinlogPaths(segmentID UniqueID) (map[UniqueID][]string, error) {
	return map[UniqueID][]string{
		1:   {"/binlog/insert/file_1"},
		100: {"/binlog/insert/file_100"},
	}, nil
}

type BuildIndexClient interface {
	BuildIndex(req *indexpb.BuildIndexRequest) (*indexpb.BuildIndexResponse, error)
	GetIndexStates(req *indexpb.IndexStatesRequest) (*indexpb.IndexStatesResponse, error)
	GetIndexFilePaths(req *indexpb.IndexFilePathsRequest) (*indexpb.IndexFilePathsResponse, error)
}

type MockBuildIndexClient struct {
	buildTime time.Time
}

func (m *MockBuildIndexClient) BuildIndex(req *indexpb.BuildIndexRequest) (*indexpb.BuildIndexResponse, error) {
	m.buildTime = time.Now()
	return &indexpb.BuildIndexResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_SUCCESS,
		},
		IndexID: int64(1),
	}, nil
}

func (m *MockBuildIndexClient) GetIndexStates(req *indexpb.IndexStatesRequest) (*indexpb.IndexStatesResponse, error) {
	now := time.Now()
	ret := &indexpb.IndexStatesResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_SUCCESS,
		},
	}
	var indexStates []*indexpb.IndexInfo
	if now.Sub(m.buildTime).Seconds() > 2 {
		for _, indexID := range req.IndexIDs {
			indexState := &indexpb.IndexInfo{
				State:   commonpb.IndexState_FINISHED,
				IndexID: indexID,
			}
			indexStates = append(indexStates, indexState)
		}
		ret.States = indexStates
		return ret, nil
	}
	for _, indexID := range req.IndexIDs {
		indexState := &indexpb.IndexInfo{
			State:   commonpb.IndexState_INPROGRESS,
			IndexID: indexID,
		}
		indexStates = append(indexStates, indexState)
	}
	ret.States = indexStates
	return ret, nil
}

func (m *MockBuildIndexClient) GetIndexFilePaths(req *indexpb.IndexFilePathsRequest) (*indexpb.IndexFilePathsResponse, error) {
	var filePathInfos []*indexpb.IndexFilePathInfo
	for _, indexID := range req.IndexIDs {
		filePaths := &indexpb.IndexFilePathInfo{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_SUCCESS,
			},
			IndexID:        indexID,
			IndexFilePaths: []string{"/binlog/index/file_1", "/binlog/index/file_2", "/binlog/index/file_3"},
		}
		filePathInfos = append(filePathInfos, filePaths)
	}

	return &indexpb.IndexFilePathsResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_SUCCESS,
		},
		FilePaths: filePathInfos,
	}, nil
}

type LoadIndexClient interface {
	LoadIndex(indexPaths []string, segmentID int64, fieldID int64, fieldName string, indexParams map[string]string) error
}

type MockLoadIndexClient struct {
}

func (m *MockLoadIndexClient) LoadIndex(indexPaths []string, segmentID int64, fieldID int64, fieldName string, indexParams map[string]string) error {
	return nil
}
