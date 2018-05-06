package main

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"strings"
	"text/template"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/schema"
)

var modelMap = map[string]string{
	"machines":     "Machine",
	"bootenvs":     "BootEnv",
	"profiles":     "Profile",
	"params":       "Param",
	"plugins":      "Plugin",
	"tasks":        "Task",
	"stages":       "Stage",
	"workflows":    "Workflow",
	"users":        "User",
	"templates":    "Template",
	"subnets":      "Subnet",
	"reservations": "Reservation",
}

func buildSchemaListFromObject(m interface{}, computed bool) *schema.Schema {
	r := &schema.Resource{
		Schema: buildSchemaFromObject(m, computed),
	}
	return &schema.Schema{
		Type:     schema.TypeList,
		Elem:     r,
		Optional: true,
		Computed: computed,
	}
}

func buildSchemaFromObject(m interface{}, computed bool) map[string]*schema.Schema {
	sm := map[string]*schema.Schema{}

	val := reflect.ValueOf(m).Elem()

	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// Skip non-exported fields
		if typeField.PkgPath != "" {
			continue
		}

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			fieldName = "PluginProvider"
		}

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			sm["Meta"] = &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: computed,
			}

			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			sm["Params"] = &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: computed,
			}
			continue
		}
		if fieldName == "Schema" {
			sm["Schema"] = &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: computed,
			}
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]

			switch listType {
			case "string":
				sm[fieldName] = &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
					Computed: computed,
				}
			case "models.DhcpOption", "*models.DhcpOption":
				sm[fieldName] = buildSchemaListFromObject(&models.DhcpOption{}, computed)

			case "models.TemplateInfo":
				sm[fieldName] = buildSchemaListFromObject(&models.TemplateInfo{}, computed)
			case "uint8":
				sm[fieldName] = &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Computed: computed,
				}
			default:
				fmt.Printf("[DEBUG] UNKNOWN List Field Name: %s (%s),\t Tag Value: %s\n",
					fieldName, typeField.Type,
					tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			// Singleton struct - encode as a list for now.
			sm[fieldName] = buildSchemaListFromObject(&models.OsInfo{}, computed)
		case "string", "net.IP", "uuid.UUID", "time.Time":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: computed,
			}
		case "bool":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: computed,
			}
		case "int", "int32", "uint8":
			sm[fieldName] = &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: computed,
			}
		default:
			fmt.Printf("[DEBUG] UNKNOWN Base Field Name: %s (%s),\t Tag Value: %s\n",
				fieldName, typeField.Type,
				tag.Get("tag_name"))
		}
	}

	return sm
}

func updateResourceData(m models.Model) (string, []string) {
	str := fmt.Sprintf(`
func update%sResourceData(m models.Model, d *schema.ResourceData) error {
	`, m.Prefix())
	str += fmt.Sprintf("obj := m.(*models.%s)\n", modelMap[m.Prefix()])

	val := reflect.ValueOf(m).Elem()
	newImports := []string{}

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			str += "d.Set(\"PluginProvider\", obj.Provider)\n"
			continue
		}

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			str += fmt.Sprintf("d.Set(\"Meta\", obj.%s)\n", fieldName)
			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			str += `
			answer := map[string]string{}
			for k, v := range obj.Params {
				b, e := json.Marshal(v)
				if e != nil {
					return e
				}
				if s, ok := v.(string); ok {
					answer[k] = s
				} else {
					answer[k] = string(b)
				}
			}
			d.Set("Params", answer)
			`
			newImports = append(newImports, "encoding/json")
			continue
		}
		if fieldName == "Schema" {
			str += `
			b, e := json.Marshal(obj.Schema)
			if e != nil {
				return e
			}
			d.Set("Schema", string(b))
			`
			newImports = append(newImports, "encoding/json")
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]

			switch listType {
			case "string", "*models.DhcpOption", "models.DhcpOption", "models.TemplateInfo":
				str += fmt.Sprintf("d.Set(\"%s\", obj.%s)\n", fieldName, fieldName)
			case "uint8":
				str += fmt.Sprintf("d.Set(\"%s\", fmt.Sprintf(\"%%s\", obj.%s))\n", fieldName, fieldName)
			default:
				log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					fieldName, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			str += fmt.Sprintf("d.Set(\"%s\", []models.OsInfo{obj.%s})\n", fieldName, fieldName)
		case "string", "net.IP", "uuid.UUID", "time.Time":
			str += fmt.Sprintf("d.Set(\"%s\", fmt.Sprintf(\"%%s\", obj.%s))\n", fieldName, fieldName)
			switch typeField.Type.String() {
			case "string":
			case "net.IP":
				newImports = append(newImports, "net")
			case "uuid.UUID":
				newImports = append(newImports, "github.com/pborman/uuid")
			case "time.Time":
				newImports = append(newImports, "time")
			}
		case "bool", "int", "int32", "uint8":
			str += fmt.Sprintf("d.Set(\"%s\", obj.%s)\n", fieldName, fieldName)
		default:
			log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				fieldName, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	str += "return nil\n}\n"
	return str, newImports
}

