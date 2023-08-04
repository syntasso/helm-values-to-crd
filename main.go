package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	pluralize "github.com/gertd/go-pluralize"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	yamlsig "sigs.k8s.io/yaml"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("expected 2 args: values file path and Kubernetes GVK.\nExample:: ./helm-values-to-crd test.yaml redis.platform.kratix.io/v1alpha1\n")
		os.Exit(1)
	}
	valuesFile := os.Args[1]
	gvk := os.Args[2]

	kind, version, group := splitGVK(gvk)

	// Read the values file
	var template map[string]interface{}
	templateBytes, err := ioutil.ReadFile(valuesFile)
	if err != nil {
		panic("failed to read file " + valuesFile + ", " + err.Error())
	}

	// Unmarshal the values file
	err = yaml.Unmarshal(templateBytes, &template)
	if err != nil {
		panic("failed tounmarshal " + valuesFile + ", " + err.Error())
	}

	// Convert the values file to a CRD
	crd, err := convertValuesToCRD(context.TODO(), template, group, version, kind)
	if err != nil {
		panic("failed to generate crd: " + err.Error())
	}

	// Marshal CRD to bytes
	bytes, err := yamlsig.Marshal(crd)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))
}

func splitGVK(gvk string) (string, string, string) {
	kind := strings.Split(gvk, ".")[0]
	version := strings.Split(gvk, "/")[len(strings.Split(gvk, "/"))-1]
	group := strings.TrimSuffix(strings.TrimPrefix(gvk, fmt.Sprintf("%s.", kind)), fmt.Sprintf("/%s", version))
	return kind, version, group
}

func convertValuesToCRD(ctx context.Context, template map[string]interface{}, group, version, kind string) (*apiextensionsv1.CustomResourceDefinition, error) {
	var validationSchema *apiextensionsv1.JSONSchemaProps = &apiextensionsv1.JSONSchemaProps{
		Type:       "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{},
	}

	//Auto inject the `spec: ` top level key
	var openAPIV3Schema *apiextensionsv1.JSONSchemaProps = &apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"spec": *validationSchema,
		},
	}

	//Add to the `spec` the keys from the values file
	for key, value := range map[string]interface{}(template) {
		validationSchema.Properties[key] = generateJSONSchemaFromValue(value)
	}

	//Generate CRD with the spec and GVK
	pluralKind := strings.ToLower(pluralize.NewClient().Plural(kind))
	xaasCRD := &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CustomResourceDefinition",
			APIVersion: apiextensionsv1.SchemeGroupVersion.Identifier(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: pluralKind + "." + group,
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: group,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Plural:   pluralKind,
				Singular: strings.ToLower(kind),
				Kind:     kind,
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    version,
					Served:  true,
					Storage: true,
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: openAPIV3Schema,
					},
				},
			},
		},
	}

	return xaasCRD, nil
}

func generateJSONSchemaFromValue(value interface{}) apiextensionsv1.JSONSchemaProps {
	boolTrue := true
	switch valueType := value.(type) {
	case string:
		return apiextensionsv1.JSONSchemaProps{
			Type: "string",
		}
	case int, int16, int32, int64, int8, float32, float64:
		return apiextensionsv1.JSONSchemaProps{
			Type: "integer",
		}
	case bool:
		return apiextensionsv1.JSONSchemaProps{
			Type: "boolean",
		}
	case []interface{}:
		v := value.([]interface{})
		var schemaV apiextensionsv1.JSONSchemaProps
		if len(v) > 0 {
			schemaV = generateJSONSchemaFromValue(v[0])
		} else {
			schemaV = apiextensionsv1.JSONSchemaProps{
				XIntOrString: true,
			}
		}
		return apiextensionsv1.JSONSchemaProps{
			Type: "array",
			Items: &apiextensionsv1.JSONSchemaPropsOrArray{
				Schema: &schemaV,
			},
		}
	case map[string]interface{}:
		jsonSchema := map[string]apiextensionsv1.JSONSchemaProps{}
		for key, value := range valueType {
			jsonSchema[key] = generateJSONSchemaFromValue(value)
		}
		return apiextensionsv1.JSONSchemaProps{
			Type:                   "object",
			Properties:             jsonSchema,
			XPreserveUnknownFields: &boolTrue,
		}
	case map[interface{}]interface{}:
		jsonSchema := map[string]apiextensionsv1.JSONSchemaProps{}
		for key, value := range valueType {
			keyString, ok := key.(string)
			if !ok {
				panic(fmt.Sprintf("key is not string: %v", key))
			}
			jsonSchema[keyString] = generateJSONSchemaFromValue(value)
		}
		return apiextensionsv1.JSONSchemaProps{
			Type:                   "object",
			Properties:             jsonSchema,
			XPreserveUnknownFields: &boolTrue,
		}
	default:
		panic("unknown type, code needs to be updated to handle this")
	}
}
