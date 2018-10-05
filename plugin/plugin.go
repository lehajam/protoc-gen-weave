// Copy makes a new blog with the same data
// func (b *Blog) Copy() github_com_iov_one_weave_orm.CloneableData {
// 	var cpy *Blog
// 	copier.Copy(cpy, b)
// 	return cpy
// }

// const PostBucketName = "posts"

// // PostBucket is a type-safe wrapper around orm.Bucket
// type PostBucket struct {
// 	orm.Bucket
// }

// // NewPostBucket initializes a PostBucket with default name
// //
// // inherit Get and Save from orm.Bucket
// // add run-time check on Save
// func NewPostBucket() PostBucket {
// 	bucket := orm.NewBucket(PostBucketName,
// 		orm.NewSimpleObj(nil, new(Post))).
// 		WithIndex("author", idxAuthor, false)
// 	return PostBucket{
// 		Bucket: bucket,
// 	}
// }

// func idxAuthor(obj orm.Object) ([]byte, error) {
// 	// these should use proper errors, but they never occur
// 	// except in case of developer error (wrong data in wrong bucket)
// 	if obj == nil {
// 		return nil, errors.New("Cannot take index of nil")
// 	}
// 	post, ok := obj.Value().(*Post)
// 	if !ok {
// 		return nil, errors.New("Can only take index of Post")
// 	}
// 	return post.Author, nil
// }

// // Save enforces the proper type
// func (b PostBucket) Save(db weave.KVStore, obj orm.Object) error {
// 	if _, ok := obj.Value().(*Post); !ok {
// 		return orm.ErrInvalidObject(obj.Value())
// 	}
// 	return b.Bucket.Save(db, obj)
// }

// func idxAuthor(obj orm.Object) ([]byte, error) {
// 	// these should use proper errors, but they never occur
// 	// except in case of developer error (wrong data in wrong bucket)
// 	if obj == nil {
// 		return nil, errors.New("Cannot take index of nil")
// 	}
// 	post, ok := obj.Value().(*Post)
// 	if !ok {
// 		return nil, errors.New("Can only take index of Post")
// 	}
// 	return post.Author, nil
// }

package plugin

import (
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/protoc-gen-gogo/generator"
	"github.com/gogo/protobuf/vanity"
	"github.com/iancoleman/strcase"
	"github.com/lehajam/protoc-gen-weave/x/bucket"
)

type plugin struct {
	*generator.Generator
	generator.PluginImports
	useGogoImport bool
}

func NewPlugin(useGogoImport bool) generator.Plugin {
	return &plugin{useGogoImport: useGogoImport}
}

func (p *plugin) Name() string {
	return "bucket"
}

func (p *plugin) Init(g *generator.Generator) {
	p.Generator = g
}

type index struct {
	name      string
	fieldName string
	unique    bool
}

func (p *plugin) Generate(file *generator.FileDescriptor) {
	if !p.useGogoImport {
		vanity.TurnOffGogoImport(file.FileDescriptorProto)
	}

	p.PluginImports = generator.NewPluginImports(p.Generator)
	weavePkg := p.NewImport("github.com/iov-one/weave")
	ormPkg := p.NewImport("github.com/iov-one/weave/orm")
	copierPkg := p.NewImport("github.com/jinzhu/copier")

	for _, msg := range file.Messages() {

		if !strings.HasSuffix(msg.GetName(), "Msg") {
			state := msg.GetName()
			bucketName := strings.ToLower(state + "s")
			bucketStruct := state + "Bucket"
			indexList := getBucketIndexList(msg)

			p.P(`func (b *`, state, `) Copy() `, ormPkg.Use(), `.CloneableData {`)
			p.In()
			p.P(`var cpy *`, state)
			p.P(copierPkg.Use(), `.Copy(cpy, b)`)
			p.P(`return cpy`)
			p.Out()
			p.P(`}`)
			p.P(``)
			p.P(`const ` + bucketStruct + `Name = "` + bucketName + `"`)
			p.P(``)
			p.P(`type ` + bucketStruct + ` struct {`)
			p.In()
			p.P(ormPkg.Use(), `.Bucket`)
			p.Out()
			p.P(`}`)
			p.P(``)
			p.P(`func New`, bucketStruct, `() `, bucketStruct, ` {`)
			p.In()
			p.P(`bucket := `, ormPkg.Use(), `.NewBucket(`, bucketStruct, `Name`, `,`)
			p.In()
			if len(indexList) == 0 {
				p.P(ormPkg.Use(), `.NewSimpleObj(nil, new(`, state, `)))`)
			} else {
				p.P(ormPkg.Use(), `.NewSimpleObj(nil, new(`, state, `))).`)
				for k, idx := range indexList {
					if k == len(indexList)-1 {
						p.P(`WithIndex("`, strcase.ToLowerCamel(idx.name), `", idx`, strcase.ToCamel(idx.name), `, `, idx.unique, `)`)
					} else {
						p.P(`WithIndex("`, strcase.ToLowerCamel(idx.name), `", idx`, strcase.ToCamel(idx.name), `, `, idx.unique, `).`)
					}
				}
			}
			p.Out()
			p.P(`return ` + bucketStruct + `{ Bucket: bucket }`)
			p.Out()
			p.P(`}`)
			p.P(``)
			p.P(`func (b `+bucketStruct+`) Save(db `, weavePkg.Use(), `.KVStore, obj `, ormPkg.Use(), `.Object) error {`)
			p.In()
			p.P(`if _, ok := obj.Value().(*` + state + `); !ok {`)
			p.In()
			p.P(`return `, ormPkg.Use(), `.ErrInvalidObject(obj.Value())`)
			p.Out()
			p.P(`}`)
			p.P(`return b.Bucket.Save(db, obj)`)
			p.Out()
			p.P(`}`)
			p.P(``)
			for _, idx := range indexList {
				p.P(`func idx`, strcase.ToCamel(idx.name), `(obj `, ormPkg.Use(), `.Object) ([]byte, error) {`)
				p.In()
				p.P(`if obj == nil {`)
				p.In()
				p.P(`return nil, fmt.Errorf("Cannot take index of nil")`)
				p.Out()
				p.P(`}`)
				p.P(`objAs, ok := obj.Value().(*`, state, `)`)
				p.P(`if !ok {`)
				p.In()
				p.P(`return nil, fmt.Errorf("Can only take index of objAs")`)
				p.Out()
				p.P(`}`)
				p.P(`return objAs.`, strcase.ToCamel(idx.fieldName), `, nil`)
				p.Out()
				p.P(`}`)
			}
		}
	}
}

func getBucketIndexList(msg *generator.Descriptor) []index {
	var indexList []index
	for _, field := range msg.Field {
		if field.Options != nil {
			v, err := proto.GetExtension(field.Options, bucket.E_Index)
			if err == nil && v.(*bucket.FieldIndex) != nil {
				fieldIndex := v.(*bucket.FieldIndex)
				name := fieldIndex.GetName()
				if name == "" {
					name = field.GetName()
				}

				indexList = append(indexList, index{name, field.GetName(), fieldIndex.GetUnique()})
			}
		}
	}
	return indexList
}
