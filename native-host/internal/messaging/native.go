package messaging

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"native-host/internal/models"
	"os"
)

func ReadMessage(reader io.Reader) (*models.Message, error) {
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	msgBytes := make([]byte, length)
	if _, err := io.ReadFull(reader, msgBytes); err != nil {
		return nil, err
	}

	var message models.Message
	if err := json.Unmarshal(msgBytes, &message); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}

	return &message, nil
}

func SendResponse(response models.Response) error {
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal response: %w", err)
	}

	length := uint32(len(data))
	if err := binary.Write(os.Stdout, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("write length: %w", err)
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return fmt.Errorf("write message: %w", err)
	}

	return nil
}

// ReadAPIRequest reads a length-prefixed JSON APIRequest from reader.
func ReadAPIRequest(reader io.Reader) (*APIRequest, error) {
	var length uint32
	if err := binary.Read(reader, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	msgBytes := make([]byte, length)
	if _, err := io.ReadFull(reader, msgBytes); err != nil {
		return nil, err
	}

	var req APIRequest
	if err := json.Unmarshal(msgBytes, &req); err != nil {
		return nil, fmt.Errorf("unmarshal api request: %w", err)
	}

	return &req, nil
}

// SendAPIResponse writes a length-prefixed JSON APIResponse to stdout.
func SendAPIResponse(resp APIResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal api response: %w", err)
	}

	length := uint32(len(data))
	buf := &bytes.Buffer{}
	if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
		return fmt.Errorf("write length: %w", err)
	}
	if _, err := buf.Write(data); err != nil {
		return fmt.Errorf("write data: %w", err)
	}
	if _, err := os.Stdout.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("write to stdout: %w", err)
	}
	return nil
}
