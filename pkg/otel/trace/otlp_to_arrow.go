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

package trace

import (
	"github.com/apache/arrow/go/v9/arrow"

	coltracepb "otel-arrow-adapter/api/go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v1 "otel-arrow-adapter/api/go.opentelemetry.io/proto/otlp/trace/v1"
	"otel-arrow-adapter/pkg/air"
	"otel-arrow-adapter/pkg/air/rfield"
	"otel-arrow-adapter/pkg/otel/common"
	"otel-arrow-adapter/pkg/otel/constants"
)

// OtlpTraceToArrowRecords converts an OTLP trace to one or more Arrow records.
func OtlpTraceToArrowRecords(rr *air.RecordRepository, request *coltracepb.ExportTraceServiceRequest) ([]arrow.Record, error) {
	for _, resourceSpans := range request.ResourceSpans {
		for _, scopeSpans := range resourceSpans.ScopeSpans {
			for _, span := range scopeSpans.Spans {
				record := air.NewRecord()

				if span.StartTimeUnixNano > 0 {
					record.U64Field(constants.START_TIME_UNIX_NANO, span.StartTimeUnixNano)
				}
				if span.EndTimeUnixNano > 0 {
					record.U64Field(constants.END_TIME_UNIX_NANO, span.EndTimeUnixNano)
				}
				common.AddResource(record, resourceSpans.Resource)
				common.AddScope(record, constants.SCOPE_SPANS, scopeSpans.Scope)

				if span.TraceId != nil && len(span.TraceId) > 0 {
					record.BinaryField(constants.TRACE_ID, span.TraceId)
				}
				if span.SpanId != nil && len(span.SpanId) > 0 {
					record.BinaryField(constants.SPAN_ID, span.SpanId)
				}
				if len(span.TraceState) > 0 {
					record.StringField(constants.TRACE_STATE, span.TraceState)
				}
				if span.ParentSpanId != nil && len(span.ParentSpanId) > 0 {
					record.BinaryField(constants.PARENT_SPAN_ID, span.SpanId)
				}
				if len(span.Name) > 0 {
					record.StringField(constants.NAME, span.Name)
				}
				record.I32Field(constants.KIND, int32(span.Kind))
				attributes := common.NewAttributes(span.Attributes)
				if attributes != nil {
					record.AddField(attributes)
				}

				if span.DroppedAttributesCount > 0 {
					record.U32Field(constants.DROPPED_ATTRIBUTES_COUNT, uint32(span.DroppedAttributesCount))
				}

				// Events
				AddEvents(record, span.Events)
				if span.DroppedEventsCount > 0 {
					record.U32Field(constants.DROPPED_EVENTS_COUNT, uint32(span.DroppedEventsCount))
				}

				// Links
				AddLinks(record, span.Links)
				if span.DroppedLinksCount > 0 {
					record.U32Field(constants.DROPPED_LINKS_COUNT, uint32(span.DroppedLinksCount))
				}

				// Status
				if span.Status != nil {
					record.I32Field(constants.STATUS, int32(span.Status.Code))
					record.StringField(constants.STATUS_MESSAGE, span.Status.Message)
				}

				rr.AddRecord(record)
			}
		}
	}

	logsRecords, err := rr.Build()
	if err != nil {
		return nil, err
	}

	result := make([]arrow.Record, 0, len(logsRecords))
	for _, record := range logsRecords {
		result = append(result, record)
	}

	return result, nil
}

func AddEvents(record *air.Record, events []*v1.Span_Event) {
	if events == nil {
		return
	}

	convertedEvents := make([]rfield.Value, 0, len(events))

	for _, event := range events {
		fields := make([]*rfield.Field, 0, 4)

		if event.TimeUnixNano > 0 {
			fields = append(fields, rfield.NewU64Field(constants.TIME_UNIX_NANO, event.TimeUnixNano))
		}
		if len(event.Name) > 0 {
			fields = append(fields, rfield.NewStringField(constants.NAME, event.Name))
		}
		if event.Attributes != nil {
			attributes := common.NewAttributes(event.Attributes)
			if attributes != nil {
				fields = append(fields, attributes)
			}
		}
		if event.DroppedAttributesCount > 0 {
			fields = append(fields, rfield.NewU32Field(constants.DROPPED_ATTRIBUTES_COUNT, uint32(event.DroppedAttributesCount)))
		}
		convertedEvents = append(convertedEvents, &rfield.Struct{
			Fields: fields,
		})
	}
	record.ListField(constants.SPAN_EVENTS, rfield.List{
		Values: convertedEvents,
	})
}

func AddLinks(record *air.Record, links []*v1.Span_Link) {
	if links == nil {
		return
	}

	convertedLinks := make([]rfield.Value, 0, len(links))

	for _, link := range links {
		fields := make([]*rfield.Field, 0, 4)

		if link.TraceId != nil && len(link.TraceId) > 0 {
			fields = append(fields, rfield.NewBinaryField(constants.TRACE_ID, link.TraceId))
		}
		if link.SpanId != nil && len(link.SpanId) > 0 {
			fields = append(fields, rfield.NewBinaryField(constants.SPAN_ID, link.SpanId))
		}
		if len(link.TraceState) > 0 {
			fields = append(fields, rfield.NewStringField(constants.TRACE_STATE, link.TraceState))
		}
		if link.Attributes != nil {
			attributes := common.NewAttributes(link.Attributes)
			if attributes != nil {
				fields = append(fields, attributes)
			}
		}
		if link.DroppedAttributesCount > 0 {
			fields = append(fields, rfield.NewU32Field(constants.DROPPED_ATTRIBUTES_COUNT, uint32(link.DroppedAttributesCount)))
		}
		convertedLinks = append(convertedLinks, &rfield.Struct{
			Fields: fields,
		})
	}
	record.ListField(constants.SPAN_LINKS, rfield.List{
		Values: convertedLinks,
	})
}
