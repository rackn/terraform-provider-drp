package drp

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/digitalrebar/provision/models"
	"github.com/hashicorp/terraform/helper/schema"
)

func buildSchemaListFromObject(m interface{}) *schema.Schema {
	r := &schema.Resource{
		Schema: buildSchemaFromObject(m),
	}
	return &schema.Schema{
		Type:     schema.TypeList,
		Elem:     r,
		Optional: true,
		Computed: true,
	}
}

func buildSchemaFromObject(m interface{}) map[string]*schema.Schema {
	sm := map[string]*schema.Schema{}

	val := reflect.ValueOf(m).Elem()

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

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if typeField.Name == "Meta" {
			sm["Meta"] = &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			}

			continue
		}

		// This is a cluster.  Terraform doesn't
		if typeField.Name == "Params" {
			// GREG: FIGURE THIS OUT!!!
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]
			if listType[0] == '*' {
				listType = listType[1:]
			}

			switch listType {
			case "string":
				sm[typeField.Name] = &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
					Computed: true,
				}
			case "models.DhcpOption":
				sm[typeField.Name] = buildSchemaListFromObject(&models.DhcpOption{})
			case "models.TemplateInfo":
				sm[typeField.Name] = buildSchemaListFromObject(&models.TemplateInfo{})
			case "models.Param":
				sm[typeField.Name] = buildSchemaListFromObject(&models.Param{})
			case "models.AvailableAction":
				sm[typeField.Name] = buildSchemaListFromObject(&models.AvailableAction{})
			default:
				log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					typeField.Name, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			// Singleton struct - encode as a list for now.
			sm[typeField.Name] = buildSchemaListFromObject(&models.OsInfo{})
		case "string", "[]uint8", "net.IP", "uuid.UUID", "time.Time":
			sm[typeField.Name] = &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			}
		case "bool":
			sm[typeField.Name] = &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			}
		case "int", "int32":
			sm[typeField.Name] = &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			}
		default:
			log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				typeField.Name, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}

	return sm
}

func buildSchema(m models.Model) *schema.Resource {
	r := &schema.Resource{
		Create: createDefaultCreateFunction(m),
		Read:   createDefaultReadFunction(m),
		Update: createDefaultUpdateFunction(m),
		Delete: createDefaultDeleteFunction(m),
		Exists: createDefaultExistsFunction(m),
		Schema: buildSchemaFromObject(m),
	}

	return r
}

func updateResourceData(m models.Model, d *schema.ResourceData) error {
	val := reflect.ValueOf(m).Elem()

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

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if typeField.Name == "Meta" {
			d.Set("Meta", valueField.Interface())
			continue
		}

		// This is a cluster.  Terraform doesn't
		if typeField.Name == "Params" {
			// GREG: FIGURE THIS OUT!!!
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]
			if listType[0] == '*' {
				listType = listType[1:]
			}

			switch listType {
			case "string", "models.DhcpOption", "models.TemplateInfo", "models.Param", "models.AvailableAction":
				d.Set(typeField.Name, valueField.Interface())
			default:
				log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					typeField.Name, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
			d.Set(typeField.Name, []*models.OsInfo{valueField.Interface().(*models.OsInfo)})
		case "string", "[]uint8", "net.IP", "uuid.UUID", "time.Time":
			d.Set(typeField.Name, fmt.Sprintf("%s", valueField.Interface()))
		case "bool":
			d.Set(typeField.Name, valueField.Interface())
		case "int", "int32":
			d.Set(typeField.Name, valueField.Interface().(int))
		default:
			log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				typeField.Name, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	return nil
}

func buildModel(m models.Model, d *schema.ResourceData) (models.Model, error) {
	new := models.Clone(m)

	val := reflect.ValueOf(new).Elem()
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

		// Meta is a constant map of strings (but shows up as a type of Meta - fix it)
		if typeField.Name == "Meta" {
			if d.HasChange("Meta") {
				// GREG: This is not quite right
				// valueField.Set(d.Get("Meta"))
			}
			continue
		}

		// This is a cluster.  Terraform doesn't
		if typeField.Name == "Params" {
			// GREG: FIGURE THIS OUT!!!
			continue
		}

		if strings.HasPrefix(typeField.Type.String(), "[]") {
			listType := typeField.Type.String()[2:]
			if listType[0] == '*' {
				listType = listType[1:]
			}

			switch listType {
			case "string":
			case "models.DhcpOption":
			case "models.TemplateInfo":
			case "models.Param":
			case "models.AvailableAction":
			default:
				log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
					typeField.Name, typeField.Type,
					valueField.Interface(), tag.Get("tag_name"))
			}
			continue
		}

		switch typeField.Type.String() {
		case "models.OsInfo":
		case "string":
			if d.HasChange(typeField.Name) {
				valueField.SetString(d.Get(typeField.Name).(string))
			}
		case "[]uint8":
		case "net.IP":
		case "uuid.UUID":
		case "time.Time":
		case "bool":
			if d.HasChange(typeField.Name) {
				valueField.SetBool(d.Get(typeField.Name).(bool))
			}
		case "int", "int32":
			if d.HasChange(typeField.Name) {
				valueField.SetInt(d.Get(typeField.Name).(int64))
			}
		default:
			log.Printf("[DEBUG] UNKNOWN Field Name: %s (%s),\t Field Value: %v,\t Tag Value: %s\n",
				typeField.Name, typeField.Type,
				valueField.Interface(), tag.Get("tag_name"))
		}
	}
	return new, nil
}

func createDefaultCreateFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sCreate] creating\n", m.Prefix())

		new, err := buildModel(m, d)
		if err != nil {
			return err
		}

		err = cc.session.CreateModel(new)
		if err != nil {
			return err
		}

		d.SetId(new.Key())

		return updateResourceData(new, d)
	}
}

func createDefaultReadFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sRead] reading %s\n", m.Prefix(), d.Id())

		answer, err := cc.session.GetModel(m.Prefix(), d.Id())
		if err != nil {
			return err
		}

		return updateResourceData(answer, d)
	}
}

func createDefaultUpdateFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sUpdate] updating %s\n", m.Prefix(), d.Id())

		base, err := cc.session.GetModel(m.Prefix(), d.Id())
		if err != nil {
			return err
		}

		mods, err := buildModel(base, d)
		if err != nil {
			return err
		}

		answer, err := cc.session.PatchTo(base, mods)
		if err != nil {
			return err
		}
		return updateResourceData(answer, d)
	}
}

func createDefaultDeleteFunction(m models.Model) func(*schema.ResourceData, interface{}) error {
	return func(d *schema.ResourceData, meta interface{}) error {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sDelete] deleting %s\n", m.Prefix(), d.Id())
		_, err := cc.session.DeleteModel(m.Prefix(), d.Id())
		return err
	}
}

func createDefaultExistsFunction(m models.Model) func(*schema.ResourceData, interface{}) (bool, error) {
	return func(d *schema.ResourceData, meta interface{}) (bool, error) {
		cc := meta.(*Config)
		log.Printf("[DEBUG] [resource%sExists] testing %s\n", m.Prefix(), d.Id())
		return cc.session.ExistsModel(m.Prefix(), d.Id())
	}
}
