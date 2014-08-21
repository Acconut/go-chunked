package chunked

import (
    "errors"
    "os"
    "path"
    "encoding/json"
    "io/ioutil"
    "regexp"
    "math"
    "bytes"

    "github.com/Acconut/go-chunked/block"
)

type Config struct {

    Blocksize uint
    Rotation uint
    NextBlock uint

}

var DefaultConfig = &Config{
    Blocksize: 100,
    Rotation: 10000,
}

type Database struct {

    dir string
    config Config
    chunks []*os.File

}

var chunkFileNameRe = regexp.MustCompile(`^chunk\.(\d+)$`)

func Create(dir string, config *Config) (*Database, error) {

    // Test whether directory exists already
    _, err := os.Stat(dir)
    if err == nil {
        return nil, errors.New("directory already used")
    }

    // Create directory
    err = os.MkdirAll(dir, os.ModePerm)
    if err != nil {
        return nil, err
    }

    // Create first chunk file
    _, err = os.Create(path.Join(dir, "chunk.0"))
    if err != nil {
        return nil, err
    }

    // Store configuration
    file, err := os.Create(path.Join(dir, "config.json"))
    if err != nil {
        return nil, err
    }

    config.NextBlock = 0
    b, err := json.Marshal(config)
    if err != nil {
        return nil, err
    }

    _, err = file.Write(b)
    if err != nil {
        return nil, err
    }

    file.Close()

    return Open(dir)

}

func Open(dir string) (*Database, error) {

    // Read configuration
    file, err := ioutil.ReadFile(path.Join(dir, "config.json"))
    if err != nil {
        return nil, err
    }

    var config Config
    err = json.Unmarshal(file, &config)

    // Get chunk files
    chunks, err := getChunkFiles(dir)
    if err != nil {
        return nil, err
    }

    db := &Database{
        config: config,
        chunks: chunks,
        dir: dir,
    }

    return db, nil
}

func getChunkFiles(dir string) ([]*os.File, error) {

    // Read directory
    content, err := ioutil.ReadDir(dir)
    if err != nil {
        return nil, err
    }

    files := make([]*os.File, 0)
    // Filter content
    for _, file := range content {
        match := chunkFileNameRe.FindAllStringSubmatch(file.Name(), -1)

        if match != nil && !file.IsDir() {

            f, err := os.OpenFile(path.Join(dir, file.Name()), os.O_RDWR, os.ModePerm)
            if err != nil {
                return nil, err
            }

            files = append(files, f)

        }
    }

    return files, nil
}

func (db *Database) Append(data []byte) (uint, error) {

    // Get blocksize
    bs := db.config.Blocksize

    // Calculate number of needed blocks
    num := uint(math.Ceil(float64(len(data)) / float64(bs)))

    // Get first position
    key := db.getFreeBlockPosition()
    pos := key

    // Specify type
    typ := int8(1)

    // Create blocks
    var i uint
    for i = 0; i < num; i++ {

        // Generate next block id if needed
        nextKey := int64(-1)
        if i != (num - 1) {
            nextKey = int64(db.getFreeBlockPosition())
        }

        startPos := i * bs
        endPos := startPos + bs

        // The last block may be not full
        if i == (num - 1) {
            endPos = uint(len(data))
        }

        dat := data[startPos:endPos]
        block := &block.Block{
            Type: int8(typ),
            Length: uint64(len(dat)),
            Data: dat,
            NextBlock: nextKey,
        }

        err := db.writeBlock(pos, block)

        if err != nil {
            return 0, err
        }

        typ = int8(2)
        pos = uint(nextKey)
    }

    return key, nil

}

func (db *Database) getFreeBlockPosition() uint {

    pos := db.config.NextBlock
    db.config.NextBlock++
    return pos

}

func (db *Database) readBlock(pos uint) (*block.Block, error) {

    // Get blocksize
    bs := db.config.Blocksize

    // Get chunk to read from
    chunkNum := int(math.Floor(float64(pos / bs)))
    chunk := db.chunks[chunkNum]

    // Get offset
    offset := int64((pos % bs) * (1 + 8 + bs + 8))

    // Buffer to read into
    buf := make([]byte, 1 + 8 + bs + 8)
    _, err := chunk.ReadAt(buf, offset)

    if err != nil {
        return nil, err
    }

    block, err := block.FromBytes(buf)
    if err != nil {
        return nil, err
    }

    return block, nil
}

func (db *Database) writeBlock(pos uint, block *block.Block) error {

    // Get blocksize
    bs := db.config.Blocksize

    // Get chunk to read from
    chunkNum := int(math.Floor(float64(pos / bs)))
    chunk := db.chunks[chunkNum]

    // Get offset
    offset := int64((pos % bs) * (1 + 8 + bs + 8))

    _, err := chunk.WriteAt(block.Bytes(bs), offset)
    return err

}

func (db *Database) Read(key uint) ([]byte, error) {

    buf := new(bytes.Buffer)
    pos := key

    for {

        block, err := db.readBlock(pos)
        if err != nil {
            return nil, err
        }
        
        buf.Write(block.Data)
        
        if block.NextBlock != -1 {
            pos = uint(block.NextBlock)
        } else {
            return buf.Bytes(), nil
        }
    }

}

func (db *Database) saveConfig() error {

    file, err := os.OpenFile(path.Join(db.dir, "config.json"), os.O_WRONLY, os.ModePerm)
    if err != nil {
        return err
    }

    b, err := json.Marshal(db.config)
    if err != nil {
        return err
    }

    _, err = file.Write(b)
    if err != nil {
        return err
    }

    return file.Close()

}

func (db *Database) Close() error {
    for _, v := range db.chunks {
        if err := v.Close(); err != nil {
            return err
        }
    }

    if err := db.saveConfig(); err != nil {
        return nil
    }

    return nil
}
