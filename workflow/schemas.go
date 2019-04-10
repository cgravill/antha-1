// Code generated by go-bindata. DO NOT EDIT.
// sources:
// schemas/workflow.schema.json (6.142kB)

package workflow

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _workflowSchemaJson = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x58\x4d\x6f\xdb\x46\x13\xbe\xeb\x57\x0c\x18\x01\x7e\xdf\x54\x22\x8d\x5e\x8a\xea\xd6\xca\x41\xea\x02\x8e\x85\x5a\x68\x81\x06\x4e\x30\x22\x47\xe2\xc6\xe4\x2e\x33\xbb\x94\xcd\xa6\xfe\xef\x05\x29\x8a\xa2\xa4\x25\x4d\xd9\x52\x92\x1b\xb5\x3b\xdf\xcf\x7c\xad\xbe\xf4\x00\x00\x9c\xbe\x08\x9c\x11\x38\xa1\x31\x89\x1e\x79\x1e\x4a\x13\xa2\xeb\xab\xd8\xbb\x57\x7c\x37\x8f\xd4\xbd\x1e\x6a\x3f\xa4\x18\x9d\x41\xc9\x50\xfe\x2c\x99\x46\x9e\xf7\x49\x2b\x59\x12\xb9\x8a\x17\x5e\xc0\x38\x37\xc3\xf3\x9f\xbc\xd5\xd9\xab\x35\x67\x40\xda\x67\x91\x18\xa1\x64\xce\xfd\xfb\xcd\xf5\x3b\xb8\x29\x48\x60\xae\x18\x56\xd7\x33\x21\x17\x50\xe9\x5e\xb3\x9a\x2c\xa1\x9c\x47\xcd\x3e\x91\x6f\xd6\xa7\x4c\x9f\x53\xc1\x94\x3b\xf0\xbe\x38\x29\x4e\x57\x22\xff\x24\xd6\xb9\xa2\xe2\xfc\xb6\x64\xc0\x20\x10\xb9\x7a\x8c\x26\xac\x12\x62\x23\x48\x3b\x23\x98\x63\xa4\xa9\x24\x49\xea\x17\x5f\x36\x52\xff\x2a\x4d\xba\x0c\xb6\xce\xb7\xac\xd3\x86\x85\x5c\x94\xd6\x55\xb7\x09\x1a\x43\x5c\xb8\xfc\xe1\x3d\x0e\xff\xf9\x65\xf8\xf7\xf9\xf0\xe7\x8f\x30\xbc\x7d\xdd\x77\x2a\xd2\xc7\x0d\x97\x73\x45\x06\x9b\xb5\xc8\x34\x8a\x76\x75\xf4\x7d\x15\xc7\x24\x4d\x7e\x3f\xbd\xbe\xb8\x1e\x81\x88\x93\x88\xf2\x23\x28\x91\x81\xdf\x70\x49\xf2\xcc\x80\x26\x92\xa0\x24\x81\x9a\x83\x09\x49\x13\x08\x99\x7f\xc0\xbd\x88\x02\xc8\xc8\xd8\x8d\xba\x94\x4b\x92\x46\x71\xd6\x6c\xd9\x16\x3a\x5f\xd3\xb6\x6d\xc8\x0f\xc4\x87\x64\x1a\xe7\x19\xe4\xfc\xe8\x9e\x3b\xb7\x56\xf9\x7f\x50\xa2\xb4\x30\x8a\x77\xd3\xa2\x83\xfb\x65\x42\x65\xef\x30\xb6\x30\xef\x65\xc8\xff\x38\xd7\xe5\x56\x75\xf8\xef\x42\x98\x30\x9d\xe5\x9f\xff\xf7\x9c\x2d\xde\x47\x7b\xa2\x4d\xec\x19\x5c\x91\xb9\xaf\xad\xe7\x2b\xac\x98\xe6\xb9\x19\xaf\xbc\x80\xe6\x42\x16\xc5\xa2\x3d\x5e\xbb\x9f\x39\x7b\x6c\x8f\x3d\xfb\xaf\x7a\xf8\xde\xac\xd0\x7e\x6e\xe8\x1a\x3d\x99\x66\x49\xc3\xd5\x96\x70\x64\xc6\x6c\x47\x76\x45\x24\x0c\xc5\xcd\x32\x5a\x82\x42\x2b\xa7\x72\x1b\xf6\xa3\xb2\x1f\x19\xd8\x07\x0c\x56\x65\xa5\x0d\x4a\xbf\x8b\x1f\xd6\x20\x55\x54\x4f\xe7\xd9\xc6\xa1\xbd\x8a\xf4\x95\x9c\x0b\x8e\xc1\x84\x42\x03\xd3\x82\x1e\x5c\x98\xe6\xdf\x42\x17\xe5\x27\x31\x2e\x6a\x12\x25\x94\x7e\x83\x28\x0d\x1f\x00\xb9\x0b\x17\xce\xe2\xec\x23\x46\xe2\x73\xaa\xcc\x59\x83\x89\xd0\xde\x0d\x7f\xe8\x37\xc4\xb1\xc9\xe1\x4e\xe9\x5e\x91\xb7\xa4\xfd\x26\x30\x6d\x48\xaf\x91\xb2\x5b\x09\x56\xc4\xed\xa7\xad\x79\x30\x56\x52\x92\x5f\x68\xfe\xc6\xa9\xbd\x36\x69\x63\x51\xe7\x44\x6f\xed\x53\x4f\x4f\xe1\x8a\xb4\x3e\xe1\xed\xad\x79\x9c\x67\xee\xe2\x04\x4d\xd9\x52\x24\x21\xf9\x77\x4f\x97\x08\x04\xb4\x14\x3e\x81\x1f\xa1\xd6\xeb\xea\x78\x2b\x22\xad\xe4\x44\x24\x64\xae\xf0\xe1\xcc\x85\x4b\x03\x31\x66\x30\xa3\x82\xd9\x47\x9d\x7f\xa0\x29\x2a\x4c\xa6\x71\x2e\x35\x61\x9a\x13\xe3\x2c\x22\xd7\x02\x71\x53\x25\xed\x17\xd2\x33\xe7\xc4\x87\xb7\x91\x9a\x61\x74\x25\x1e\x88\xfb\x9d\xbb\xd3\x71\x72\xc1\x86\x87\x35\x55\x57\xb1\x1e\xe7\xa1\x6e\x77\xba\x21\x93\x7a\x35\x5a\xa7\x26\x78\x7b\xf9\xab\x0d\xc0\x23\x8f\xb0\x0b\xc1\xe4\x5b\x05\xef\x29\x28\xb7\x97\x4e\xdd\xe4\x57\x46\xe9\x87\xc7\x95\x39\x56\x71\x2c\xcc\x0b\x64\x9e\xa2\x29\xd4\x02\x68\xef\x0f\xf5\x31\x7d\x64\xec\xaa\xad\xb0\x68\x23\xc7\x0d\x76\xb9\x32\x4d\xd0\xbc\x04\xc5\x93\x44\x7c\xc7\xed\xc1\xb6\xb1\xad\x28\x54\x23\xf4\xc8\x48\xbc\xd9\x80\x7c\x7c\x28\xac\xaf\xb1\x6e\x86\x77\x75\xa0\xa2\x7b\x78\x7a\x47\xd9\x3c\x03\xe3\x19\x71\xcb\x32\xd2\xb2\x84\x35\x77\x9b\xc3\xd5\x1c\xb4\xb3\x35\xe4\x9c\xe1\x94\x1a\x38\xec\x7d\xbb\x45\x8f\x33\x41\xc6\x98\x0c\x71\xf7\x95\xfa\x2b\x55\xce\x6e\x9a\x76\xaa\x95\xda\xf6\x75\xe4\xaa\xb9\x51\x29\x5b\x6a\xb1\xba\xb7\x4f\x5c\xbf\xb2\x67\x4a\x1c\x0b\x89\x51\xb7\x32\x9a\x22\x2f\xa8\x65\x7c\xbc\x40\xdb\x49\xc0\x2a\xa3\x33\xa8\x2c\xb7\xa3\x65\x31\xf0\x34\xcd\xad\xb1\x77\xee\xa9\x39\xa4\xb9\x55\xc5\xf2\xc2\xbe\x79\xca\x7a\xa9\x3c\x1f\xec\xda\x6b\x87\xa4\xbe\x13\x1e\x7b\x5d\x2b\x44\x7f\xe3\xb7\xfa\x61\xcf\x90\xdd\x77\xfa\x85\x62\xa1\x9f\xf3\x44\xff\x3e\x5e\xe8\xf5\x70\x84\x04\x1b\xc0\xda\x9c\xbe\x94\x49\x6a\x26\x11\x1a\x2a\xfe\x32\x72\x61\x9c\x32\x93\x34\x51\x36\x00\x94\x19\xdc\x51\xe6\x2d\x31\x4a\x49\x03\x32\xc1\x12\x23\x11\x40\x48\x4c\x2d\x61\x82\x2e\x73\xa4\x8a\xd0\x77\xf2\x70\x5e\x3d\x77\x7a\x8f\xbd\xff\x02\x00\x00\xff\xff\xc6\x10\xe4\x66\xfe\x17\x00\x00")

func workflowSchemaJsonBytes() ([]byte, error) {
	return bindataRead(
		_workflowSchemaJson,
		"workflow.schema.json",
	)
}

func workflowSchemaJson() (*asset, error) {
	bytes, err := workflowSchemaJsonBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "workflow.schema.json", size: 6142, mode: os.FileMode(0640), modTime: time.Unix(1554888838, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0x7, 0x19, 0xb9, 0xe7, 0x52, 0x23, 0xf1, 0x62, 0xc6, 0x3d, 0x9e, 0xf, 0xcd, 0x5b, 0xb, 0xc3, 0x68, 0xc0, 0x62, 0x78, 0xab, 0xf6, 0xe3, 0xda, 0x42, 0x74, 0x26, 0xd8, 0x73, 0xad, 0x86, 0xbe}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"workflow.schema.json": workflowSchemaJson,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"workflow.schema.json": &bintree{workflowSchemaJson, map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}