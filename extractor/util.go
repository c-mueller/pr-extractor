package extractor

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
)

func getPullRequestId(evt PREvent) string {
	prIdInput := fmt.Sprintf("%s#%d", evt.GetRepoName(), evt.GetPullRequestNumber())
	//h := sha256.New()
	//h.Write([]byte(prIdInput))
	//return hex.EncodeToString(h.Sum([]byte{}))
	return prIdInput
}

func getEventId(evt PREvent) string {
	evtIdInput := fmt.Sprintf("%s-%s", evt.GetEventTimestamp().String(), evt.GetPullRequestURL())
	h := sha256.New()
	h.Write([]byte(evtIdInput))
	resultHash := hex.EncodeToString(h.Sum([]byte{}))
	return resultHash
}

func GzipCompress(data []byte) ([]byte, error) {
	var outputBuffer bytes.Buffer
	compressionWriter := gzip.NewWriter(&outputBuffer)
	_, err := compressionWriter.Write(data)
	if err != nil {
		return nil, err
	}
	compressionWriter.Close()

	return outputBuffer.Bytes(), nil
}

func GzipExtract(data []byte) ([]byte, error) {
	inputBuffer := bytes.NewReader(data)
	compressionReader, err := gzip.NewReader(inputBuffer)
	if err != nil {
		return nil, err
	}

	defer compressionReader.Close()

	return ioutil.ReadAll(compressionReader)
}
