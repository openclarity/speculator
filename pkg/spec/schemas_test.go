// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spec

import (
	"encoding/json"
	"reflect"
	"testing"

	spec "github.com/getkin/kin-openapi/openapi3"
)

var (
	stringNumberObject = createObjectSchema(map[string]*spec.Schema{
		spec.TypeString: spec.NewStringSchema(),
		spec.TypeNumber: spec.NewFloat64Schema(),
	})
	stringBooleanObject = createObjectSchema(map[string]*spec.Schema{
		spec.TypeString:  spec.NewStringSchema(),
		spec.TypeBoolean: spec.NewBoolSchema(),
	})
	stringIntegerObject = createObjectSchema(map[string]*spec.Schema{
		spec.TypeString:  spec.NewStringSchema(),
		spec.TypeInteger: spec.NewInt64Schema(),
	})
)

func marshal(obj interface{}) string {
	objB, _ := json.Marshal(obj)
	return string(objB)
}

func createObjectSchema(properties map[string]*spec.Schema) *spec.Schema {
	return spec.NewObjectSchema().WithProperties(properties)
}

func createObjectSchemaWithRef(properties map[string]*spec.SchemaRef) *spec.Schema {
	objectSchema := spec.NewObjectSchema()
	for name, ref := range properties {
		objectSchema.WithPropertyRef(name, ref)
	}

	return objectSchema
}

func Test_findDefinition(t *testing.T) {
	type args struct {
		schemas spec.Schemas
		schema  *spec.Schema
	}
	tests := []struct {
		name        string
		args        args
		wantDefName string
		wantExist   bool
	}{
		{
			name: "identical string schema exist",
			args: args{
				schemas: spec.Schemas{
					"string": &spec.SchemaRef{Value: spec.NewStringSchema()},
				},
				schema: spec.NewStringSchema(),
			},
			wantDefName: "string",
			wantExist:   true,
		},
		{
			name: "identical string schema does not exist",
			args: args{
				schemas: spec.Schemas{
					"string": &spec.SchemaRef{Value: spec.NewStringSchema().WithFormat("format")},
				},
				schema: spec.NewStringSchema(),
			},
			wantDefName: "",
			wantExist:   false,
		},
		{
			name: "identical object schema exist (object order is different)",
			args: args{
				schemas: spec.Schemas{
					"object": &spec.SchemaRef{Value: spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
						spec.TypeObject: stringIntegerObject,
						spec.TypeString: spec.NewStringSchema(),
					})},
				},
				schema: createObjectSchema(
					map[string]*spec.Schema{
						spec.TypeString: spec.NewStringSchema(),
						spec.TypeObject: createObjectSchema(
							map[string]*spec.Schema{
								spec.TypeInteger: spec.NewInt64Schema(),
								spec.TypeString:  spec.NewStringSchema(),
							},
						),
					},
				),
			},
			wantDefName: "object",
			wantExist:   true,
		},
		{
			name: "identical object schema does not exist",
			args: args{
				schemas: spec.Schemas{
					"object": &spec.SchemaRef{Value: spec.NewObjectSchema().WithProperties(map[string]*spec.Schema{
						spec.TypeString: spec.NewStringSchema(),
						spec.TypeObject: stringIntegerObject,
					})},
				},
				schema: createObjectSchema(
					map[string]*spec.Schema{
						spec.TypeString: spec.NewStringSchema(),
						spec.TypeObject: stringNumberObject,
					},
				),
			},
			wantDefName: "",
			wantExist:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDefName, gotExist := findScheme(tt.args.schemas, tt.args.schema)
			if gotDefName != tt.wantDefName {
				t.Errorf("findScheme() gotDefName = %v, want %v", gotDefName, tt.wantDefName)
			}
			if gotExist != tt.wantExist {
				t.Errorf("findScheme() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
		})
	}
}

