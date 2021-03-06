// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package air

import (
	"github.com/apache/arrow/go/v9/arrow"
	"github.com/apache/arrow/go/v9/arrow/memory"

	config2 "otel-arrow-adapter/pkg/air/config"
)

type RecordRepository struct {
	config *config2.Config

	// A map of SchemaId to RecordBuilder.
	builders map[string]*RecordBuilder

	// ToDo check if release is called properly
	allocator *memory.GoAllocator
}

func NewRecordRepository(config *config2.Config) *RecordRepository {
	return &RecordRepository{
		config:    config,
		builders:  make(map[string]*RecordBuilder),
		allocator: memory.NewGoAllocator(),
	}
}

func (rr *RecordRepository) AddRecord(record *Record) {
	record.Normalize()
	schemaId := record.SchemaId()

	if rb, ok := rr.builders[schemaId]; ok {
		rb.AddRecord(record)
	} else {
		rr.builders[schemaId] = NewRecordBuilderWithRecord(rr.allocator, record, rr.config)
	}
}

// RecordBuilderCount returns the number of non-empty RecordBuilder in the repository.
func (rr *RecordRepository) RecordBuilderCount() int {
	count := 0
	for _, rb := range rr.builders {
		if !rb.IsEmpty() {
			count++
		}
	}
	return count
}

func (rr *RecordRepository) Build() (map[string]arrow.Record, error) {
	recordBatches := make(map[string]arrow.Record)

	for schemaId, builder := range rr.builders {
		if !builder.IsEmpty() {
			record, err := builder.Build(rr.allocator)
			if err != nil {
				return nil, err
			}
			recordBatches[schemaId] = record
		}
	}

	return recordBatches, nil
}

func (rr *RecordRepository) Optimize() {
	for _, rb := range rr.builders {
		rb.Optimize()
	}
}

func (rr *RecordRepository) Metadata() []*RecordBuilderMetadata {
	var metadata []*RecordBuilderMetadata
	for schemaId, rb := range rr.builders {
		if !rb.IsEmpty() {
			metadata = append(metadata, rb.Metadata(schemaId))
		}
	}
	return metadata
}
