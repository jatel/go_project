package filestore

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/BlockABC/wallet-btc-service/common/log"
	"github.com/BlockABC/wallet-btc-service/common/utility"
)

// file store block height and omni block height
const (
	BLOCK = "block"
	OMNI  = "omni"
)

type RepairStore struct {
	Filename string
	Rw       sync.RWMutex
}

var RepairStoreInstance = NewRepairStore()

func NewRepairStore() *RepairStore {
	baseDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	baseDir = strings.Replace(baseDir, "\\", "/", -1)
	return &RepairStore{Filename: baseDir + "/repairStore"}
}

func (store *RepairStore) GetBlockBegin() (error, int32) {
	if !utility.IsFileExist(store.Filename) {
		return nil, 0
	}

	store.Rw.RLock()
	defer store.Rw.RUnlock()

	fs, err := os.Open(store.Filename)
	if nil != err {
		log.Log.Error(err, " GetBlockBegin open store file fail")
		return err, 0
	}
	defer fs.Close()

	// read file
	br := bufio.NewReader(fs)
	for {
		info, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read store file fail")
			return err, 0
		}

		if isBlock := strings.HasPrefix(string(info), BLOCK); isBlock {
			result := strings.Split(string(info), "_")
			if 2 != len(result) {
				return errors.New("invalid block info string"), 0
			}
			height, err := strconv.ParseInt(result[1], 10, 32)
			if nil != err {
				return err, 0
			}

			return nil, int32(height)
		}
	}

	return nil, 0
}

func (store *RepairStore) GetOmniBegin() (error, int32) {
	if !utility.IsFileExist(store.Filename) {
		return nil, 0
	}

	store.Rw.RLock()
	defer store.Rw.RUnlock()

	fs, err := os.Open(store.Filename)
	if nil != err {
		log.Log.Error(err, " GetOmniBegin open store file fail")
		return err, 0
	}
	defer fs.Close()

	// read file
	br := bufio.NewReader(fs)
	for {
		info, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read store file fail")
			return err, 0
		}

		if isOmni := strings.HasPrefix(string(info), OMNI); isOmni {
			result := strings.Split(string(info), "_")
			if 2 != len(result) {
				return errors.New("invalid omni info string"), 0
			}
			height, err := strconv.ParseInt(result[1], 10, 32)
			if nil != err {
				return err, 0
			}

			return nil, int32(height)
		}
	}

	return nil, 0
}

func (store *RepairStore) SaveOmniBegin(height int32) error {
	store.Rw.Lock()
	defer store.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(store.Filename, os.O_CREATE|os.O_RDWR, 0666)
	if nil != err {
		log.Log.Error(err, " store omni begin open file fail, omni begin:", height)
		return err
	}
	defer fs.Close()

	// read file
	oldInfo := []string{}
	br := bufio.NewReader(fs)
	for {
		info, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read store file fail")
			return err
		}
		oldInfo = append(oldInfo, string(info))
	}
	fs.Seek(0, 0)

	// new omni
	bFind := false
	for i := 0; i < len(oldInfo); i++ {
		if isOmni := strings.HasPrefix(oldInfo[i], OMNI); isOmni {
			bFind = true
			oldInfo[i] = fmt.Sprintf("omni_%d", height)
			break
		}
	}

	if !bFind {
		oldInfo = append(oldInfo, fmt.Sprintf("omni_%d", height))
	}

	// write info
	buf := bufio.NewWriter(fs)
	for _, oneInfo := range oldInfo {
		if _, err := fmt.Fprintln(buf, oneInfo); nil != err {
			log.Log.Error(err, " store omni begin write file fail, omni height:", height)
			return err
		}
	}
	return buf.Flush()
}

func (store *RepairStore) SaveBlockBegin(height int32) error {
	store.Rw.Lock()
	defer store.Rw.Unlock()

	// open file
	fs, err := os.OpenFile(store.Filename, os.O_CREATE|os.O_RDWR, 0666)
	if nil != err {
		log.Log.Error(err, " store block begin open file fail, block begin:", height)
		return err
	}
	defer fs.Close()

	// read file
	oldInfo := []string{}
	br := bufio.NewReader(fs)
	for {
		info, _, err := br.ReadLine()
		if err == io.EOF {
			break
		} else if nil != err {
			log.Log.Error(err, " read store file fail")
			return err
		}
		oldInfo = append(oldInfo, string(info))
	}
	fs.Seek(0, 0)

	// new omni
	bFind := false
	for i := 0; i < len(oldInfo); i++ {
		if isOmni := strings.HasPrefix(oldInfo[i], BLOCK); isOmni {
			bFind = true
			oldInfo[i] = fmt.Sprintf("block_%d", height)
			break
		}
	}

	if !bFind {
		oldInfo = append(oldInfo, fmt.Sprintf("block_%d", height))
	}

	// write info
	buf := bufio.NewWriter(fs)
	for _, oneInfo := range oldInfo {
		if _, err := fmt.Fprintln(buf, oneInfo); nil != err {
			log.Log.Error(err, " store block begin write file fail, block height:", height)
			return err
		}
	}
	return buf.Flush()
}
