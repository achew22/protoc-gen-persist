// Copyright 2017, TCN Inc.
// All rights reserved.

// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:

//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of TCN Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.

// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package generator

import (
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type FileStruct struct {
	Desc          *descriptor.FileDescriptorProto
	ImportList    *Imports
	Dependency    bool        // if is dependency
	Structures    *StructList // all structures in the file
	AllStructures *StructList // all structures in all the files
	ServiceList   *Services
}

func NewFileStruct(desc *descriptor.FileDescriptorProto, allStructs *StructList, dependency bool) *FileStruct {
	ret := &FileStruct{
		Desc:          desc,
		ImportList:    EmptyImportList(),
		Structures:    &StructList{},
		ServiceList:   &Services{},
		AllStructures: allStructs,
		Dependency:    dependency,
	}
	return ret
}

func (f *FileStruct) GetOrigName() string {
	return f.Desc.GetName()
}

func (f *FileStruct) GetPackageName() string {
	return f.Desc.GetPackage()
}

func (f *FileStruct) GetFileName() string {
	return strings.Replace(f.Desc.GetName(), ".proto", ".persist.go", -1)
}

func (f *FileStruct) GetServices() *Services {
	return f.ServiceList
}

func (f *FileStruct) GetGoPackage() string {
	if f.Desc.Options != nil && f.Desc.GetOptions().GoPackage != nil {
		switch {
		case strings.Contains(f.Desc.GetOptions().GetGoPackage(), ";"):
			idx := strings.LastIndex(f.Desc.GetOptions().GetGoPackage(), ";")
			return f.Desc.GetOptions().GetGoPackage()[idx+1:]
		case strings.Contains(f.Desc.GetOptions().GetGoPackage(), "/"):
			idx := strings.LastIndex(f.Desc.GetOptions().GetGoPackage(), "/")
			return f.Desc.GetOptions().GetGoPackage()[idx+1:]
		default:
			return f.Desc.GetOptions().GetGoPackage()
		}

	} else {
		return strings.Replace(f.Desc.GetPackage(), ".", "_", -1)
	}
}

func (f *FileStruct) GetGoPath() string {
	if f.Desc.Options != nil && f.Desc.GetOptions().GoPackage != nil {
		switch {
		case strings.Contains(f.Desc.GetOptions().GetGoPackage(), ";"):
			idx := strings.LastIndex(f.Desc.GetOptions().GetGoPackage(), ";")
			return f.Desc.GetOptions().GetGoPackage()[0:idx]
		case strings.Contains(f.Desc.GetOptions().GetGoPackage(), "/"):
			return f.Desc.GetOptions().GetGoPackage()
		default:
			return f.Desc.GetOptions().GetGoPackage()
		}

	} else {
		return strings.Replace(f.Desc.GetPackage(), ".", "_", -1)
	}
}

func (f *FileStruct) ProcessImportsForType(name string) {
	typ := f.AllStructures.GetStructByProtoName(name)
	if typ != nil {
		for _, file := range *typ.GetImportedFiles() {
			if file.GetOrigName() != f.GetOrigName() {
				f.ImportList.GetOrAddImport(file.GetGoPackage(), file.GetGoPath())
			}
		}
	} else {
		logrus.Fatalf("Can't find structure %s!", name)
	}
}

func (f *FileStruct) ProcessImports() {
	for _, srv := range *f.ServiceList {
		if srv.IsServiceEnabled() {
			srv.ProcessImports()
			for _, m := range *srv.Methods {
				f.ProcessImportsForType(m.Desc.GetInputType())
				f.ProcessImportsForType(m.Desc.GetOutputType())
			}
		}
	}
}

func (f *FileStruct) Process() {
	// collect file defined messages
	for _, m := range f.Desc.GetMessageType() {
		s := f.AllStructures.AddMessage(m, nil, f.GetPackageName(), f)
		f.Structures.Append(s)
	}
	// collect file defined enums
	for _, e := range f.Desc.GetEnumType() {
		s := f.AllStructures.AddEnum(e, nil, f.GetPackageName(), f)
		f.Structures.Append(s)
	}

	for _, s := range f.Desc.GetService() {
		f.ServiceList.AddService(f.GetPackageName(), s, f.AllStructures, f)
	}

}

func (f *FileStruct) Generate() []byte {
	// f.Process()
	logrus.WithField("imports", f.ImportList).Debug("import list")
	return ExecuteFileTemplate(f)
}

// FileList ----------------

type FileList []*FileStruct

func NewFileList() *FileList {
	return &FileList{}
}

func (fl *FileList) FindFile(desc *descriptor.FileDescriptorProto) *FileStruct {
	for _, f := range *fl {
		if f.Desc.GetName() == desc.GetName() {
			return f
		}
	}
	return nil
}

func (fl *FileList) GetOrCreateFile(desc *descriptor.FileDescriptorProto, allStructs *StructList, dependency bool) *FileStruct {
	if f := fl.FindFile(desc); f != nil {
		return f
	}
	f := NewFileStruct(desc, allStructs, dependency)
	*fl = append(*fl, f)
	return f
}

func (fl *FileList) Process() {
	for _, file := range *fl {
		file.Process()
	}
}

func (fl *FileList) Append(file *FileStruct) {
	*fl = append(*fl, file)
}