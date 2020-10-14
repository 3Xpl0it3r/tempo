package encoding

import (
	"bytes"
	"context"
	"io"
	"sort"
)

type dedupingFinder struct {
	ra            io.ReaderAt
	sortedRecords []*Record
	combiner      ObjectCombiner
}

func NewDedupingFinder(sortedRecords []*Record, ra io.ReaderAt, combiner ObjectCombiner) Finder {
	return &dedupingFinder{
		ra:            ra,
		sortedRecords: sortedRecords,
		combiner:      combiner,
	}
}

func (f *dedupingFinder) Find(id ID) ([]byte, error) {
	i := sort.Search(len(f.sortedRecords), func(idx int) bool {
		return bytes.Compare(f.sortedRecords[idx].ID, id) >= 0
	})

	if i < 0 || i >= len(f.sortedRecords) {
		return nil, nil
	}

	var bytesFound []byte

	for {
		record := f.sortedRecords[i]

		bytesOne, err := f.findOne(id, record)
		if err != nil {
			return nil, err
		}

		bytesFound = f.combiner.Combine(bytesFound, bytesOne)

		// we need to check the next record to see if it also matches our id
		i++
		if i >= len(f.sortedRecords) {
			break
		}

		if !bytes.Equal(f.sortedRecords[i].ID, id) {
			break
		}
	}

	return bytesFound, nil
}

func (f *dedupingFinder) findOne(id ID, record *Record) ([]byte, error) {
	buff := make([]byte, record.Length)
	_, err := f.ra.ReadAt(buff, int64(record.Start))
	if err != nil {
		return nil, err
	}

	iter := NewIterator(bytes.NewReader(buff))
	iter, err = NewDedupingIterator(iter, f.combiner)
	if err != nil {
		return nil, err
	}

	for {
		foundID, b, err := iter.Next(context.TODO())
		if foundID == nil {
			break
		}
		if err != nil {
			return nil, err
		}
		if bytes.Equal(foundID, id) {
			return b, nil
		}
	}

	return nil, nil
}
