//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.
package upside_down

import (
	"bytes"

	"github.com/couchbaselabs/bleve/index"
	"github.com/couchbaselabs/bleve/index/store"
)

type UpsideDownCouchTermFieldReader struct {
	index    *UpsideDownCouch
	iterator store.KVIterator
	count    uint64
	term     []byte
	field    uint16
}

func newUpsideDownCouchTermFieldReader(index *UpsideDownCouch, term []byte, field uint16) (*UpsideDownCouchTermFieldReader, error) {
	tfr := NewTermFrequencyRow(term, field, "", 0, 0)
	it := index.store.Iterator(tfr.Key())

	var count uint64 = 0
	key, val, valid := it.Current()
	if valid {
		if bytes.Equal(key, tfr.Key()) {
			tfr, err := NewTermFrequencyRowKV(key, val)
			if err != nil {
				return nil, err
			}
			count = tfr.freq
		}
	}

	return &UpsideDownCouchTermFieldReader{
		index:    index,
		iterator: it,
		count:    count,
		term:     term,
		field:    field,
	}, nil
}

func (r *UpsideDownCouchTermFieldReader) Count() uint64 {
	return r.count
}

func (r *UpsideDownCouchTermFieldReader) Next() (*index.TermFieldDoc, error) {
	r.iterator.Next()
	key, val, valid := r.iterator.Current()
	if valid {
		testfr := NewTermFrequencyRow(r.term, r.field, "", 0, 0)
		if !bytes.HasPrefix(key, testfr.Key()) {
			// end of the line
			return nil, nil
		}
		tfr, err := NewTermFrequencyRowKV(key, val)
		if err != nil {
			return nil, err
		}
		return &index.TermFieldDoc{
			ID:      string(tfr.doc),
			Freq:    tfr.freq,
			Norm:    float64(tfr.norm),
			Vectors: r.index.termFieldVectorsFromTermVectors(tfr.vectors),
		}, nil
	} else {
		return nil, nil
	}
}

func (r *UpsideDownCouchTermFieldReader) Advance(docId string) (*index.TermFieldDoc, error) {
	tfr := NewTermFrequencyRow(r.term, r.field, docId, 0, 0)
	r.iterator.Seek(tfr.Key())
	key, val, valid := r.iterator.Current()
	if valid {
		testfr := NewTermFrequencyRow(r.term, r.field, "", 0, 0)
		if !bytes.HasPrefix(key, testfr.Key()) {
			// end of the line
			return nil, nil
		}
		tfr, err := NewTermFrequencyRowKV(key, val)
		if err != nil {
			return nil, err
		}
		return &index.TermFieldDoc{
			ID:      string(tfr.doc),
			Freq:    tfr.freq,
			Norm:    float64(tfr.norm),
			Vectors: r.index.termFieldVectorsFromTermVectors(tfr.vectors),
		}, nil
	} else {
		return nil, nil
	}
}

func (r *UpsideDownCouchTermFieldReader) Close() {
	r.iterator.Close()
}
