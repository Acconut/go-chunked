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
)

type Config struct {
    // Maximum size of a block in bytes
    Blocksize uint
    // Maximum number of blocks in a chunk file
    Chunksize uint
    // Internal representation of the next free block (don't use)
    NextBlock uint

}

// Default configuration used to create a new database.
var DefaultConfig = &Config{
    Blocksize: 100,
    Chunksize: 10000,
}

type Database struct {

    dir string
    config Config
    chunks []*os.File

}

var chunkFileNameRe = regexp.MustCompile(`^chunk\.(\d+)$`)

// Creates a new database in the empty directory.
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

// Opens a database from the specified directory.
// No configuration is needed because it's stored on disk.
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

// Append the value to the database.
// The first return value is the key which can be used to read it again.
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
        block := &Block{
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

func (db *Database) readBlock(pos uint) (*Block, error) {

    // Get blocksize and chunksize
    cs := db.config.Chunksize
    bs := db.config.Blocksize

    // Get chunk to read from
    chunkNum := int(math.Floor(float64(pos / cs)))

    // Test if chunk exists
    if chunkNum >= len(db.chunks) {
        return nil, errors.New("key not found")
    }
    chunk := db.chunks[chunkNum]

    // Get offset
    offset := int64((pos % bs) * (1 + 8 + bs + 8))

    // Buffer to read into
    buf := make([]byte, 1 + 8 + bs + 8)
    _, err := chunk.ReadAt(buf, offset)

    if err != nil {
        if err.Error() == "EOF" {
            return nil, errors.New("key not found")
        }
        return nil, err
    }

    block, err := blockFromBytes(buf)
    if err != nil {
        return nil, err
    }

    return block, nil
}

func (db *Database) writeBlock(pos uint, block *Block) error {

    // Get blocksize and chunksize
    cs := db.config.Chunksize
    bs := db.config.Blocksize

    // Get chunk to read from
    chunkNum := int(math.Floor(float64(pos / cs)))
    chunk := db.chunks[chunkNum]

    // Get offset
    offset := int64((pos % bs) * (1 + 8 + bs + 8))

    _, err := chunk.WriteAt(block.Bytes(bs), offset)
    return err

}

// Reads value from the database pointed to by the key.
func (db *Database) Read(key uint) ([]byte, error) {

    buf := new(bytes.Buffer)
    pos := key
    firstBlock := true

    for {

        block, err := db.readBlock(pos)
        if err != nil {
            return nil, err
        }

        if firstBlock && block.Type != 1 {
            return nil, errors.New("key not found")
        }
        firstBlock = false

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

// Closes all open file descriptors to ensure no data is lost and saves the configuration.
// Be sure to always close your database to avoid data corruption and loss.
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

// Remove a value and key from the database
func (db *Database) Delete(key uint) error {

    block := &Block{
        Type: int8(0),
        Length: 0,
        Data: []byte{},
        NextBlock: -1,
    }

    err := db.writeBlock(key, block)

    return err
}
