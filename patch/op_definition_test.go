package patch_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"

	. "github.com/SUSE/go-patch/patch"
)

var _ = Describe("NewOpsFromDefinitions", func() {
	var (
		path                    = "/abc"
		from                    = "/abc"
		invalidPath             = "abc"
		invalidFrom             = "abc"
		errorMsg                = "error"
		val         interface{} = 123
		complexVal  interface{} = map[interface{}]interface{}{123: 123}
		trueBool                = true
	)

	It("supports 'replace', 'remove', 'test', 'qcopy', 'qmove' operations", func() {
		opDefs := []OpDefinition{
			{Type: "replace", Path: &path, Value: &val},
			{Type: "remove", Path: &path},
			{Type: "test", Path: &path, Value: &val},
			{Type: "test", Path: &path, Absent: &trueBool},
			{Type: "qcopy", Path: &path, From: &from},
			{Type: "qmove", Path: &path, From: &from},
		}

		ops, err := NewOpsFromDefinitions(opDefs)
		Expect(err).ToNot(HaveOccurred())

		Expect(ops).To(Equal(Ops([]Op{
			ReplaceOp{Path: MustNewPointerFromString("/abc"), Value: 123},
			RemoveOp{Path: MustNewPointerFromString("/abc")},
			TestOp{Path: MustNewPointerFromString("/abc"), Value: 123},
			TestOp{Path: MustNewPointerFromString("/abc"), Absent: true},
			QCopyOp{Path: MustNewPointerFromString("/abc"), From: MustNewPointerFromString("/abc")},
			QMoveOp{Path: MustNewPointerFromString("/abc"), From: MustNewPointerFromString("/abc")},
		})))
	})

	It("returns error if operation type is unknown", func() {
		_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "op"}})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`Unknown operation [0] with type 'op' within
{
  "Type": "op"
}`))
	})

	It("returns error if operation type is find since it's not useful in list of operations", func() {
		_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "find"}})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Unknown operation [0] with type 'find'"))
	})

	It("allows values to be complex in error messages", func() {
		_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "op", Path: &invalidPath, Value: &complexVal}})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(Equal(`Unknown operation [0] with type 'op' within
{
  "Type": "op",
  "Path": "abc",
  "Value": "<redacted>"
}`))
	})

	Describe("replace", func() {
		It("allows error description", func() {
			opDefs := []OpDefinition{{Type: "replace", Path: &path, Value: &val, Error: &errorMsg}}

			ops, err := NewOpsFromDefinitions(opDefs)
			Expect(err).ToNot(HaveOccurred())

			Expect(ops).To(Equal(Ops([]Op{
				DescriptiveOp{
					Op:       ReplaceOp{Path: MustNewPointerFromString("/abc"), Value: 123},
					ErrorMsg: errorMsg,
				},
			})))
		})

		It("requires path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "replace"}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Replace operation [0]: Missing path within
{
  "Type": "replace"
}`))
		})

		It("requires value", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "replace", Path: &path}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Replace operation [0]: Missing value within
{
  "Type": "replace",
  "Path": "/abc"
}`))
		})

		It("requires valid path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "replace", Path: &invalidPath, Value: &val}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Replace operation [0]: Invalid path: Expected to start with '/' within
{
  "Type": "replace",
  "Path": "abc",
  "Value": "<redacted>"
}`))
		})
	})

	Describe("remove", func() {
		It("allows error description", func() {
			opDefs := []OpDefinition{{Type: "remove", Path: &path, Error: &errorMsg}}

			ops, err := NewOpsFromDefinitions(opDefs)
			Expect(err).ToNot(HaveOccurred())

			Expect(ops).To(Equal(Ops([]Op{
				DescriptiveOp{
					Op:       RemoveOp{Path: MustNewPointerFromString("/abc")},
					ErrorMsg: errorMsg,
				},
			})))
		})

		It("requires path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "remove"}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Remove operation [0]: Missing path within
{
  "Type": "remove"
}`))
		})

		It("does not allow value", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "remove", Path: &path, Value: &val}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Remove operation [0]: Cannot specify value within
{
  "Type": "remove",
  "Path": "/abc",
  "Value": "<redacted>"
}`))
		})

		It("requires valid path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "remove", Path: &invalidPath}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Remove operation [0]: Invalid path: Expected to start with '/' within
{
  "Type": "remove",
  "Path": "abc"
}`))
		})
	})

	Describe("test", func() {
		It("allows error description", func() {
			opDefs := []OpDefinition{{Type: "test", Path: &path, Value: &val, Error: &errorMsg}}

			ops, err := NewOpsFromDefinitions(opDefs)
			Expect(err).ToNot(HaveOccurred())

			Expect(ops).To(Equal(Ops([]Op{
				DescriptiveOp{
					Op:       TestOp{Path: MustNewPointerFromString("/abc"), Value: 123},
					ErrorMsg: errorMsg,
				},
			})))
		})

		It("requires path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "test"}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Test operation [0]: Missing path within
{
  "Type": "test"
}`))
		})

		It("requires value or absent flag", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "test", Path: &path}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Test operation [0]: Missing value or absent within
{
  "Type": "test",
  "Path": "/abc"
}`))
		})

		It("requires valid path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "test", Path: &invalidPath, Value: &val}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`Test operation [0]: Invalid path: Expected to start with '/' within
{
  "Type": "test",
  "Path": "abc",
  "Value": "<redacted>"
}`))
		})
	})

	Describe("qcopy", func() {
		It("requires path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qcopy", From: &from}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QCopy operation [0]: Missing path within
{
  "Type": "qcopy",
  "From": "/abc"
}`))
		})

		It("requires from", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qcopy", Path: &path}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QCopy operation [0]: Missing from within
{
  "Type": "qcopy",
  "Path": "/abc"
}`))
		})

		It("does not allow value", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qcopy", Path: &path, From: &from, Value: &val}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QCopy operation [0]: Cannot specify value within
{
  "Type": "qcopy",
  "Path": "/abc",
  "From": "/abc",
  "Value": "<redacted>"
}`))
		})

		It("requires valid path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qcopy", Path: &invalidPath, From: &from}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QCopy operation [0]: Invalid path: Expected to start with '/' within
{
  "Type": "qcopy",
  "Path": "abc",
  "From": "/abc"
}`))
		})

		It("requires valid from", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qcopy", Path: &path, From: &invalidFrom}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QCopy operation [0]: Invalid from: Expected to start with '/' within
{
  "Type": "qcopy",
  "Path": "/abc",
  "From": "abc"
}`))
		})
	})

	Describe("qmove", func() {
		It("requires path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qmove", From: &from}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QMove operation [0]: Missing path within
{
  "Type": "qmove",
  "From": "/abc"
}`))
		})

		It("requires from", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qmove", Path: &path}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QMove operation [0]: Missing from within
{
  "Type": "qmove",
  "Path": "/abc"
}`))
		})

		It("does not allow value", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qmove", Path: &path, From: &from, Value: &val}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QMove operation [0]: Cannot specify value within
{
  "Type": "qmove",
  "Path": "/abc",
  "From": "/abc",
  "Value": "<redacted>"
}`))
		})

		It("requires valid path", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qmove", Path: &invalidPath, From: &from}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QMove operation [0]: Invalid path: Expected to start with '/' within
{
  "Type": "qmove",
  "Path": "abc",
  "From": "/abc"
}`))
		})

		It("requires valid from", func() {
			_, err := NewOpsFromDefinitions([]OpDefinition{{Type: "qmove", Path: &path, From: &invalidFrom}})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(`QMove operation [0]: Invalid from: Expected to start with '/' within
{
  "Type": "qmove",
  "Path": "/abc",
  "From": "abc"
}`))
		})
	})
})

