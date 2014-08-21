package chunked

import (
    "testing"
    "os"
)

func TestCreateDatabase(t *testing.T) {

    // Remove test db
    err := os.RemoveAll("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    config := DefaultConfig
    config.Blocksize = 32
    config.Chunksize = 30
    db, err := Create("./test-db", config)
    if err != nil {
        t.Fatal(err)
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestRecreateDatabase(t *testing.T) {
    _, err := Create("./test-db", DefaultConfig)
    if err != nil {
        if err.Error() != "directory already used" {
            t.Fatal("wrong error")
        }
    } else {
        t.Fail()
    }
}

func TestOpenDatabase(t *testing.T) {
    db, err := Open("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestAppend(t *testing.T) {
    db, err := Open("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    key, err := db.Append([]byte("hello world"))
    if err != nil {
        t.Fatal(err)
    }

    if key != 0 {
        t.Fatal("wrong key")
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestRead(t *testing.T) {
    db, err := Open("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    value, err := db.Read(0)
    if err != nil {
        t.Fatal(err)
    }

    if string(value) != "hello world" {
        t.Fatal("wrong value")
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestBigAppendAndRead(t *testing.T) {
    db, err := Open("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    key, err := db.Append([]byte("a strign bigger then the blocksize foo bar lol"))
    if err != nil {
        t.Fatal(err)
    }

    if key != 1 {
        t.Fatal("wrong key")
    }

    value, err := db.Read(key)
    if err != nil {
        t.Fatal(err)
    }

    if string(value) != "a strign bigger then the blocksize foo bar lol" {
        t.Fatal("wrong value")
    }

    // Reading continuation block
    value, err = db.Read(key + 1)
    if value != nil {
        t.Fatal("value should not be presented");
    }
    if err.Error() != "key not found" {
        t.Fatal("key shouldn't be found")
    }

    // Reading key bigger than blocksize
    value, err = db.Read(99)
    if value != nil {
        t.Fatal("value should not be presented");
    }
    if err.Error() != "key not found" {
        t.Fatal("key shouldn't be found")
    }

    value, err = db.Read(20)
    if value != nil {
        t.Fatal("value should not be presented");
    }
    if err.Error() != "key not found" {
        t.Fatal("key shouldn't be found")
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}

func TestDelete(t *testing.T) {
    db, err := Open("./test-db")
    if err != nil {
        t.Fatal(err)
    }

    if err := db.Delete(0); err != nil {
        t.Fatal(err)
    }

    // Try reading value
    value, err := db.Read(0)
    if value != nil {
        t.Fatal("value should not be presented");
    }
    if err.Error() != "key not found" {
        t.Fatal("key shouldn't be found")
    }

    if err = db.Close(); err != nil {
        t.Fatal(err)
    }
}
