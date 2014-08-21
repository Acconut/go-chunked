package chunked

import (
    "testing"
)

func TestBlockConversion(t *testing.T) {
    
    b := Block{
        Type: 1,
        Length: 11,
        Data: []byte("hello world"),
        NextBlock: -1,
    }
    
    bytes := b.Bytes(32)
    
    block2, err := blockFromBytes(bytes)
    if err != nil {
        t.Fatal(err)
    }
    
    if block2.Type != 1 {
        t.Fatal("wrong type")
    }
    
    if block2.Length != 11 {
        t.Fatal("wrong length")
    }
    
    if string(block2.Data) != "hello world" {
        t.Fatal("wrong data")
    }
    
    if block2.NextBlock != -1 {
        t.Fatal("wrong next block")
    }
    
}