func buildModel(m models.Model) (string, []string) {
	str := fmt.Sprintf(`
func build%sModel(m models.Model, d *schema.ResourceData) (models.Model, error) {
	new := models.Clone(m)
	`, m.Prefix())
	str += fmt.Sprintf("obj := new.(*models.%s)\n", modelMap[m.Prefix()])

	addClose := false
	newImports := []string{}

	val := reflect.ValueOf(m).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag
		if addClose {
			str += "}\n"
			addClose = false
		}

		// Skip the access and validation fields
		if typeField.Name == "Access" || typeField.Name == "Validation" {
			continue
		}

		// Skip the Profile - deprecated fields
		if typeField.Name == "Profile" {
			continue
		}

		fieldName := typeField.Name
		// Provider is reserved Terraform name
		if fieldName == "Provider" {
			str += "obj.Provider = d.Get(\"PluginProvider\").(string)\n"
			continue
		}

		str += fmt.Sprintf("if d.HasChange(\"%s\") {\n", fieldName)
		addClose = true

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if fieldName == "Meta" {
			str += `
			obj.Meta = models.Meta{}
			ms := d.Get("Meta").(map[string]interface{})
			for k, v := range ms {
				obj.Meta[k] = v.(string)
			}
			`
			continue
		}

		//
		// This is a cluster.  Terraform doesn't generic interface{}
		// basically, interface{} and map[string]interface{}
		//
		// Will try some things.
		//
		if fieldName == "Params" {
			str += `
			answer := d.Get("Params").(map[string]interface{})
			obj.Params = map[string]interface{}{}

			for k, v := range answer {
				s := v.(string)

				var i interface{}
				if e := json.Unmarshal([]byte(s), &i); e != nil {
					i = s
				}

				obj.Params[k] = i
			}
			`
			newImports = append(newImports, "encoding/json")
			continue
		}
		if fieldName == "Schema" {
			str += `
			s := d.Get("Schema").(string)
			var i interface{}
			if e := json.Unmarshal([]byte(s), &i); e != nil {
				return nil, e
			}
			obj.Schema = i
			`
			newImports = append(newImports, "encoding/json")
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]
			subType := typeField.Type.Elem()

			switch listType {
			case "string", "models.TemplateInfo",
				"models.DhcpOption", "*models.DhcpOption":

				str += fmt.Sprintf(`
				data := d.Get("%s").([]interface{})
				v := make([]%s, 0, len(data))
				for _, s := range data {
					var no %s
					if err := utils.Remarshal(s, &no); err != nil {
						return nil, err
					}
					v = append(v, no)
				}
				obj.%s = v
				`, fieldName, listType, subType, fieldName)
				newImports = append(newImports, "github.com/VictorLowther/jsonpatch2/utils")

			case "uint8":
				fmt.Printf("[DEBUG] list of %s not support for push to DRP\n", listType)
			default:
				fmt.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					fieldName, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			str += fmt.Sprintf(`
			data := d.Get("%s").([]interface{})
			for _, s := range data {
				no := models.OsInfo{}
				if err := utils.Remarshal(s, &no); err != nil {
					return nil, err
				}
				obj.%s = no
				break
			}
			`, fieldName, fieldName)
			newImports = append(newImports, "github.com/VictorLowther/jsonpatch2/utils")
		case "string":
			str += fmt.Sprintf("obj.%s = d.Get(\"%s\").(string)\n", fieldName, fieldName)
		case "net.IP":
			str += fmt.Sprintf(`
			ip := net.ParseIP(d.Get("%s").(string))
			obj.%s = ip
			`, fieldName, fieldName)
			newImports = append(newImports, "net")
		case "uuid.UUID":
			str += fmt.Sprintf(`
			uu := uuid.Parse(d.Get("%s").(string))
			obj.%s = uu
			`, fieldName, fieldName)
			newImports = append(newImports, "github.com/pborman/uuid")
		case "time.Time":
			str += fmt.Sprintf(`
			if t, e := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST",
				d.Get("%s").(string)); e != nil {
				return nil, e
			} else {
				obj.%s = t
			}
			`, fieldName, fieldName)
			newImports = append(newImports, "time")
		case "bool":
			str += fmt.Sprintf("obj.%s = d.Get(\"%s\").(bool)\n", fieldName, fieldName)
		case "int":
			str += fmt.Sprintf("obj.%s = d.Get(\"%s\").(int)\n", fieldName, fieldName)
		case "int32":
			str += fmt.Sprintf("obj.%s = int32(d.Get(\"%s\").(int))\n", fieldName, fieldName)
		case "uint8":
			str += fmt.Sprintf("obj.%s = uint8(d.Get(\"%s\").(int))\n", fieldName, fieldName)
		default:
			fmt.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				fieldName, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	if addClose {
		str += "}\n"
		addClose = false
	}
	str += "return new, nil\n}\n"
	return str, newImports
}

type SchemaHolder struct {
	Schema   *schema.Schema
	IsNested bool
}

func (s *SchemaHolder) ElemToString() string {
	var str string
	if rs, ok := s.Schema.Elem.(*schema.Resource); ok {
		str = "&schema.Resource{\nSchema: map[string]*schema.Schema{\n"
		for k, sch := range rs.Schema {
			schStr, _ := schemaCode(sch, "", true)
			str += fmt.Sprintf("%q: %s,\n", k, schStr)
		}
		str += "},\n}"
	} else if sch, ok := s.Schema.Elem.(*schema.Schema); ok {
		str, _ = schemaCode(sch, "", true)
	} else {
		str = fmt.Sprintf("\nGREG: %T\n", s.Schema.Elem)
	}
	return str
}

func schemaCode(s *schema.Schema, setFunc string, isNested bool) (string, error) {
	buf := bytes.NewBuffer([]byte{})

	sh := &SchemaHolder{
		Schema:   s,
		IsNested: isNested,
	}

	if err := schemaTemplate.Execute(buf, sh); err != nil {
		return "", err
	}

	return buf.String(), nil
}

var schemaTemplate = template.Must(template.New("schema").Parse(`{{if .IsNested}}&schema.Schema{{end}}{{"{"}}{{if not .IsNested}}
{{end}}Type: schema.{{.Schema.Type}},{{if ne .Schema.Description ""}}
Description: {{printf "%q" .Schema.Description}},{{end}}{{if .Schema.Required}}
Required: {{.Schema.Required}},{{end}}{{if .Schema.Optional}}
Optional: {{.Schema.Optional}},{{end}}{{if .Schema.ForceNew}}
ForceNew: {{.Schema.ForceNew}},{{end}}{{if .Schema.Computed}}
Computed: {{.Schema.Computed}},{{end}}{{if gt .Schema.MaxItems 0}}
MaxItems: {{.Schema.MaxItems}},{{end}}{{if .Schema.Elem}}
Elem: {{.ElemToString}},{{end}}{{if not .IsNested}}
{{end}}{{"}"}}`))
