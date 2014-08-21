package chunked

import (
    "bytes"
    "encoding/binary"
)

type Block struct {
    Type int8
    Data []byte
    Length uint64
    NextBlock int64
}

func (b *Block) Bytes(blocksize uint) ([]byte) {

    buf := new(bytes.Buffer)

    var data = []interface{}{
        b.Type,
        b.NextBlock,
        b.Length,
        b.Data,
    }

    for _, v := range data {
        _ = binary.Write(buf, binary.BigEndian, v)
    }

    buf.Write(make([]byte, int(blocksize) - len(b.Data)))
    
    return buf.Bytes()

}

func blockFromBytes(buf []byte) (*Block, error) {

    var block Block
    reader := bytes.NewReader(buf)

    // Read type from buffer
    if err := binary.Read(reader, binary.BigEndian, &block.Type); err != nil {
        return nil, err
    }

    // Read next block id
    if err := binary.Read(reader, binary.BigEndian, &block.NextBlock); err != nil {
        return nil, err
    }

    // Read length
    if err := binary.Read(reader, binary.BigEndian, &block.Length); err != nil {
        return nil, err
    }

    // Read data
    block.Data = buf[1+8+8:1+8+8 + block.Length]

    return &block, nil

}