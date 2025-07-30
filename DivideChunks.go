package udm

// DivideChunks divides a file of the given size into a specified number of chunks,
// returning a slice where each element represents the size of a chunk in bytes.
//
// Parameters:
//   - fileSize:   The total size of the file in bytes.
//   - chunkCount: The number of chunks to divide the file into.
//
// Returns:
//   - []int64: A slice containing the size of each chunk in bytes.
//
// Notes:
//   - The function ensures that the sum of all chunk sizes equals fileSize.
//   - Any remainder bytes (if fileSize is not evenly divisible by chunkCount)
//     are added to the second-to-last chunk to avoid underflow in the last chunk.
//
// Example:
//
// chunks := DivideChunks(info.Filesize, 8)
//
//	for i, chunkSize := range chunks {
//		fmt.Printf("Chunk %d: %d \n", i, chunkSize)
//		totalChunkSize += int(chunkSize)
//	}
//
// fmt.Printf("TOtal chunk size got :: %d", totalChunkSize)
func DivideChunks(fileSize int64, chunkCount int) []int64 {
	chunks := make([]int64, chunkCount)

	chunkSize := int64(fileSize / int64(chunkCount)) // Ensure it is a floor value

	totalAllocatedSize := chunkCount * int(chunkSize)
	underFlowSize := fileSize - int64(totalAllocatedSize)

	for i := 0; i < chunkCount; i++ {
		// Include underFLow into the last chunk info Ensure no underflow Exists
		if i == chunkCount-2 {
			chunks[i] = chunkSize + underFlowSize
			continue
		}

		chunks[i] = chunkSize
	}

	return chunks
}
