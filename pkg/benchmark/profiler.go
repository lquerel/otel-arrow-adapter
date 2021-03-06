package benchmark

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
)

type Profiler struct {
	batchSizes []int
	benchmarks []*ProfilerResult
}

func NewProfiler(batchSizes []int) *Profiler {
	return &Profiler{
		batchSizes: batchSizes,
		benchmarks: []*ProfilerResult{},
	}
}

func (p *Profiler) Profile(profileable ProfileableSystem, maxIter uint64) error {
	tags := strings.Join(profileable.Tags()[:], "+")

	p.benchmarks = append(p.benchmarks, &ProfilerResult{
		benchName: profileable.Name(),
		summaries: []BatchSummary{},
		tags:      tags,
	})

	for _, batchSize := range p.batchSizes {
		fmt.Printf("Profiling '%s' (compression=%v, batch-size=%d)\n", profileable.Name(), profileable.CompressionAlgorithm(), batchSize)

		uncompressedSize := NewMetric()
		compressedSize := NewMetric()
		batchCeation := NewMetric()
		processing := NewMetric()
		serialization := NewMetric()
		deserialization := NewMetric()
		compression := NewMetric()
		decompression := NewMetric()
		totalTime := NewMetric()
		processingResults := []string{}

		profileable.InitBatchSize(batchSize)

		for _i := uint64(0); _i < maxIter; _i++ {
			maxBatchCount := uint64(math.Ceil(float64(profileable.DatasetSize()) / float64(batchSize)))
			startAt := 0
			for _j := uint64(0); _j < maxBatchCount; _j++ {
				correctedBatchSize := min(profileable.DatasetSize()-startAt, batchSize)
				profileable.PrepareBatch(startAt, correctedBatchSize)

				start := time.Now()

				// Batch creation
				profileable.CreateBatch(startAt, correctedBatchSize)
				afterBatchCreation := time.Now()

				// Processing
				result := profileable.Process()
				afterProcessing := time.Now()
				processingResults = append(processingResults, result)

				// Serialization
				buffers, err := profileable.Serialize()
				if err != nil {
					return err
				}
				afterSerialization := time.Now()
				uncompressedSizeBytes := 0
				for _, buffer := range buffers {
					uncompressedSizeBytes += len(buffer)
				}
				uncompressedSize.Record(float64(uncompressedSizeBytes))

				// Compression
				var compressedBuffers [][]byte
				for _, buffer := range buffers {
					compressedBuffer, err := Compress(profileable.CompressionAlgorithm(), buffer)
					if err != nil {
						return err
					}
					compressedBuffers = append(compressedBuffers, compressedBuffer)
				}
				afterCompression := time.Now()
				compressedSizeBytes := 0
				for _, buffer := range compressedBuffers {
					compressedSizeBytes += len(buffer)
				}
				compressedSize.Record(float64(compressedSizeBytes))

				// Decompression
				var uncompressedBuffers [][]byte
				for _, buffer := range compressedBuffers {
					uncompressedBuffer, err := Decompress(profileable.CompressionAlgorithm(), buffer)
					if err != nil {
						return err
					}
					uncompressedBuffers = append(uncompressedBuffers, uncompressedBuffer)
				}
				afterDecompression := time.Now()
				if !bytesEqual(buffers, uncompressedBuffers) {
					return fmt.Errorf("Buffers are not equal after decompression")
				}

				// Deserialization
				profileable.Deserialize(buffers)
				afterDeserialization := time.Now()
				profileable.Clear()

				batchCeation.Record(float64(afterBatchCreation.Sub(start).Seconds()))
				processing.Record(float64(afterProcessing.Sub(afterBatchCreation).Seconds()))
				serialization.Record(float64(afterSerialization.Sub(afterProcessing).Seconds()))
				compression.Record(float64(afterCompression.Sub(afterSerialization).Seconds()))
				decompression.Record(float64(afterDecompression.Sub(afterCompression).Seconds()))
				deserialization.Record(float64(afterDeserialization.Sub(afterDecompression).Seconds()))

				totalTime.Record(
					float64(afterBatchCreation.Sub(start).Seconds()) +
						float64(afterProcessing.Sub(afterBatchCreation).Seconds()) +
						float64(afterSerialization.Sub(afterProcessing).Seconds()) +
						float64(afterCompression.Sub(afterSerialization).Seconds()) +
						float64(afterDecompression.Sub(afterCompression).Seconds()) +
						float64(afterDeserialization.Sub(afterDecompression).Seconds()),
				)
			}
		}

		profileable.ShowStats()
		currentBenchmark := p.benchmarks[len(p.benchmarks)-1]
		currentBenchmark.summaries = append(currentBenchmark.summaries, BatchSummary{
			batchSize:            batchSize,
			uncompressedSizeByte: uncompressedSize.ComputeSummary(),
			compressedSizeByte:   compressedSize.ComputeSummary(),
			batchCreationSec:     batchCeation.ComputeSummary(),
			processingSec:        processing.ComputeSummary(),
			serializationSec:     serialization.ComputeSummary(),
			deserializationSec:   deserialization.ComputeSummary(),
			compressionSec:       compression.ComputeSummary(),
			decompressionSec:     decompression.ComputeSummary(),
			totalTimeSec:         totalTime.ComputeSummary(),
			processingResults:    processingResults,
		})
	}
	return nil
}

