package resource

import "os"

const (
	// ArgMergeRaw indicates the merged data should be in same format of raw data
	ArgMergeRaw = "merge_raw"
)

type LayerFile struct {
	Path string
	Data []byte
}

type MergedFile struct {
	Data []byte
}

type Descriptor interface {
	ReadResource(fs Fs) ([]LayerFile, error)
	MatchResource(path string) bool
	Merge(layers []LayerFile, args map[string]interface{}) (*MergedFile, error)
	Parse(merged *MergedFile) (interface{}, error)
}

type SimpleFile struct {
	Name    string
	MergeFn func(layers []LayerFile) ([]byte, error)
	ParseFn func(data []byte) (interface{}, error)
}

func (f SimpleFile) ReadResource(fs Fs) ([]LayerFile, error) {
	data, err := ReadFile(fs, f.Name)
	if os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return []LayerFile{{Path: f.Name, Data: data}}, nil
}

func (f SimpleFile) MatchResource(path string) bool {
	return path == f.Name
}

func (f SimpleFile) Merge(layers []LayerFile, args map[string]interface{}) (*MergedFile, error) {
	if f.MergeFn != nil {
		data, err := f.MergeFn(layers)
		if err != nil {
			return nil, err
		}
		return &MergedFile{Data: data}, nil
	}
	file := layers[len(layers)-1]
	return &MergedFile{Data: file.Data}, nil
}

func (f SimpleFile) Parse(merged *MergedFile) (interface{}, error) {
	if f.ParseFn == nil {
		return merged.Data, nil
	}
	return f.ParseFn(merged.Data)
}