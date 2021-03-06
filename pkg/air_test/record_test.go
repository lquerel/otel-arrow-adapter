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

package air_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"otel-arrow-adapter/pkg/air"
	"otel-arrow-adapter/pkg/air/rfield"
)

func TestValue(t *testing.T) {
	t.Parallel()

	record := air.NewRecord()
	record.StringField("b", "b")
	record.StructField("a", rfield.Struct{
		Fields: []*rfield.Field{
			{Name: "e1", Value: &rfield.String{Value: "e1"}},
			{Name: "b1", Value: &rfield.String{Value: "b1"}},
			{Name: "c1", Value: &rfield.Struct{
				Fields: []*rfield.Field{
					{Name: "x", Value: &rfield.String{Value: "x"}},
					{Name: "t", Value: &rfield.String{Value: "t"}},
					{Name: "z", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I64{Value: 1},
							&rfield.I64{Value: 2},
						},
					}},
					{Name: "a", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.Struct{
								Fields: []*rfield.Field{
									{Name: "f2_3_4_2", Value: &rfield.String{Value: "f2_3_4_2"}},
									{Name: "f2_3_4_1", Value: &rfield.String{Value: "f2_3_4_1"}},
								},
							},
						},
					}},
				},
			}},
		},
	})
	record.Normalize()

	v := record.ValueByPath([]int{0, 0}) // field "b"
	if v.(*rfield.String).Value != "b1" {
		t.Errorf("expected the value of field \"a.b1\" to be \"b1\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 1, 0, 0, 0}) // field "a.c1.a.f2_3_4_1"
	if v.(*rfield.String).Value != "f2_3_4_1" {
		t.Errorf("expected the value of field \"a.c1.a.f2_3_4_1\" to be \"f2_3_4_1\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 1, 1}) // field "a.c1.t"
	if v.(*rfield.String).Value != "t" {
		t.Errorf("expected the value of field \"a.c1.t\" to be \"t\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 1, 2}) // field "a.c1.x"
	if v.(*rfield.String).Value != "x" {
		t.Errorf("expected the value of field \"a.c1.x\" to be \"x\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 1, 3, 0}) // field "a.c1.z[0]"
	if v.(*rfield.I64).Value != 1 {
		t.Errorf("expected the value of field \"a.c1.z[0]\" to be \"1\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 1, 3, 1}) // field "a.c1.z[1]"
	if v.(*rfield.I64).Value != 2 {
		t.Errorf("expected the value of field \"a.c1.z[1]\" to be \"2\", got %v", v)
	}

	v = record.ValueByPath([]int{0, 2}) // field "a.e1"
	if v.(*rfield.String).Value != "e1" {
		t.Errorf("expected the value of field \"a.e1\" to be \"e1\", got %v", v)
	}

	v = record.ValueByPath([]int{1}) // field "b"
	if v.(*rfield.String).Value != "b" {
		t.Errorf("expected the value of field \"b\" to be \"b\", got %v", v)
	}
}

func TestCompare(t *testing.T) {
	t.Parallel()

	record1 := GenComplexRecord(1)
	record1.Normalize()

	record2 := GenComplexRecord(2)
	record2.Normalize()

	// Compare the two records based on the field "b".
	sortBy := [][]int{
		{1}, // field "b"
	}
	result := record1.Compare(record2, sortBy)
	if result != 0 {
		t.Errorf("expected the comparison of record1 and record2 to be 0, got %v", result)
	}
	result = record2.Compare(record1, sortBy)
	if result != 0 {
		t.Errorf("expected the comparison of record1 and record2 to be 0, got %v", result)
	}

	// Compare the two records based on the fields "b" and "ts".
	sortBy = [][]int{
		{1}, // field "b"
		{3}, // field "ts"
	}
	result = record1.Compare(record2, sortBy)
	if result != -1 {
		t.Errorf("expected the comparison of record1 and record2 to be -1, got %v", result)
	}
	result = record2.Compare(record1, sortBy)
	if result != 1 {
		t.Errorf("expected the comparison of record2 and record1 to be 1, got %v", result)
	}
}

func TestRecordNormalize(t *testing.T) {
	t.Parallel()

	record := air.NewRecord()
	record.StringField("b", "")
	record.StructField("a", rfield.Struct{
		Fields: []*rfield.Field{
			{Name: "e", Value: &rfield.String{Value: ""}},
			{Name: "b", Value: &rfield.String{Value: ""}},
			{Name: "c", Value: &rfield.Struct{
				Fields: []*rfield.Field{
					{Name: "x", Value: &rfield.String{Value: ""}},
					{Name: "t", Value: &rfield.String{Value: ""}},
					{Name: "z", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I64{Value: 1},
							&rfield.I64{Value: 2},
						},
					}},
					{Name: "a", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.Struct{
								Fields: []*rfield.Field{
									{Name: "f2_3_4_2", Value: &rfield.String{Value: "f2_3_4_2"}},
									{Name: "f2_3_4_1", Value: &rfield.String{Value: "f2_3_4_1"}},
								},
							},
						},
					}},
				},
			}},
		},
	})
	record.Normalize()

	expected_record := air.NewRecord()
	expected_record.StructField("a", rfield.Struct{
		Fields: []*rfield.Field{
			{Name: "b", Value: &rfield.String{Value: ""}},
			{Name: "c", Value: &rfield.Struct{
				Fields: []*rfield.Field{
					{Name: "a", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.Struct{
								Fields: []*rfield.Field{
									{Name: "f2_3_4_1", Value: &rfield.String{Value: "f2_3_4_1"}},
									{Name: "f2_3_4_2", Value: &rfield.String{Value: "f2_3_4_2"}},
								},
							},
						},
					}},
					{Name: "t", Value: &rfield.String{Value: ""}},
					{Name: "x", Value: &rfield.String{Value: ""}},
					{Name: "z", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I64{Value: 1},
							&rfield.I64{Value: 2},
						},
					}},
				},
			}},
			{Name: "e", Value: &rfield.String{Value: ""}},
		},
	})
	expected_record.StringField("b", "")

	if !cmp.Equal(record, expected_record, cmp.AllowUnexported(air.Record{}, rfield.Struct{}, rfield.List{})) {
		t.Errorf("Expected: %+v\nGot: %+v", expected_record, record)
	}
}

func TestRecordSchemaId(t *testing.T) {
	t.Parallel()

	record := air.NewRecord()
	record.StringField("b", "")
	record.StructField("a", rfield.Struct{
		Fields: []*rfield.Field{
			{Name: "e", Value: &rfield.String{Value: ""}},
			{Name: "b", Value: &rfield.String{Value: ""}},
			{Name: "c", Value: &rfield.Struct{
				Fields: []*rfield.Field{
					{Name: "y", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I8{Value: 1},
							&rfield.I64{Value: 2},
							&rfield.String{Value: "true"},
						},
					}},
					{Name: "x", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I8{Value: 1},
							&rfield.I64{Value: 2},
							&rfield.Bool{Value: true},
						},
					}},
					{Name: "t", Value: &rfield.String{Value: ""}},
					{Name: "z", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.I8{Value: 1},
							&rfield.I64{Value: 2},
						},
					}},
					{Name: "a", Value: &rfield.List{
						Values: []rfield.Value{
							&rfield.Struct{
								Fields: []*rfield.Field{
									{Name: "f2_3_4_2", Value: &rfield.I8{Value: 1}},
									{Name: "f2_3_4_1", Value: &rfield.I8{Value: 2}},
								},
							},
							&rfield.Struct{
								Fields: []*rfield.Field{
									{Name: "f2_3_4_3", Value: &rfield.String{Value: "f2_3_4_3"}},
									{Name: "f2_3_4_1", Value: &rfield.String{Value: "f2_3_4_1"}},
								},
							},
						},
					}},
				},
			}},
		},
	})

	record.Normalize()
	id := record.SchemaId()
	expectedSchemaId := "a:{b:Str,c:{a:[{f2_3_4_1:Str,f2_3_4_2:I8,f2_3_4_3:Str}],t:Str,x:[I64],y:[Str],z:[I64]},e:Str},b:Str"
	if id != expectedSchemaId {
		t.Errorf("Expected: %s\nGot: %s", expectedSchemaId, id)
	}
}