func (p *Profiler) CheckProcessingResults() {
	for batchIdx := range p.batchSizes {
		if len(p.benchmarks) == 0 {
			continue
		}

		var refProcessingResults []string
		for _, benchmark := range p.benchmarks {
			if len(refProcessingResults) == 0 {
				refProcessingResults = benchmark.summaries[batchIdx].processingResults
			} else {
				if !stringsEqual(refProcessingResults, benchmark.summaries[batchIdx].processingResults) {
					panic("Processing results are not equal")
				}
			}
		}
	}
}

func (p *Profiler) PrintResults() {
	p.PrintStepsTiming()
	p.PrintCompressionRatio()
}

func (p *Profiler) PrintStepsTiming() {
	header := []string{"Steps"}
	for _, benchmark := range p.benchmarks {
		header = append(header, fmt.Sprintf("%s %s - p99", benchmark.benchName, benchmark.tags))
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(header)

	values := make(map[string]*Summary)
	for _, result := range p.benchmarks {
		for _, summary := range result.summaries {
			key := fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "batch_creation_sec")
			values[key] = summary.batchCreationSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "processing_sec")
			values[key] = summary.processingSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "serialization_sec")
			values[key] = summary.serializationSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "compression_sec")
			values[key] = summary.compressionSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "decompression_sec")
			values[key] = summary.decompressionSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "deserialization_sec")
			values[key] = summary.deserializationSec
			key = fmt.Sprintf("%s:%s:%d:%s", result.benchName, result.tags, summary.batchSize, "total_time_sec")
			values[key] = summary.totalTimeSec
		}
	}

	transform := func(value float64) float64 { return value * 1000.0 }
	p.AddSection("Batch creation (ms)", "batch_creation_sec", table, values, transform)
	// p.AddSection("Batch processing (ms)", "processing_sec", table, values, transform)
	p.AddSection("Serialization (ms)", "serialization_sec", table, values, transform)
	p.AddSection("Compression (ms)", "compression_sec", table, values, transform)
	p.AddSection("Decompression (ms)", "decompression_sec", table, values, transform)
	p.AddSection("Deserialisation (ms)", "deserialization_sec", table, values, transform)
	p.AddSection("Total time (ms)", "total_time_sec", table, values, transform)

	table.Render()
}

func (p *Profiler) PrintCompressionRatio() {

}

func (p *Profiler) AddSection(_label string, _step string, _table *tablewriter.Table, _values map[string]*Summary, _transform func(float64) float64) {
	// ToDo add section
}

