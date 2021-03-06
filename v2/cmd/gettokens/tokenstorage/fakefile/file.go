package fakefile

import "io"

type File struct {
	data []byte
	ptr  int
}

func (f *File) String() string {
	return string(f.data)
}

func (f *File) Read(p []byte) (int, error) {
	cnt := 0
	for i := range p {
		fi := i + f.ptr
		if fi >= len(f.data) {
			return cnt, io.EOF
		}

		p[i] = f.data[fi]
		cnt++
	}

	f.ptr += cnt

	return len(p), nil
}

func (f *File) Write(p []byte) (int, error) {
	for i := range p {
		if i >= len(f.data) {
			f.data = append(f.data, p[i])
		} else {
			f.data[i] = p[i]
		}
	}

	f.ptr += len(p)
	return len(p), nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	f.ptr = int(offset)
	return int64(f.ptr), nil
}

func New(initialData ...byte) *File {
	return &File{data: initialData}
}
