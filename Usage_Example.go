package main

import "udm/ufs"

func main() {
	// url := "https://drive.usercontent.google.com/download?id=1d1EBTcLHYQiv93O4nyBBjbK_Wc-2f5qX&export=download&authuser=0&confirm=t&uuid=5cccf6aa-fc97-4bff-89f3-e4339a189778&at=AN8xHoo83pMm2eQ2GwbC6YHA5eK0:1753245767095"
	// info, err := GetServerData(url)
	// if err != nil {
	// 	fmt.Println("Error:", err)
	// 	return
	// }
	// fmt.Printf("Filename: %s\n", info.Filename)
	// fmt.Printf("Size: %d bytes\n", info.Filesize)
	// fmt.Printf("Filetype: %s\n", info.Filetype)
	// fmt.Printf("Accepts Range Requests: %v\n", info.AcceptsRanges)
	// fmt.Printf("Final URL after redirect: %s\n", info.FinalURL)

	// println("\n\n")
	// println("Chunks info::")

	// totalChunkSize := 0

	// chunks := DivideChunks(info.Filesize, 8)
	// for i, chunkSize := range chunks {
	// 	fmt.Printf("Chunk %d: %d \n", i, chunkSize)
	// 	totalChunkSize += int(chunkSize)
	// }

	// fmt.Printf("TOtal chunk size got :: %d", totalChunkSize)

	filename := ufs.GenerateUniqueFilename("go.mod")
	println("Generated unique filename:", filename)

	chunkNames := ufs.GenerateChunkFileNames(filename, 8, "./chunks/")
	for _, chunkName := range chunkNames {
		println("Chunk Name:", chunkName)
	}

}
