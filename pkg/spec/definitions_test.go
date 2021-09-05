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

	"github.com/go-openapi/spec"
)

var (
	stringNumberObject = createObjectSchema(
		[]string{
			schemaTypeString,
			schemaTypeNumber,
		},
		[]spec.Schema{
			*spec.StringProperty(),
			*spec.Float32Property(),
		},
	)
	stringBooleanObject = createObjectSchema(
		[]string{
			schemaTypeString,
			schemaTypeBoolean,
		},
		[]spec.Schema{
			*spec.StringProperty(),
			*spec.BooleanProperty(),
		},
	)
	stringIntegerObject = createObjectSchema(
		[]string{
			schemaTypeString,
			schemaTypeInteger,
		},
		[]spec.Schema{
			*spec.StringProperty(),
			*spec.Int64Property(),
		},
	)
)

func marshal(obj interface{}) string {
	objB, _ := json.Marshal(obj)
	return string(objB)
}

func createObjectSchema(names []string, objSchemas []spec.Schema) *spec.Schema {
	schema := &spec.Schema{}
	schema.AddType(schemaTypeObject, "")
	for i := range names {
		schema.SetProperty(names[i], objSchemas[i])
	}
	return schema
}

func Test_findDefinition(t *testing.T) {
	type args struct {
		definitions map[string]spec.Schema
		schema      *spec.Schema
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
				definitions: map[string]spec.Schema{
					"string": *spec.StringProperty(),
				},
				schema: spec.StringProperty(),
			},
			wantDefName: "string",
			wantExist:   true,
		},
		{
			name: "identical string schema does not exist",
			args: args{
				definitions: map[string]spec.Schema{
					"string": *spec.StrFmtProperty("format"),
				},
				schema: spec.StringProperty(),
			},
			wantDefName: "",
			wantExist:   false,
		},
		{
			name: "identical object schema exist (object order is different)",
			args: args{
				definitions: map[string]spec.Schema{
					"object": *createObjectSchema(
						[]string{
							schemaTypeObject,
							schemaTypeString,
						},
						[]spec.Schema{
							*stringIntegerObject,
							*spec.StringProperty(),
						},
					),
				},
				schema: createObjectSchema(
					[]string{
						schemaTypeString,
						schemaTypeObject,
					},
					[]spec.Schema{
						*spec.StringProperty(),
						*createObjectSchema(
							[]string{
								schemaTypeInteger,
								schemaTypeString,
							},
							[]spec.Schema{
								*spec.Int64Property(),
								*spec.StringProperty(),
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
				definitions: map[string]spec.Schema{
					"object": *createObjectSchema(
						[]string{
							schemaTypeString,
							schemaTypeObject,
						},
						[]spec.Schema{
							*spec.StringProperty(),
							*stringIntegerObject,
						},
					),
				},
				schema: createObjectSchema(
					[]string{
						schemaTypeString,
						schemaTypeObject,
					},
					[]spec.Schema{
						*spec.StringProperty(),
						*stringNumberObject,
					},
				),
			},
			wantDefName: "",
			wantExist:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDefName, gotExist := findDefinition(tt.args.definitions, tt.args.schema)
			if gotDefName != tt.wantDefName {
				t.Errorf("findDefinition() gotDefName = %v, want %v", gotDefName, tt.wantDefName)
			}
			if gotExist != tt.wantExist {
				t.Errorf("findDefinition() gotExist = %v, want %v", gotExist, tt.wantExist)
			}
		})
	}
}

func Test_getUniqueDefName(t *testing.T) {
	type args struct {
		definitions map[string]spec.Schema
		name        string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "name does not exist",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *stringIntegerObject,
				},
				name: "no-test",
			},
			want: "no-test_0",
		},
		{
			name: "name exist once",
			args: args{
				definitions: map[string]spec.Schema{
					"test_0": *stringIntegerObject,
				},
				name: "test",
			},
			want: "test_1",
		},
		{
			name: "name exist multiple times",
			args: args{
				definitions: map[string]spec.Schema{
					"test":   *stringIntegerObject,
					"test_0": *stringNumberObject,
					"test_1": *stringBooleanObject,
				},
				name: "test",
			},
			want: "test_2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getUniqueDefName(tt.args.definitions, tt.args.name); got != tt.want {
				t.Errorf("getUniqueDefName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_schemaToRef(t *testing.T) {
	type args struct {
		definitions map[string]spec.Schema
		schema      *spec.Schema
		defNameHint string
		depth       int
	}
	tests := []struct {
		name               string
		args               args
		wantRetDefinitions map[string]spec.Schema
		wantRetSchema      *spec.Schema
	}{
		{
			name: "nil schema",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema:      nil,
				defNameHint: "",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
			},
			wantRetSchema: nil,
		},
		{
			name: "array schema with nil items",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema:      spec.ArrayProperty(nil),
				defNameHint: "",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
			},
			wantRetSchema: spec.ArrayProperty(nil),
		},
		{
			name: "array schema with non object items - no change for definitions",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema:      spec.ArrayProperty(spec.BooleanProperty()),
				defNameHint: "",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
			},
			wantRetSchema: spec.ArrayProperty(spec.BooleanProperty()),
		},
		{
			name: "array schema with non object items - no change for definitions",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema:      spec.ArrayProperty(spec.BooleanProperty()),
				defNameHint: "",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
			},
			wantRetSchema: spec.ArrayProperty(spec.BooleanProperty()),
		},
		{
			name: "array schema with object items - use hint name",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema:      spec.ArrayProperty(stringNumberObject),
				defNameHint: "hint",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
				"hint": *stringNumberObject,
			},
			wantRetSchema: spec.ArrayProperty(spec.RefSchema(definitionsRefPrefix + "hint")),
		},
		{
			name: "array schema with object items - hint name already exist",
			args: args{
				definitions: map[string]spec.Schema{
					"hint": *spec.BooleanProperty(),
				},
				schema:      spec.ArrayProperty(stringNumberObject),
				defNameHint: "hint",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"hint":   *spec.BooleanProperty(),
				"hint_0": *stringNumberObject,
			},
			wantRetSchema: spec.ArrayProperty(spec.RefSchema(definitionsRefPrefix + "hint_0")),
		},
		{
			name: "primitive type",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *spec.BooleanProperty(),
				},
				schema: spec.Int64Property(),
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *spec.BooleanProperty(),
			},
			wantRetSchema: spec.Int64Property(),
		},
		{
			name: "empty object - no new definition",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *stringNumberObject,
				},
				schema: createObjectSchema(nil, nil),
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *stringNumberObject,
			},
			wantRetSchema: createObjectSchema(nil, nil),
		},
		{
			name: "object - definition exist",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *stringNumberObject,
				},
				schema: stringNumberObject,
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *stringNumberObject,
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "test"),
		},
		{
			name: "object - definition does not exist",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *stringBooleanObject,
				},
				schema: stringNumberObject,
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test":          *stringBooleanObject,
				"number_string": *stringNumberObject,
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "number_string"),
		},
		{
			name: "object - definition does not exist - use hint",
			args: args{
				definitions: map[string]spec.Schema{
					"test": *stringBooleanObject,
				},
				schema:      stringNumberObject,
				defNameHint: "hint",
			},
			wantRetDefinitions: map[string]spec.Schema{
				"test": *stringBooleanObject,
				"hint": *stringNumberObject,
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "hint"),
		},
		{
			name: "object in object",
			args: args{
				definitions: nil,
				schema: createObjectSchema(
					[]string{
						schemaTypeString,
						schemaTypeObject,
					},
					[]spec.Schema{
						*spec.StringProperty(),
						*stringNumberObject,
					},
				),
			},
			wantRetDefinitions: map[string]spec.Schema{
				"object": *stringNumberObject,
				"object_string": *createObjectSchema(
					[]string{
						schemaTypeObject,
						schemaTypeString,
					},
					[]spec.Schema{
						*spec.RefSchema(definitionsRefPrefix + "object"),
						*spec.StringProperty(),
					},
				),
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "object_string"),
		},
		{
			name: "array of object in an object",
			args: args{
				definitions: nil,
				schema: createObjectSchema(
					[]string{
						schemaTypeBoolean,
						"objects", // use plural to check the removal of the "s"
					},
					[]spec.Schema{
						*spec.BooleanProperty(),
						*spec.ArrayProperty(stringNumberObject),
					},
				),
			},
			wantRetDefinitions: map[string]spec.Schema{
				"object": *stringNumberObject,
				"boolean_objects": *createObjectSchema(
					[]string{
						schemaTypeBoolean,
						"objects",
					},
					[]spec.Schema{
						*spec.BooleanProperty(),
						*spec.ArrayProperty(spec.RefSchema(definitionsRefPrefix + "object")),
					},
				),
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "boolean_objects"),
		},
		{
			name: "object in object in object - max depth was reached after 1 object - ref was not created",
			args: args{
				definitions: nil,
				schema: createObjectSchema(
					[]string{
						"obj1",
					},
					[]spec.Schema{
						*createObjectSchema(
							[]string{
								"obj2",
							},
							[]spec.Schema{
								*stringNumberObject,
							},
						),
					},
				),
				depth: maxSchemaToRefDepth - 1,
			},
			wantRetDefinitions: map[string]spec.Schema{
				"obj1": *createObjectSchema(
					[]string{
						"obj1",
					},
					[]spec.Schema{
						*createObjectSchema(
							[]string{
								"obj2",
							},
							[]spec.Schema{
								*stringNumberObject,
							},
						),
					},
				),
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "obj1"),
		},
		{
			name: "object in object in object - max depth was reached after 2 objects - ref was not created",
			args: args{
				definitions: nil,
				schema: createObjectSchema(
					[]string{
						"obj1",
					},
					[]spec.Schema{
						*createObjectSchema(
							[]string{
								"obj2",
								"string",
							},
							[]spec.Schema{
								*stringNumberObject,
								*spec.StringProperty(),
							},
						),
					},
				),
				depth: maxSchemaToRefDepth - 2,
			},
			wantRetDefinitions: map[string]spec.Schema{
				"obj1": *createObjectSchema(
					[]string{
						"obj2",
						"string",
					},
					[]spec.Schema{
						*stringNumberObject,
						*spec.StringProperty(),
					},
				),
				"obj1_0": *createObjectSchema(
					[]string{
						"obj1",
					},
					[]spec.Schema{
						*spec.RefSchema(definitionsRefPrefix + "obj1"),
					},
				),
			},
			wantRetSchema: spec.RefSchema(definitionsRefPrefix + "obj1_0"),
		},
		{
			name: "max depth was reached - ref was not created",
			args: args{
				definitions: nil,
				schema: createObjectSchema(
					[]string{
						schemaTypeBoolean,
						"objects", // use plural to check the removal of the "s"
					},
					[]spec.Schema{
						*spec.BooleanProperty(),
						*spec.ArrayProperty(stringNumberObject),
					},
				),
				depth: maxSchemaToRefDepth,
			},
			wantRetDefinitions: nil,
			wantRetSchema: createObjectSchema(
				[]string{
					schemaTypeBoolean,
					"objects", // use plural to check the removal of the "s"
				},
				[]spec.Schema{
					*spec.BooleanProperty(),
					*spec.ArrayProperty(stringNumberObject),
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetDefinitions, gotRetSchema := schemaToRef(tt.args.definitions, tt.args.schema, tt.args.defNameHint, tt.args.depth)
			if !reflect.DeepEqual(gotRetDefinitions, tt.wantRetDefinitions) {
				t.Errorf("schemaToRef() gotRetDefinitions = %v, want %v", marshal(gotRetDefinitions), marshal(tt.wantRetDefinitions))
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

func Test_updateDefinitions(t *testing.T) {
	op := NewOperation(t, interaction).Op
	retOp := NewOperation(t, interaction).Op
	retOp.Parameters[0].Schema = spec.RefSchema(definitionsRefPrefix + "active_certificateVersion_controllerInstanceInfo_policyAndAppVersion_version")
	response := retOp.Responses.StatusCodeResponses[200]
	response.Schema = spec.RefSchema(definitionsRefPrefix + "cvss")
	retOp.Responses.StatusCodeResponses[200] = response

	type args struct {
		definitions map[string]spec.Schema
		op          *spec.Operation
	}
	tests := []struct {
		name               string
		args               args
		wantRetDefinitions map[string]spec.Schema
		wantRetOperation   *spec.Operation
	}{
		{
			name: "sanity",
			args: args{
				definitions: nil,
				op:          op,
			},
			wantRetDefinitions: map[string]spec.Schema{
				"controllerInstanceInfo": *createObjectSchema(
					[]string{
						"replicaId",
					},
					[]spec.Schema{
						*spec.StringProperty(),
					},
				),
				"active_certificateVersion_controllerInstanceInfo_policyAndAppVersion_version": *createObjectSchema(
					[]string{
						"active",
						"certificateVersion",
						"controllerInstanceInfo",
						"policyAndAppVersion",
						"version",
					},
					[]spec.Schema{
						*spec.BooleanProperty(),
						*spec.StrFmtProperty(formatUUID),
						*spec.RefSchema(definitionsRefPrefix + "controllerInstanceInfo"),
						*spec.Int64Property(),
						*spec.StringProperty(),
					},
				),
				"cvs": *createObjectSchema(
					[]string{
						"score",
						"vector",
					},
					[]spec.Schema{
						*spec.Float64Property(),
						*spec.StringProperty(),
					},
				),
				"cvss": *createObjectSchema(
					[]string{
						"cvss",
					},
					[]spec.Schema{
						*spec.ArrayProperty(spec.RefSchema(definitionsRefPrefix + "cvs")),
					},
				),
			},
			wantRetOperation: retOp,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRetDefinitions, gotRetOperation := updateDefinitions(tt.args.definitions, tt.args.op)
			if !reflect.DeepEqual(gotRetDefinitions, tt.wantRetDefinitions) {
				t.Errorf("updateDefinitions() gotRetDefinitions = %v, want %v", marshal(gotRetDefinitions), marshal(tt.wantRetDefinitions))
			}
			if !reflect.DeepEqual(gotRetOperation, tt.wantRetOperation) {
				t.Errorf("updateDefinitions() gotRetOperation = %v, want %v", marshal(gotRetOperation), marshal(tt.wantRetOperation))
			}
		})
	}
}
