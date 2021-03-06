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

package column

import (
	"github.com/apache/arrow/go/v9/arrow"
	"github.com/apache/arrow/go/v9/arrow/array"
	"github.com/apache/arrow/go/v9/arrow/memory"

	"otel-arrow-adapter/pkg/air/rfield"
)

// I8Column is a column of int8 data.
type I8Column struct {
	// name of the column.
	name string
	// data of the column.
	data []*int8
}

// U8Column is a column of int8 data.
type I16Column struct {
	// name of the column.
	name string
	// data of the column.
	data []*int16
}

// I32Column is a column of int32 data.
type I32Column struct {
	// name of the column.
	name string
	// data of the column.
	data []*int32
}

// I64Column is a column of int64 data.
type I64Column struct {
	// name of the column.
	name string
	// data of the column.
	data []*int64
}

// MakeI8Column creates a new I8 column.
func MakeI8Column(name string) I8Column {
	return I8Column{
		name: name,
		data: []*int8{},
	}
}

// MakeI16Column creates a new I16 column.
func MakeI16Column(name string) I16Column {
	return I16Column{
		name: name,
		data: []*int16{},
	}
}

// MakeI32Column creates a new I32 column.
func MakeI32Column(name string) I32Column {
	return I32Column{
		name: name,
		data: []*int32{},
	}
}

// MakeI64Column creates a new I64 column.
func MakeI64Column(name string) I64Column {
	return I64Column{
		name: name,
		data: []*int64{},
	}
}

// Name returns the name of the column.
func (c *I8Column) Name() string {
	return c.name
}

func (c *I8Column) Type() arrow.DataType {
	return arrow.PrimitiveTypes.Int8
}

// Push adds a new value to the column.
func (c *I8Column) Push(data *int8) {
	c.data = append(c.data, data)
}

// Len returns the number of values in the column.
func (c *I8Column) Len() int {
	return len(c.data)
}

// NewArrowField creates a I8 schema field.
func (c *I8Column) NewArrowField() *arrow.Field {
	return &arrow.Field{Name: c.name, Type: arrow.PrimitiveTypes.Int8}
}

// NewArray creates and initializes a new Arrow Array for the column.
func (c *I8Column) NewArray(allocator *memory.GoAllocator) arrow.Array {
	builder := array.NewInt8Builder(allocator)
	builder.Reserve(len(c.data))
	for _, v := range c.data {
		if v == nil {
			builder.AppendNull()
		} else {
			builder.UnsafeAppend(*v)
		}
	}
	c.Clear()
	return builder.NewArray()
}

// Clear clears the int8 data in the column but keep the original memory buffer allocated.
func (c *I8Column) Clear() {
	c.data = c.data[:0]
}

func (c *I8Column) PushFromValues(_ *rfield.FieldPath, data []rfield.Value) {
	for _, value := range data {
		v, err := value.AsI8()
		if err != nil {
			panic(err)
		}
		c.data = append(c.data, v)
	}
}

// Name returns the name of the column.
func (c *I16Column) Name() string {
	return c.name
}

func (c *I16Column) Type() arrow.DataType {
	return arrow.PrimitiveTypes.Int16
}

// Push adds a new value to the column.
func (c *I16Column) Push(data *int16) {
	c.data = append(c.data, data)
}

// Len returns the number of values in the column.
func (c *I16Column) Len() int {
	return len(c.data)
}

// NewArrowField creates a I16 schema field.
func (c *I16Column) NewArrowField() *arrow.Field {
	return &arrow.Field{Name: c.name, Type: arrow.PrimitiveTypes.Int16}
}

// NewArray creates and initializes a new Arrow Array for the column.
func (c *I16Column) NewArray(allocator *memory.GoAllocator) arrow.Array {
	builder := array.NewInt16Builder(allocator)
	builder.Reserve(len(c.data))
	for _, v := range c.data {
		if v == nil {
			builder.AppendNull()
		} else {
			builder.UnsafeAppend(*v)
		}
	}
	c.Clear()
	return builder.NewArray()
}

// Clear clears the int16 data in the column but keep the original memory buffer allocated.
func (c *I16Column) Clear() {
	c.data = c.data[:0]
}

func (c *I16Column) PushFromValues(_ *rfield.FieldPath, data []rfield.Value) {
	for _, value := range data {
		v, err := value.AsI16()
		if err != nil {
			panic(err)
		}
		c.data = append(c.data, v)
	}
}

// Name returns the name of the column.
func (c *I32Column) Name() string {
	return c.name
}

func (c *I32Column) Type() arrow.DataType {
	return arrow.PrimitiveTypes.Int32
}

// Push adds a new value to the column.
func (c *I32Column) Push(data *int32) {
	c.data = append(c.data, data)
}

// Len returns the number of values in the column.
func (c *I32Column) Len() int {
	return len(c.data)
}

// Clear clears the int32 data in the column but keep the original memory buffer allocated.
func (c *I32Column) Clear() {
	c.data = c.data[:0]
}

func (c *I32Column) PushFromValues(_ *rfield.FieldPath, data []rfield.Value) {
	for _, value := range data {
		v, err := value.AsI32()
		if err != nil {
			panic(err)
		}
		c.data = append(c.data, v)
	}
}

// NewArrowField creates a I32 schema field.
func (c *I32Column) NewArrowField() *arrow.Field {
	return &arrow.Field{Name: c.name, Type: arrow.PrimitiveTypes.Int32}
}

// NewArray creates and initializes a new Arrow Array for the column.
func (c *I32Column) NewArray(allocator *memory.GoAllocator) arrow.Array {
	builder := array.NewInt32Builder(allocator)
	builder.Reserve(len(c.data))
	for _, v := range c.data {
		if v == nil {
			builder.AppendNull()
		} else {
			builder.UnsafeAppend(*v)
		}
	}
	c.Clear()
	return builder.NewArray()
}

// Name returns the name of the column.
func (c *I64Column) Name() string {
	return c.name
}

// Push adds a new value to the column.
func (c *I64Column) Push(data *int64) {
	c.data = append(c.data, data)
}

func (c *I64Column) PushFromValues(_ *rfield.FieldPath, data []rfield.Value) {
	for _, value := range data {
		i64, err := value.AsI64()
		if err != nil {
			panic(err)
		}
		c.data = append(c.data, i64)
	}
}

// Len returns the number of values in the column.
func (c *I64Column) Len() int {
	return len(c.data)
}

// Clear clears the int64 data in the column but keep the original memory buffer allocated.
func (c *I64Column) Clear() {
	c.data = c.data[:0]
}

// NewArrowField creates a I64 schema field.
func (c *I64Column) NewArrowField() *arrow.Field {
	return &arrow.Field{Name: c.name, Type: arrow.PrimitiveTypes.Int64}
}

func (c *I64Column) Type() arrow.DataType {
	return arrow.PrimitiveTypes.Int64
}

func (c *I64Column) Build(allocator *memory.GoAllocator) (*arrow.Field, arrow.Array, error) {
	return c.NewArrowField(), c.NewArray(allocator), nil
}

// NewArray creates and initializes a new Arrow Array for the column.
func (c *I64Column) NewArray(allocator *memory.GoAllocator) arrow.Array {
	builder := array.NewInt64Builder(allocator)
	builder.Reserve(len(c.data))
	for _, v := range c.data {
		if v == nil {
			builder.AppendNull()
		} else {
			builder.UnsafeAppend(*v)
		}
	}
	c.Clear()
	return builder.NewArray()
}