func Test_getUniqueDefName(t *testing.T) {
	type args struct {
		schemas spec.Schemas
		name    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "name does not exist",
			args: args{
				schemas: spec.Schemas{
					"string": &spec.SchemaRef{Value: stringIntegerObject},
				},
				name: "no-test",
			},
			want: "no-test_0",
		},
		{
			name: "name exist once",
			args: args{
				schemas: spec.Schemas{
					"test_0": &spec.SchemaRef{Value: stringIntegerObject},
				},
				name: "test",
			},
			want: "test_1",
		},
		{
			name: "name exist multiple times",
			args: args{
				schemas: spec.Schemas{
					"test":   &spec.SchemaRef{Value: stringIntegerObject},
					"test_0": &spec.SchemaRef{Value: stringNumberObject},
					"test_1": &spec.SchemaRef{Value: stringBooleanObject},
				},
				name: "test",
			},
			want: "test_2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUniqueSchemeName(tt.args.schemas, tt.args.name); got != tt.want {
				t.Errorf("getUniqueSchemeName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createArraySchemaWithRefItems(name string) *spec.Schema {
	arraySchemaWithRefItems := spec.NewArraySchema()
	arraySchemaWithRefItems.Items = spec.NewSchemaRef(schemasRefPrefix+name, nil)
	return arraySchemaWithRefItems
}

func Test_schemaToRef(t *testing.T) {

	type args struct {
		schemas     spec.Schemas
		schema      *spec.Schema
		defNameHint string
		depth       int
	}
	tests := []struct {
		name           string
		args           args
		wantRetSchemas spec.Schemas
		wantRetSchema  *spec.SchemaRef
	}{
		{
			name: "nil schema",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema:      nil,
				defNameHint: "",
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
			},
			wantRetSchema: nil,
		},
		{
			name: "array schema with nil items",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema:      spec.NewArraySchema().WithItems(nil),
				defNameHint: "",
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
			},
			wantRetSchema: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(nil)),
		},
		{
			name: "array schema with non object items - no change for definitions",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema:      spec.NewArraySchema().WithItems(spec.NewBoolSchema()),
				defNameHint: "",
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
			},
			wantRetSchema: spec.NewSchemaRef("", spec.NewArraySchema().WithItems(spec.NewBoolSchema())),
		},
		{
			name: "array schema with object items - use hint name",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema:      spec.NewArraySchema().WithItems(stringNumberObject),
				defNameHint: "hint",
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				"hint": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef("", createArraySchemaWithRefItems("hint")),
		},
		{
			name: "array schema with object items - hint name already exist",
			args: args{
				schemas: spec.Schemas{
					"hint": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema:      spec.NewArraySchema().WithItems(stringNumberObject),
				defNameHint: "hint",
			},
			wantRetSchemas: spec.Schemas{
				"hint":   &spec.SchemaRef{Value: spec.NewBoolSchema()},
				"hint_0": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef("", createArraySchemaWithRefItems("hint")),
		},
		{
			name: "primitive type",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
				},
				schema: spec.NewInt64Schema(),
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: spec.NewBoolSchema()},
			},
			wantRetSchema: spec.NewSchemaRef("", spec.NewInt64Schema()),
		},
		{
			name: "empty object - no new definition",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: stringNumberObject},
				},
				schema: spec.NewObjectSchema(),
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef("", spec.NewObjectSchema()),
		},
		{
			name: "object - definition exist",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: stringNumberObject},
				},
				schema: stringNumberObject,
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"test", nil),
		},
		{
			name: "object - definition does not exist",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: stringBooleanObject},
				},
				schema: stringNumberObject,
			},
			wantRetSchemas: spec.Schemas{
				"test":          &spec.SchemaRef{Value: stringBooleanObject},
				"number_string": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"number_string", nil),
		},
		{
			name: "object - definition does not exist - use hint",
			args: args{
				schemas: spec.Schemas{
					"test": &spec.SchemaRef{Value: stringBooleanObject},
				},
				schema:      stringNumberObject,
				defNameHint: "hint",
			},
			wantRetSchemas: spec.Schemas{
				"test": &spec.SchemaRef{Value: stringBooleanObject},
				"hint": &spec.SchemaRef{Value: stringNumberObject},
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"hint", nil),
		},
		{
			name: "object in object",
			args: args{
				schemas: nil,
				schema: createObjectSchema(
					map[string]*spec.Schema{
						spec.TypeString: spec.NewStringSchema(),
						spec.TypeObject: stringNumberObject,
					},
				),
			},
			wantRetSchemas: spec.Schemas{
				"object": spec.NewSchemaRef("", stringNumberObject),
				"object_string": spec.NewSchemaRef("", createObjectSchemaWithRef(
					map[string]*spec.SchemaRef{
						spec.TypeObject: spec.NewSchemaRef(schemasRefPrefix+"object", nil),
						spec.TypeString: spec.NewSchemaRef("", spec.NewStringSchema()),
					},
				)),
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"object_string", nil),
		},
		{
			name: "array of object in an object",
			args: args{
				schemas: nil,
				schema: createObjectSchema(
					map[string]*spec.Schema{
						spec.TypeBoolean: spec.NewBoolSchema(),

						/*use plural to check the removal of the "s"*/
						"objects": spec.NewArraySchema().WithItems(stringNumberObject),
					},
				),
			},
			wantRetSchemas: spec.Schemas{
				"object": spec.NewSchemaRef("", stringNumberObject),
				"boolean_objects": spec.NewSchemaRef("", createObjectSchemaWithRef(
					map[string]*spec.SchemaRef{
						spec.TypeBoolean: spec.NewSchemaRef("", spec.NewBoolSchema()),
						"objects":        spec.NewSchemaRef("", createArraySchemaWithRefItems(schemasRefPrefix+"object")),
					},
				)),
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"boolean_objects", nil),
		},
		{
			name: "object in object in object - max depth was reached after 1 object - ref was not created",
			args: args{
				schemas: nil,
				schema: createObjectSchema(
					map[string]*spec.Schema{
						"obj1": createObjectSchema(
							map[string]*spec.Schema{
								"obj2": stringNumberObject,
							},
						),
					},
				),
				depth: maxSchemaToRefDepth - 1,
			},
			wantRetSchemas: spec.Schemas{
				"obj1": spec.NewSchemaRef("", createObjectSchema(
					map[string]*spec.Schema{
						"obj2": stringNumberObject,
					},
				),
				),
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"obj1", nil),
		},
		{
			name: "object in object in object - max depth was reached after 2 objects - ref was not created",
			args: args{
				schemas: nil,
				schema: createObjectSchema(
					map[string]*spec.Schema{
						"obj1": createObjectSchema(
							map[string]*spec.Schema{
								"obj2":   stringNumberObject,
								"string": spec.NewStringSchema(),
							},
						),
					},
				),
				depth: maxSchemaToRefDepth - 2,
			},
			wantRetSchemas: spec.Schemas{
				"obj1": spec.NewSchemaRef("", createObjectSchema(
					map[string]*spec.Schema{
						"obj2":   stringNumberObject,
						"string": spec.NewStringSchema(),
					},
				),
				),
				"obj1_0": spec.NewSchemaRef("", createObjectSchemaWithRef(
					map[string]*spec.SchemaRef{
						"obj1": spec.NewSchemaRef(schemasRefPrefix+"obj1", nil),
					},
				),
				),
			},
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"obj1_0", nil),
		},
		{
			name: "max depth was reached - ref was not created",
			args: args{
				schemas: nil,
				schema: createObjectSchema(
					map[string]*spec.Schema{
						spec.TypeBoolean: spec.NewBoolSchema(),

						/*use plural to check the removal of the "s"*/
						"objects": spec.NewArraySchema().WithItems(stringNumberObject),
					},
				),
				depth: maxSchemaToRefDepth,
			},
			wantRetSchemas: nil,
			wantRetSchema: spec.NewSchemaRef(schemasRefPrefix+"obj1_0", createObjectSchema(
				map[string]*spec.Schema{
					spec.TypeBoolean: spec.NewBoolSchema(),

					/*use plural to check the removal of the "s"*/
					"objects": spec.NewArraySchema().WithItems(stringNumberObject),
				},
			)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetDefinitions, gotRetSchema := schemaToRef(tt.args.schemas, tt.args.schema, tt.args.defNameHint, tt.args.depth)
			if !reflect.DeepEqual(gotRetDefinitions, tt.wantRetSchemas) {
				t.Errorf("schemaToRef() gotRetDefinitions = %v, want %v", marshal(gotRetDefinitions), marshal(tt.wantRetSchemas))
			}
			if !reflect.DeepEqual(gotRetSchema, tt.wantRetSchema) {
				t.Errorf("schemaToRef() gotRetSchema = %v, want %v", marshal(gotRetSchema), marshal(tt.wantRetSchema))
			}
		})
	}
}