var _ = Describe("NewOpDefinitionsFromOps", func() {
	It("supports 'replace', 'remove', 'test', 'qcopy', 'qmove' operations serialized", func() {
		ops := Ops([]Op{
			ReplaceOp{Path: MustNewPointerFromString("/abc"), Value: 123},
			RemoveOp{Path: MustNewPointerFromString("/abc")},
			TestOp{Path: MustNewPointerFromString("/abc"), Value: 123},
			TestOp{Path: MustNewPointerFromString("/abc"), Absent: true},
			QCopyOp{Path: MustNewPointerFromString("/abc"), From: MustNewPointerFromString("/abc")},
			QMoveOp{Path: MustNewPointerFromString("/abc"), From: MustNewPointerFromString("/abc")},
		})

		opDefs, err := NewOpDefinitionsFromOps(ops)
		Expect(err).ToNot(HaveOccurred())

		bs, err := yaml.Marshal(opDefs)
		Expect(err).ToNot(HaveOccurred())
		Expect("\n" + string(bs)).To(Equal(`
- type: replace
  path: /abc
  value: 123
- type: remove
  path: /abc
- type: test
  path: /abc
  value: 123
- type: test
  path: /abc
  absent: true
- type: qcopy
  path: /abc
  from: /abc
- type: qmove
  path: /abc
  from: /abc
`))

		bs, err = json.MarshalIndent(opDefs, "", "    ")
		Expect(string(bs)).To(Equal(`[
    {
        "Type": "replace",
        "Path": "/abc",
        "Value": 123
    },
    {
        "Type": "remove",
        "Path": "/abc"
    },
    {
        "Type": "test",
        "Path": "/abc",
        "Value": 123
    },
    {
        "Type": "test",
        "Path": "/abc",
        "Absent": true
    },
    {
        "Type": "qcopy",
        "Path": "/abc",
        "From": "/abc"
    },
    {
        "Type": "qmove",
        "Path": "/abc",
        "From": "/abc"
    }
]`))
	})
})