func (p *Profiler) ExportMetricsTimesCSV(filePrefix string) {
	file, err := os.OpenFile(fmt.Sprintf("%s_times.csv", filePrefix), os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	dataWriter := bufio.NewWriter(file)

	_, err = dataWriter.WriteString("batch_size,duration_ms,protocol,step\n")
	if err != nil {
		panic(fmt.Sprintf("failed writing to file: %s", err))
	}

	for batchIdx, batchSize := range p.batchSizes {
		if len(p.benchmarks) == 0 {
			continue
		}

		for _, result := range p.benchmarks {
			batchCreationMs := result.summaries[batchIdx].batchCreationSec.P99
			serializationMs := result.summaries[batchIdx].serializationSec.P99
			compressionMs := result.summaries[batchIdx].compressionSec.P99
			decompressionMs := result.summaries[batchIdx].decompressionSec.P99
			deserializationMs := result.summaries[batchIdx].deserializationSec.P99

			_, err = dataWriter.WriteString(fmt.Sprintf("%d,%f,%s [%s],0_Batch_creation\n", batchSize, batchCreationMs, result.benchName, result.tags))
			if err != nil {
				panic(fmt.Sprintf("failed writing to file: %s", err))
			}
			_, err = dataWriter.WriteString(fmt.Sprintf("%d,%f,%s [%s],1_Serialization\n", batchSize, serializationMs, result.benchName, result.tags))
			if err != nil {
				panic(fmt.Sprintf("failed writing to file: %s", err))
			}
			_, err = dataWriter.WriteString(fmt.Sprintf("%d,%f,%s [%s],2_Compression\n", batchSize, compressionMs, result.benchName, result.tags))
			if err != nil {
				panic(fmt.Sprintf("failed writing to file: %s", err))
			}
			_, err = dataWriter.WriteString(fmt.Sprintf("%d,%f,%s [%s],3_Decompression\n", batchSize, decompressionMs, result.benchName, result.tags))
			if err != nil {
				panic(fmt.Sprintf("failed writing to file: %s", err))
			}
			_, err = dataWriter.WriteString(fmt.Sprintf("%d,%f,%s [%s],4_Deserialization\n", batchSize, deserializationMs, result.benchName, result.tags))
			if err != nil {
				panic(fmt.Sprintf("failed writing to file: %s", err))
			}
		}
	}

	err = dataWriter.Flush()
	if err != nil {
		panic(fmt.Sprintf("failed flushing the file: %s", err))
	}
	err = file.Close()
	if err != nil {
		panic(fmt.Sprintf("failed closing the file: %s", err))
	}
}

func (p *Profiler) ExportMetricsBytesCSV(filePrefix string) {
	file, err := os.OpenFile(fmt.Sprintf("%s_bytes.csv", filePrefix), os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}

	dataWriter := bufio.NewWriter(file)

	_, err = dataWriter.WriteString("batch_size,iteration,compressed_size_byte,uncompressed_size_byte,Protocol\n")
	if err != nil {
		panic(fmt.Sprintf("failed writing to file: %s", err))
	}

	for batchIdx, batchSize := range p.batchSizes {
		if len(p.benchmarks) == 0 {
			continue
		}

		numSamples := len(p.benchmarks[0].summaries[batchIdx].batchCreationSec.Values)
		for sampleIdx := 0; sampleIdx < numSamples; sampleIdx++ {
			for _, result := range p.benchmarks {
				line := fmt.Sprintf("%d,%d", batchSize, sampleIdx)
				compressedSizeByte := result.summaries[batchIdx].compressedSizeByte.Values[sampleIdx]
				uncompressedSizeByte := result.summaries[batchIdx].uncompressedSizeByte.Values[sampleIdx]

				line += fmt.Sprintf(",%f,%f,%s [%s]\n", compressedSizeByte, uncompressedSizeByte, result.benchName, result.tags)
				_, err = dataWriter.WriteString(line)
				if err != nil {
					panic(fmt.Sprintf("failed writing to file: %s", err))
				}
			}
		}
	}

	err = dataWriter.Flush()
	if err != nil {
		panic(fmt.Sprintf("failed flushing the file: %s", err))
	}
	err = file.Close()
	if err != nil {
		panic(fmt.Sprintf("failed closing the file: %s", err))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func bytesEqual(buffers1, buffers2 [][]byte) bool {
	if len(buffers1) != len(buffers2) {
		return false
	}
	for i := range buffers1 {
		if !bytes.Equal(buffers1[i], buffers2[i]) {
			return false
		}
	}
	return true
}

func stringsEqual(buffers1, buffers2 []string) bool {
	if len(buffers1) != len(buffers2) {
		return false
	}
	for i, v1 := range buffers1 {
		if v1 != buffers2[i] {
			return false
		}
	}
	return true
}