var interactionReqBody = `{"active":true,
"certificateVersion":"86eb5278-676a-3b7c-b29d-4a57007dc7be",
"controllerInstanceInfo":{"replicaId":"portshift-agent-66fc77c848-tmmk8"},
"policyAndAppVersion":1621477900361,
"version":"1.147.1"}`

var interactionRespBody = `{"cvss":[{"score":7.8,"vector":"AV:L/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H"}]}`

var interaction = &HTTPInteractionData{
	ReqBody:  interactionReqBody,
	RespBody: interactionRespBody,
	ReqHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	RespHeaders: map[string]string{
		contentTypeHeaderName: mediaTypeApplicationJSON,
	},
	statusCode: 200,
}

func createArraySchemaWithRef(ref string) *spec.Schema {
	arraySchema := spec.NewArraySchema()
	arraySchema.Items = &spec.SchemaRef{Ref: ref}
	return arraySchema
}

func Test_updateSchemas(t *testing.T) {
	op := NewOperation(t, interaction).Op
	retOp := NewOperation(t, interaction).Op
	retOp.Parameters[0] = &spec.ParameterRef{
		Ref: schemasRefPrefix + "active_certificateVersion_controllerInstanceInfo_policyAndAppVersion_version",
	}
	retOp.Responses["200"] = &spec.ResponseRef{
		Ref: schemasRefPrefix + "cvss",
	}

	type args struct {
		definitions spec.Schemas
		op          *spec.Operation
	}
	tests := []struct {
		name             string
		args             args
		wantRetSchemas   spec.Schemas
		wantRetOperation *spec.Operation
	}{
		{
			name: "sanity",
			args: args{
				definitions: nil,
				op:          op,
			},
			wantRetSchemas: spec.Schemas{
				"controllerInstanceInfo": spec.NewSchemaRef("", createObjectSchema(
					map[string]*spec.Schema{
						"replicaId": spec.NewStringSchema(),
					},
				)),
				"active_certificateVersion_controllerInstanceInfo_policyAndAppVersion_version": spec.NewSchemaRef("", createObjectSchemaWithRef(
					map[string]*spec.SchemaRef{
						"active":                 spec.NewSchemaRef("", spec.NewBoolSchema()),
						"certificateVersion":     spec.NewSchemaRef("", spec.NewUUIDSchema()),
						"controllerInstanceInfo": spec.NewSchemaRef(schemasRefPrefix+"controllerInstanceInfo", nil),
						"policyAndAppVersion":    spec.NewSchemaRef("", spec.NewInt64Schema()),
						"version":                spec.NewSchemaRef("", spec.NewStringSchema()),
					},
				)),
				"cvs": spec.NewSchemaRef("", createObjectSchema(
					map[string]*spec.Schema{
						"score":  spec.NewFloat64Schema(),
						"vector": spec.NewStringSchema(),
					},
				)),
				"cvss": spec.NewSchemaRef("", createObjectSchema(
					map[string]*spec.Schema{
						"cvss": createArraySchemaWithRef(schemasRefPrefix + "cvs"),
					},
				)),
			},
			wantRetOperation: retOp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetDefinitions, gotRetOperation := updateSchemas(tt.args.definitions, tt.args.op)
			if !reflect.DeepEqual(gotRetDefinitions, tt.wantRetSchemas) {
				t.Errorf("updateSchemas() gotRetDefinitions = %v, want %v", marshal(gotRetDefinitions), marshal(tt.wantRetSchemas))
			}
			if !reflect.DeepEqual(gotRetOperation, tt.wantRetOperation) {
				t.Errorf("updateSchemas() gotRetOperation = %v, want %v", marshal(gotRetOperation), marshal(tt.wantRetOperation))
			}
		})
	}
}
